package chainManager

import (
	"ShardDB/proto"
	"ShardDB/util"
	"encoding/json"
	"github.com/boltdb/bolt"
	"strconv"
)

type ItemsKey struct {
	Iid uint64
}

type Items struct {
	Key      ItemsKey
	Price    float32
	Category int
}

func RubisScheds() {
	LockSchedulers["Items"] = NewScheduler()

	AccessSchedulers["Items"] = NewScheduler()
}

func RubisChains(db *bolt.DB) []*Chain {
	return []*Chain{RubisGetItemByCategoryImpl(db)}
}

func RubisGetItemByCategoryImpl(db *bolt.DB) *Chain {
	// Create the list of hops
	hops := make([]*Hop, 1)

	hops[0] = &Hop{
		id:         0,
		hopType:    util.GlobalHop,
		isReadOnly: true,
		process:    RubisGetItemByCategoryHop,
		aggregate:  RubisGetItemByCategoryHopAggregate,
		conflicts:  map[int32]int32{},
	}

	return &Chain{db: db, hops: hops, NextReq: RubisGetItemByCategoryNextReq}
}

func RubisGetItemByCategoryHop(tx *bolt.Tx, in *proto.TrxReq) (*proto.TrxRes, error) {
	var rwSet []*proto.RWSet
	category, _ := strconv.ParseUint(in.Params["Category"], 10, 64)

	b := tx.Bucket([]byte("Items"))
	c := b.Cursor()

	item := Items{}
	itemList := make([]Items, 0)
	for k, v := c.First(); k != nil; k, v = c.Next() {
		json.Unmarshal(v, &item)
		if item.Category == int(category) {
			itemList = append(itemList, item)
			rwSet = AddReadSet(rwSet, "Items", k)
		}
	}

	itemString, _ := json.Marshal(itemList)
	var results []string
	results = append(results, string(itemString))

	return &proto.TrxRes{Status: proto.Status_Success, Info: in.Info, Results: results, RWSets: rwSet}, nil
}

func RubisGetItemByCategoryHopAggregate(results [][]string, params map[string]string) map[string]string {
	var items []Items
	params["itemList"] = ListConcatenate(results, items)
	return params
}

func RubisGetItemByCategoryNextReq(in *proto.TrxRes, params map[string]string,
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
		// nothing to do
		break
	default:
		return nil, params, 0
	}

	// Fixed steps
	req.Params = params
	req.Info.Hopid += 1
	req.Dependency = in.Dependency
	return req, params, nextShard
}
