package chainManager

import (
	"ShardDB/proto"
	"ShardDB/util"
	"encoding/json"
	"github.com/boltdb/bolt"
	"strconv"
)

func TPCCDeliveryChainImpl(db *bolt.DB) *Chain {
	// Create the list of hops
	hops := make([]*Hop, 10)

	for i := 0; i < 10; i++ {
		hops[i] = &Hop{
			id:         int32(i),
			hopType:    util.NormalHop,
			isReadOnly: false,
			process:    TPCCDeliveryDistrictHop,
			conflicts:  map[int32]int32{0: 11, 2: int32(i)},
		}
	}

	return &Chain{db: db, hops: hops, NextReq: TPCCDeliveryNextReq, Sort: defaultSort, GetShards: TPCCDeliveryGetShard}
}

func TPCCDeliveryDistrictHop(tx *bolt.Tx, in *proto.TrxReq) (*proto.TrxRes, error) {
	var rwSet []*proto.RWSet
	wid, _ := strconv.ParseUint(in.Params["wid"], 10, 64)
	carrier, _ := strconv.ParseUint(in.Params["carrier"], 10, 64)
	did := uint64(in.Info.Hopid - 1)
	key, district := getDistrict(tx, wid, did)
	rwSet = AddWriteSet(rwSet, "District", key)

	district.D_NEXT_NO_ID += 1
	next_no_id := district.D_NEXT_NO_ID

	dBytes, _ := json.Marshal(district)
	putData(tx.Bucket([]byte("tDistrict")), key, dBytes)

	oKey, order := getOrder(tx, wid, did, next_no_id)
	order.O_CARRIER_ID = carrier
	obytes, _ := json.Marshal(order)
	putData(tx.Bucket([]byte("tOrder")), oKey, obytes)

	// Simplified logic of customer operation
	cid := order.O_C_ID
	cKey, customer := getCustomer(tx, wid, did, cid)
	rwSet = AddWriteSet(rwSet, "Customer", cKey)
	customer.C_BALANCE += float32(order.O_OL_CNT)
	cbytes, _ := json.Marshal(customer)
	putData(tx.Bucket([]byte("tCustomer")), cKey, cbytes)

	return &proto.TrxRes{Status: proto.Status_Success, Info: in.Info, RWSets: rwSet}, nil
}

func TPCCDeliveryNextReq(in *proto.TrxRes, params map[string]string,
	shards int) (*proto.TrxReq, map[string]string, int) {
	// Fixed steps to set up the transaction request for the next hop
	hopId := in.Info.Hopid
	req := &proto.TrxReq{
		Info: in.Info,
	}
	nextShard := 0

	if hopId >= 0 && hopId < 10 {
		nextShard, _ = strconv.Atoi(params["wid"])
		nextShard = nextShard*10 + int(hopId)
	} else {
		return nil, params, 0
	}

	// Fixed steps
	req.Params = params
	req.Info.Hopid += 1
	req.Dependency = in.Dependency
	return req, params, nextShard
}

func TPCCDeliveryGetShard(params map[string]string) []int32 {
	shards := make([]int32, 10)

	wid, _ := strconv.Atoi(params["wid"])

	for i := 0; i < 10; i++ {
		shards[i] = int32(wid*10 + i)
	}

	return shards
}
