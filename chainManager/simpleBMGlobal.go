package chainManager

import (
	"ShardDB/proto"
	"ShardDB/util"
	"log"
	"strconv"

	"github.com/boltdb/bolt"
)

const SimpleBMnumAperServer = 1
const SimpleBMnumBperServer = 100
const SimpleBMnumCperServer = 100
const SimpleBMnumDperServer = 100
const SimpleBMnumEperServer = 100

func SimpleBMScheds() {
	LockSchedulers["A"] = NewScheduler()
	LockSchedulers["B"] = NewScheduler()
	LockSchedulers["C"] = NewScheduler()
	LockSchedulers["D"] = NewScheduler()
	LockSchedulers["E"] = NewScheduler()

	AccessSchedulers["A"] = NewScheduler()
	AccessSchedulers["B"] = NewScheduler()
	AccessSchedulers["C"] = NewScheduler()
	AccessSchedulers["D"] = NewScheduler()
	AccessSchedulers["E"] = NewScheduler()

	AccessDSchedulers["A"] = NewScheduler()
	AccessDSchedulers["B"] = NewScheduler()
	AccessDSchedulers["C"] = NewScheduler()
	AccessDSchedulers["D"] = NewScheduler()
	AccessDSchedulers["E"] = NewScheduler()
}

func SimpleBMChains(db *bolt.DB) []*Chain {
	return []*Chain{SimpleBMIC3Impl1(db), SimpleBMIC3Impl2(db)}
}

func SimpleBMIC3Impl1(db *bolt.DB) *Chain {
	hops := make([]*Hop, 5)
	hops[0] = &Hop{
		id:         0,
		hopType:    util.MergedHopBegin,
		isReadOnly: true,
		process:    SimpleBMC1HopA,
		conflicts:  map[int32]int32{1: 0},
	}

	hops[1] = &Hop{
		id:         1,
		hopType:    util.MergedHopEnd,
		isReadOnly: false,
		process:    SimpleBMC1HopB,
		conflicts:  map[int32]int32{0: 1, 1: 1},
	}

	hops[2] = &Hop{
		id:         2,
		hopType:    util.MergedHopBegin,
		isReadOnly: true,
		process:    SimpleBMC1HopC,
		conflicts:  map[int32]int32{0: 2, 1: 2},
	}

	hopD := &Hop{
		id:         3,
		hopType:    util.MergedHop,
		isReadOnly: false,
		process:    SimpleBMC1HopD,
		conflicts:  map[int32]int32{},
	}

	hopE := &Hop{
		id:         4,
		hopType:    util.MergedHopEnd,
		isReadOnly: false,
		process:    SimpleBMC1HopE,
		conflicts:  map[int32]int32{0: 2, 1: 2},
	}

	if util.IsTransactionReordered {
		hops[0].hopType = util.NormalHop
		hops[1].hopType = util.NormalHop
		hopE.id = 3
		hops[3] = hopE
		hopD.id = 4
		hopD.hopType = util.NormalHop
		hops[4] = hopD
	} else {
		log.Println("SimpleBMIC3Impl1: ", "2PC on AB")
		hops[3] = hopD
		hops[4] = hopE
	}

	return &Chain{db: db, hops: hops, NextReq: SimpleBM1NextReq, Sort: SimpleBMSort, GetShards: SimpleBMGetShards1}
}

func SimpleBMIC3Impl2(db *bolt.DB) *Chain {
	hops := make([]*Hop, 4)
	hopA := &Hop{
		id:         0,
		hopType:    util.NormalHop,
		isReadOnly: false,
		process:    SimpleBMC2HopA,
		conflicts:  map[int32]int32{0: 0, 1: 0},
	}

	hopB := &Hop{
		id:         1,
		hopType:    util.NormalHop,
		isReadOnly: false,
		process:    SimpleBMC2HopB,
		conflicts:  map[int32]int32{0: 1, 1: 1},
	}

	hopE := &Hop{
		id:         2,
		hopType:    util.MergedHopBegin,
		isReadOnly: true,
		process:    SimpleBMC2HopE,
		conflicts:  map[int32]int32{0: 2, 1: 2},
	}

	hopC := &Hop{
		id:         3,
		hopType:    util.MergedHopEnd,
		isReadOnly: false,
		process:    SimpleBMC2HopC,
		conflicts:  map[int32]int32{0: 2, 1: 2},
	}

	if util.IsTransactionReordered {
		hops[0] = hopA
		hops[1] = hopB
		log.Println("SimpleBMIC3Impl2: ", "Normal on AB")
	} else {
		log.Println("SimpleBMIC3Impl2: ", "2PC on AB")
		if util.IsWoundWaitLock {
			hopB.hopType = util.MergedHopBegin
			hopB.id = 0
			hopB.conflicts = map[int32]int32{0: 0, 1: 0}
			hopA.hopType = util.MergedHopEnd
			hopA.id = 1
			hops[0] = hopB
			hops[1] = hopA
		} else {
			hopA.hopType = util.MergedHopBegin
			hopB.hopType = util.MergedHopEnd
			hops[0] = hopA
			hops[1] = hopB
		}
	}

	if util.IsWoundWaitLock {
		hops[2] = hopE
		hops[3] = hopC
	} else {
		hopC.id = 2
		hopC.hopType = util.MergedHopBegin
		hopE.id = 3
		hopE.hopType = util.MergedHopEnd
		hops[2] = hopC
		hops[3] = hopE
	}

	return &Chain{db: db, hops: hops, NextReq: SimpleBM2NextReq, Sort: SimpleBMSort, GetShards: SimpleBMGetShards2}
}

func SimpleBMSort(params map[string]string) map[string]string {
	return params
}

func SimpleBMGetShards1(params map[string]string) []int32 {
	shards := make([]int32, 5)
	aid, _ := strconv.Atoi(params["Aid"])
	bid, _ := strconv.Atoi(params["Bid"])
	cid, _ := strconv.Atoi(params["Cid"])
	did, _ := strconv.Atoi(params["Did"])
	eid, _ := strconv.Atoi(params["Eid"])

	if util.IsTransactionReordered {
		shards[0] = int32(aid / SimpleBMnumAperServer)
		shards[1] = int32(bid / SimpleBMnumBperServer)
		shards[2] = int32(cid / SimpleBMnumCperServer)
		shards[3] = int32(eid / SimpleBMnumEperServer)
		shards[4] = int32(did / SimpleBMnumDperServer)

	} else {
		shards[0] = int32(aid / SimpleBMnumAperServer)
		shards[1] = int32(bid / SimpleBMnumBperServer)
		shards[2] = int32(cid / SimpleBMnumCperServer)
		shards[4] = int32(eid / SimpleBMnumEperServer)
		shards[3] = int32(did / SimpleBMnumDperServer)
	}
	return shards
}

func SimpleBMGetShards2(params map[string]string) []int32 {
	shards := make([]int32, 4)
	aid, _ := strconv.Atoi(params["Aid"])
	bid, _ := strconv.Atoi(params["Bid"])
	cid, _ := strconv.Atoi(params["Cid"])
	eid, _ := strconv.Atoi(params["Eid"])

	if !util.IsTransactionReordered && util.IsWoundWaitLock {
		shards[1] = int32(aid / SimpleBMnumAperServer)
		shards[0] = int32(bid / SimpleBMnumBperServer)
	} else {
		shards[0] = int32(aid / SimpleBMnumAperServer)
		shards[1] = int32(bid / SimpleBMnumBperServer)
	}

	if util.IsWoundWaitLock {
		shards[2] = int32(eid / SimpleBMnumEperServer)
		shards[3] = int32(cid / SimpleBMnumCperServer)
	} else {
		shards[3] = int32(eid / SimpleBMnumEperServer)
		shards[2] = int32(cid / SimpleBMnumCperServer)
	}

	return shards
}

func SimpleBMC1HopA(tx *bolt.Tx, in *proto.TrxReq) (*proto.TrxRes, error) {
	var rwSet []*proto.RWSet
	objectA := tx.Bucket([]byte("tA"))
	aidBytes := []byte(in.Params["Aid"])

	log.Println("SimpleBMC1HopA: ", in.Params)
	avalue := objectA.Get(aidBytes)

	rwSet = AddReadSet(rwSet, "A", aidBytes)

	var results []string
	results = append(results, string(avalue))

	return &proto.TrxRes{Status: proto.Status_Success, Info: in.Info, Results: results, RWSets: rwSet}, nil
}

func SimpleBMC2HopA(tx *bolt.Tx, in *proto.TrxReq) (*proto.TrxRes, error) {
	var rwSet []*proto.RWSet
	objectA := tx.Bucket([]byte("tA"))
	aidBytes := []byte(in.Params["Aid"])

	rwSet = AddWriteSet(rwSet, "A", aidBytes)

	putData(objectA, aidBytes, []byte("A"))
	return &proto.TrxRes{Status: proto.Status_Success, Info: in.Info, RWSets: rwSet}, nil
}

func SimpleBMC1HopB(tx *bolt.Tx, in *proto.TrxReq) (*proto.TrxRes, error) {
	var rwSet []*proto.RWSet
	objectB := tx.Bucket([]byte("tB"))
	bidBytes := []byte(in.Params["Bid"])

	rwSet = AddWriteSet(rwSet, "B", bidBytes)

	putData(objectB, bidBytes, []byte("A"))
	return &proto.TrxRes{Status: proto.Status_Success, Info: in.Info, RWSets: rwSet}, nil
}

func SimpleBMC2HopB(tx *bolt.Tx, in *proto.TrxReq) (*proto.TrxRes, error) {
	var rwSet []*proto.RWSet
	objectB := tx.Bucket([]byte("tB"))
	bidBytes := []byte(in.Params["Bid"])

	rwSet = AddWriteSet(rwSet, "B", bidBytes)

	putData(objectB, bidBytes, []byte("B"))
	return &proto.TrxRes{Status: proto.Status_Success, Info: in.Info, RWSets: rwSet}, nil
}

func SimpleBMC1HopC(tx *bolt.Tx, in *proto.TrxReq) (*proto.TrxRes, error) {
	var rwSet []*proto.RWSet
	objectC := tx.Bucket([]byte("tC"))
	cidBytes := []byte(in.Params["Cid"])

	cvalue := objectC.Get(cidBytes)

	rwSet = AddReadSet(rwSet, "C", cidBytes)

	var results []string
	results = append(results, string(cvalue))

	return &proto.TrxRes{Status: proto.Status_Success, Info: in.Info, Results: results, RWSets: rwSet}, nil
}

func SimpleBMC2HopC(tx *bolt.Tx, in *proto.TrxReq) (*proto.TrxRes, error) {
	var rwSet []*proto.RWSet
	objectC := tx.Bucket([]byte("tC"))
	cidBytes := []byte(in.Params["Cid"])

	rwSet = AddWriteSet(rwSet, "C", cidBytes)
	putData(objectC, cidBytes, []byte("C"))

	return &proto.TrxRes{Status: proto.Status_Success, Info: in.Info, RWSets: rwSet}, nil
}

func SimpleBMC1HopD(tx *bolt.Tx, in *proto.TrxReq) (*proto.TrxRes, error) {
	var rwSet []*proto.RWSet
	objectD := tx.Bucket([]byte("tD"))
	didBytes := []byte(in.Params["Did"])

	rwSet = AddWriteSet(rwSet, "D", didBytes)

	putData(objectD, didBytes, []byte("D"))
	return &proto.TrxRes{Status: proto.Status_Success, Info: in.Info, RWSets: rwSet}, nil
}

func SimpleBMC1HopE(tx *bolt.Tx, in *proto.TrxReq) (*proto.TrxRes, error) {
	var rwSet []*proto.RWSet
	objectE := tx.Bucket([]byte("tE"))
	eidBytes := []byte(in.Params["Eid"])

	rwSet = AddWriteSet(rwSet, "E", eidBytes)
	putData(objectE, eidBytes, []byte("E"))

	return &proto.TrxRes{Status: proto.Status_Success, Info: in.Info, RWSets: rwSet}, nil
}

func SimpleBMC2HopE(tx *bolt.Tx, in *proto.TrxReq) (*proto.TrxRes, error) {
	var rwSet []*proto.RWSet
	objectE := tx.Bucket([]byte("tE"))
	eidBytes := []byte(in.Params["Eid"])

	evalue := objectE.Get(eidBytes)

	rwSet = AddReadSet(rwSet, "E", eidBytes)

	var results []string
	results = append(results, string(evalue))

	return &proto.TrxRes{Status: proto.Status_Success, Info: in.Info, Results: results, RWSets: rwSet}, nil
}

func SimpleBM1NextReq(in *proto.TrxRes, params map[string]string,
	shards int) (*proto.TrxReq, map[string]string, int) {
	hopId := in.Info.Hopid
	req := &proto.TrxReq{
		Info: in.Info,
	}
	nextShard := 0

	switch hopId {
	case 0:
		nextShard, _ = strconv.Atoi(params["Aid"])
		nextShard = nextShard / SimpleBMnumAperServer
	case 1:
		nextShard, _ = strconv.Atoi(params["Bid"])
		nextShard = nextShard / SimpleBMnumBperServer
	case 2:
		nextShard, _ = strconv.Atoi(params["Cid"])
		nextShard = nextShard / SimpleBMnumCperServer
	case 3:
		if util.IsTransactionReordered {
			nextShard, _ = strconv.Atoi(params["Eid"])
			nextShard = nextShard / SimpleBMnumEperServer
		} else {
			nextShard, _ = strconv.Atoi(params["Did"])
			nextShard = nextShard / SimpleBMnumDperServer
		}

	case 4:
		if util.IsTransactionReordered {
			nextShard, _ = strconv.Atoi(params["Did"])
			nextShard = nextShard / SimpleBMnumDperServer
		} else {
			nextShard, _ = strconv.Atoi(params["Eid"])
			nextShard = nextShard / SimpleBMnumEperServer
		}

	default:
		return nil, params, 0
	}

	// Fixed steps
	req.Params = params
	req.Info.Hopid += 1
	req.Dependency = in.Dependency
	return req, params, nextShard
}

func SimpleBM2NextReq(in *proto.TrxRes, params map[string]string,
	shards int) (*proto.TrxReq, map[string]string, int) {
	hopId := in.Info.Hopid
	req := &proto.TrxReq{
		Info: in.Info,
	}
	nextShard := 0

	switch hopId {
	case 0:
		if !util.IsTransactionReordered && util.IsWoundWaitLock {
			nextShard, _ = strconv.Atoi(params["Bid"])
			nextShard = nextShard / SimpleBMnumBperServer
		} else {
			nextShard, _ = strconv.Atoi(params["Aid"])
			nextShard = nextShard / SimpleBMnumAperServer
		}
	case 1:
		if !util.IsTransactionReordered && util.IsWoundWaitLock {
			nextShard, _ = strconv.Atoi(params["Aid"])
			nextShard = nextShard / SimpleBMnumAperServer
		} else {
			nextShard, _ = strconv.Atoi(params["Bid"])
			nextShard = nextShard / SimpleBMnumBperServer
		}
	case 2:
		if util.IsWoundWaitLock {
			nextShard, _ = strconv.Atoi(params["Eid"])
			nextShard = nextShard / SimpleBMnumEperServer
		} else {
			nextShard, _ = strconv.Atoi(params["Cid"])
			nextShard = nextShard / SimpleBMnumCperServer
		}
	case 3:
		if util.IsWoundWaitLock {
			nextShard, _ = strconv.Atoi(params["Cid"])
			nextShard = nextShard / SimpleBMnumCperServer
		} else {
			nextShard, _ = strconv.Atoi(params["Eid"])
			nextShard = nextShard / SimpleBMnumEperServer
		}
	default:
		return nil, params, 0
	}

	// Fixed steps
	req.Params = params
	req.Info.Hopid += 1
	req.Dependency = in.Dependency
	return req, params, nextShard
}
