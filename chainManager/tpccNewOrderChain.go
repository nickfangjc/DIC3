package chainManager

import (
	"ShardDB/proto"
	"ShardDB/util"
	"encoding/json"
	"github.com/boltdb/bolt"
	"sort"
	"strconv"
)

const numOrderLine = 10

func TPCCNewOrderChainImpl(db *bolt.DB) *Chain {
	// Create the list of hops
	hops := make([]*Hop, 12)

	hops[0] = &Hop{
		id:         0,
		hopType:    util.NormalHop,
		isReadOnly: true,
		process:    TPCCNewOrderWareHouseHop,
		conflicts:  map[int32]int32{},
	}

	for i := 1; i < 11; i++ {
		hops[i] = &Hop{
			id:         int32(i),
			hopType:    util.MergedHop,
			isReadOnly: false,
			process:    TPCCNewOrderStockHop,
			conflicts:  map[int32]int32{0: int32(i)},
		}
	}
	hops[1].hopType = util.MergedHopBegin
	hops[10].hopType = util.MergedHopEnd

	hops[11] = &Hop{
		id:         11,
		hopType:    util.NormalHop,
		isReadOnly: false,
		process:    TPCCNewOrderDistrictHop,
		conflicts:  map[int32]int32{0: 11, 2: 0},
	}

	return &Chain{db: db, hops: hops, NextReq: TPCCNewOrderNextReq, Sort: TPCCNewOrderSort, GetShards: TPCCNewOrderGetShard}
}

func TPCCNewOrderWareHouseHop(tx *bolt.Tx, in *proto.TrxReq) (*proto.TrxRes, error) {
	var rwSet []*proto.RWSet
	key, warehouse := getWareHouse(tx, in.Params)

	rwSet = AddReadSet(rwSet, "Warehouse", key)

	var results []string
	results = append(results, strconv.FormatFloat(float64(warehouse.W_TAX), 'f', 2, 64))

	return &proto.TrxRes{Status: proto.Status_Success, Info: in.Info, Results: results, RWSets: rwSet}, nil
}

func TPCCNewOrderStockHop(tx *bolt.Tx, in *proto.TrxReq) (*proto.TrxRes, error) {
	var rwSet []*proto.RWSet

	wid, _ := strconv.ParseUint(in.Params["iwid"], 10, 64)
	iid, _ := strconv.ParseUint(in.Params["iid"], 10, 64)

	key, item := getItem(tx, wid, iid)

	rwSet = AddWriteSet(rwSet, "Item", key)

	itemBytes, _ := json.Marshal(item)
	putData(tx.Bucket([]byte("tItem")), key, itemBytes)

	var results []string
	results = append(results, strconv.FormatFloat(float64(item.I_PRICE), 'f', 2, 64))

	return &proto.TrxRes{Status: proto.Status_Success, Info: in.Info, Results: results, RWSets: rwSet}, nil
}

func TPCCNewOrderDistrictHop(tx *bolt.Tx, in *proto.TrxReq) (*proto.TrxRes, error) {
	var rwSet []*proto.RWSet

	wid, _ := strconv.ParseUint(in.Params["wid"], 10, 64)
	did, _ := strconv.ParseUint(in.Params["did"], 10, 64)
	cid, _ := strconv.ParseUint(in.Params["cid"], 10, 64)

	dKey, d := getDistrict(tx, wid, did)

	oid := d.D_NEXT_O_ID
	d.D_NEXT_O_ID += 1

	dBytes, _ := json.Marshal(d)
	putData(tx.Bucket([]byte("tDistrict")), dKey, dBytes)
	rwSet = AddWriteSet(rwSet, "District", dKey)

	newOrder := NewOrder{
		NO_O_ID: oid,
		NO_D_ID: did,
		NO_W_ID: wid,
	}
	newOrderBytes, _ := json.Marshal(newOrder)
	putData(tx.Bucket([]byte("tNewOrder")), newOrderBytes, []byte{})

	order := Order{
		O_KEY:       newOrder,
		O_C_ID:      cid,
		O_OL_CNT:    numOrderLine,
		O_ALL_LOCAL: false,
	}
	orderBytes, _ := json.Marshal(order)
	putData(tx.Bucket([]byte("tOrder")), newOrderBytes, orderBytes)

	cKey, _ := getCustomer(tx, wid, did, cid)
	rwSet = AddWriteSet(rwSet, "Customer", cKey)

	return &proto.TrxRes{Status: proto.Status_Success, Info: in.Info, RWSets: rwSet}, nil
}

func TPCCNewOrderNextReq(in *proto.TrxRes, params map[string]string,
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
		fallthrough
	case 2:
		fallthrough
	case 3:
		fallthrough
	case 4:
		fallthrough
	case 5:
		fallthrough
	case 6:
		fallthrough
	case 7:
		fallthrough
	case 8:
		fallthrough
	case 9:
		fallthrough
	case 10:
		// Set the parameter of the repeated hop
		hop := strconv.FormatUint(uint64(hopId), 10)
		params["iwid"] = params["iwid"+hop]
		params["iid"] = params["iid"+hop]

		// Find the hopId by warehouse id and item id
		w, _ := strconv.Atoi(params["iwid"])
		i, _ := strconv.Atoi(params["iid"])
		nextShard = w*10 + i%10
	case 11:

		// Find the hopId by warehouseid and district id
		w, _ := strconv.Atoi(params["wid"])
		d, _ := strconv.Atoi(params["did"])
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

func TPCCNewOrderGetShard(params map[string]string) []int32 {
	shards := make([]int32, 12)

	wid, _ := strconv.Atoi(params["wid"])
	shards[0] = int32(wid * 10)

	for i := 1; i < 11; i++ {
		hop := strconv.FormatUint(uint64(i), 10)
		iwid, _ := strconv.Atoi(params["iwid"+hop])
		iid, _ := strconv.Atoi(params["iid"+hop])
		shards[i] = int32(iwid*10 + iid%10)
	}

	d, _ := strconv.Atoi(params["did"])
	shards[11] = int32(wid*10 + d)

	return shards
}

type ItemTuple struct {
	iid  string
	iwid string
	node int
}

func TPCCNewOrderSort(params map[string]string) map[string]string {
	newParams := make(map[string]string)

	itemList := make([]ItemTuple, numOrderLine)
	for i := 1; i <= numOrderLine; i++ {
		hop := strconv.FormatUint(uint64(i), 10)

		// Find the hopId by warehouse id and item id
		wid, _ := strconv.Atoi(params["iwid"+hop])
		iid, _ := strconv.Atoi(params["iid"+hop])
		node := wid*10 + iid%10

		itemList[i-1] = ItemTuple{params["iid"+hop], params["iwid"+hop], node}
	}

	sort.Slice(itemList, func(i, j int) bool {
		if itemList[i].node != itemList[j].node {
			return itemList[i].node < itemList[j].node
		} else if itemList[i].iwid != itemList[j].iwid {
			return itemList[i].iwid < itemList[j].iwid
		} else {
			return itemList[i].iid < itemList[j].iid
		}
	})

	for k, v := range params {
		newParams[k] = v
	}

	for i := 1; i <= numOrderLine; i++ {
		newParams["iid"+strconv.FormatUint(uint64(i), 10)] = itemList[i-1].iid
		newParams["iwid"+strconv.FormatUint(uint64(i), 10)] = itemList[i-1].iwid
	}

	return newParams
}
