package chainManager

import (
	"ShardDB/proto"
	"ShardDB/util"
	"encoding/json"
	"github.com/boltdb/bolt"
	"strconv"
)

func TPCCPaymentChainImpl(db *bolt.DB) *Chain {
	// Create the list of hops
	hops := make([]*Hop, 3)

	hops[0] = &Hop{
		id:         0,
		hopType:    util.NormalHop,
		isReadOnly: false,
		process:    TPCCPaymentWareHouseHop,
		conflicts:  map[int32]int32{},
	}

	hops[1] = &Hop{
		id:         11,
		hopType:    util.NormalHop,
		isReadOnly: false,
		process:    TPCCPaymentDistrictHop,
		conflicts:  map[int32]int32{},
	}

	hops[2] = &Hop{
		id:         11,
		hopType:    util.NormalHop,
		isReadOnly: false,
		process:    TPCCPaymentCustomerHop,
		conflicts:  map[int32]int32{},
	}

	return &Chain{db: db, hops: hops, NextReq: TPCCPaymentNextReq, Sort: defaultSort, GetShards: TPCCPaymentGetShard}
}

func TPCCPaymentWareHouseHop(tx *bolt.Tx, in *proto.TrxReq) (*proto.TrxRes, error) {
	var rwSet []*proto.RWSet
	key, warehouse := getWareHouse(tx, in.Params)

	rwSet = AddWriteSet(rwSet, "Warehouse", key)

	amount, _ := strconv.ParseUint(in.Params["amount"], 10, 64)
	warehouse.W_YTD += float32(amount)

	wBytes, _ := json.Marshal(warehouse)
	putData(tx.Bucket([]byte("tWarehouses")), key, wBytes)

	return &proto.TrxRes{Status: proto.Status_Success, Info: in.Info, RWSets: rwSet}, nil
}

func TPCCPaymentDistrictHop(tx *bolt.Tx, in *proto.TrxReq) (*proto.TrxRes, error) {
	var rwSet []*proto.RWSet
	wid, _ := strconv.ParseUint(in.Params["wid"], 10, 64)
	did, _ := strconv.ParseUint(in.Params["did"], 10, 64)
	key, district := getDistrict(tx, wid, did)

	rwSet = AddWriteSet(rwSet, "District", key)

	amount, _ := strconv.ParseUint(in.Params["amount"], 10, 64)
	district.D_TAX += float32(amount)

	dBytes, _ := json.Marshal(district)
	putData(tx.Bucket([]byte("tDistrict")), key, dBytes)

	return &proto.TrxRes{Status: proto.Status_Success, Info: in.Info, RWSets: rwSet}, nil
}

func TPCCPaymentCustomerHop(tx *bolt.Tx, in *proto.TrxReq) (*proto.TrxRes, error) {
	var rwSet []*proto.RWSet

	wid, _ := strconv.ParseUint(in.Params["cwid"], 10, 64)
	did, _ := strconv.ParseUint(in.Params["cdid"], 10, 64)
	cid, _ := strconv.ParseUint(in.Params["cid"], 10, 64)

	key, customer := getCustomer(tx, wid, did, cid)
	rwSet = AddWriteSet(rwSet, "Customer", key)

	amount, _ := strconv.ParseUint(in.Params["amount"], 10, 64)
	customer.C_BALANCE -= float32(amount)
	if customer.C_CREDIT == "BC" {
		customer.C_DATA = in.Params["amount"] + "|" + customer.C_DATA
	}

	cbytes, _ := json.Marshal(customer)
	putData(tx.Bucket([]byte("tCustomer")), key, cbytes)

	return &proto.TrxRes{Status: proto.Status_Success, Info: in.Info, RWSets: rwSet}, nil
}

func TPCCPaymentNextReq(in *proto.TrxRes, params map[string]string,
	shards int) (*proto.TrxReq, map[string]string, int) {
	// Fixed steps to set up the transaction request for the next hop
	hopId := in.Info.Hopid
	req := &proto.TrxReq{
		Info: in.Info,
	}
	nextShard := 0

	// Get the hopId for next hop
	// Set up the transaction parameters accordingly from the result of the last hop
	switch hopId {
	case 0:
		// First hop, find hopId of the warehouse
		nextShard, _ = strconv.Atoi(params["wid"])
		nextShard = nextShard * 10
	case 1:
		w, _ := strconv.Atoi(params["wid"])
		d, _ := strconv.Atoi(params["did"])
		nextShard = w*10 + d
	case 2:
		w, _ := strconv.Atoi(params["cwid"])
		d, _ := strconv.Atoi(params["cdid"])
		nextShard = w*10 + d
	default:
		return nil, params, 0
	}

	// Fixed steps
	req.Params = params
	req.Info.Hopid += 1
	req.Dependency = in.Dependency
	return req, params, nextShard
}

func TPCCPaymentGetShard(params map[string]string) []int32 {
	shards := make([]int32, 3)

	wid, _ := strconv.Atoi(params["wid"])
	shards[0] = int32(wid * 10)
	did, _ := strconv.Atoi(params["did"])
	shards[1] = int32(wid*10 + did%10)
	cwid, _ := strconv.Atoi(params["cwid"])
	cdid, _ := strconv.Atoi(params["cdid"])
	shards[2] = int32(cwid*10 + cdid%10)
	return shards
}
