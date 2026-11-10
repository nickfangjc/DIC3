package main

import (
	chainMgr "ShardDB/chainManager"
	"ShardDB/server"
	"ShardDB/util"
	"math/rand"
	"strconv"
)

func main() {
	rand.Seed(2068)
	//rand.Seed(time.Now().UnixNano())
	client := server.SetupClient(2)
	client.ExpTime = 15
	client.WarmUpTime = 5
	client.PrintChain[0] = "T1"
	client.PrintChain[1] = "T2"
	numShards := client.ShardNumber
	client.GenRequest = func(threadId int, count int, isReady bool) (int, map[string]string, int) {
		if !isReady {
			// Warmup using NewOrder transaction
			return genT1(numShards)
		}

		txnType := rand.Intn(100)

		if txnType < 50 {
			return genT1(numShards)
		} else {
			return genT2(numShards)
		}
	}

	client.StartExperiment()
}

func genT1(numShards int) (int, map[string]string, int) {
	params := genIdsWithLocality(numShards)

	shardId, _ := strconv.ParseUint(params["Aid"], 10, 64)
	return 0, params, int(shardId / chainMgr.SimpleBMnumAperServer)
}

func genT2(numShards int) (int, map[string]string, int) {
	params := genIdsWithLocality(numShards)

	shardId := uint64(0)
	if !util.IsTransactionReordered && util.IsWoundWaitLock {
		shardId, _ = strconv.ParseUint(params["Bid"], 10, 64)
		shardId = shardId / chainMgr.SimpleBMnumBperServer
	} else {
		shardId, _ = strconv.ParseUint(params["Aid"], 10, 64)
		shardId = shardId / chainMgr.SimpleBMnumAperServer
	}

	return 1, params, int(shardId)
}

func genIds(numShards int) map[string]string {
	return map[string]string{
		"Aid": strconv.FormatInt(int64(rand.Intn(numShards*chainMgr.SimpleBMnumAperServer)), 10),
		"Bid": strconv.FormatInt(int64(rand.Intn(numShards*chainMgr.SimpleBMnumBperServer)), 10),
		"Cid": strconv.FormatInt(int64(rand.Intn(numShards*chainMgr.SimpleBMnumCperServer)), 10),
		"Did": strconv.FormatInt(int64(rand.Intn(numShards*chainMgr.SimpleBMnumDperServer)), 10),
		"Eid": strconv.FormatInt(int64(rand.Intn(numShards*chainMgr.SimpleBMnumEperServer)), 10),
	}
}

func genIdsWithLocality(numShards int) map[string]string {
	aid := rand.Intn(numShards * chainMgr.SimpleBMnumAperServer)
	aServer := aid / chainMgr.SimpleBMnumAperServer
	aCluster := aServer / 10

	bid := rand.Intn(numShards * chainMgr.SimpleBMnumBperServer)
	bCluster := bid / chainMgr.SimpleBMnumBperServer / 10
	cid := rand.Intn(numShards * chainMgr.SimpleBMnumCperServer)
	cCluster := cid / chainMgr.SimpleBMnumCperServer / 10
	if numShards > 1 && rand.Intn(100) < 1 {
		for bCluster == aCluster {
			bid = rand.Intn(numShards * chainMgr.SimpleBMnumBperServer)
			bCluster = bid / chainMgr.SimpleBMnumBperServer / 10
		}
	} else {
		bServer := aCluster*10 + rand.Intn(10)
		bid = bServer*chainMgr.SimpleBMnumBperServer + rand.Intn(chainMgr.SimpleBMnumBperServer)
	}

	if numShards > 1 && rand.Intn(100) < 1 {
		for cCluster == aCluster {
			cid = rand.Intn(numShards * chainMgr.SimpleBMnumCperServer)
			cCluster = cid / chainMgr.SimpleBMnumCperServer / 10
		}
	} else {
		cServer := aCluster*10 + rand.Intn(10)
		cid = cServer*chainMgr.SimpleBMnumCperServer + rand.Intn(chainMgr.SimpleBMnumCperServer)
		cCluster = cid / chainMgr.SimpleBMnumCperServer / 10
	}

	did := rand.Intn(numShards * chainMgr.SimpleBMnumDperServer)
	dCluster := did / chainMgr.SimpleBMnumDperServer / 10
	eid := rand.Intn(numShards * chainMgr.SimpleBMnumEperServer)
	eCluster := eid / chainMgr.SimpleBMnumEperServer / 10
	if numShards > 1 && rand.Intn(100) < 20 {
		for dCluster == cCluster {
			did = rand.Intn(numShards * chainMgr.SimpleBMnumDperServer)
			dCluster = did / chainMgr.SimpleBMnumDperServer / 10
		}
		for eCluster == cCluster {
			eid = rand.Intn(numShards * chainMgr.SimpleBMnumEperServer)
			eCluster = eid / chainMgr.SimpleBMnumEperServer / 10
		}
	} else {
		dServer := cCluster*10 + rand.Intn(10)
		did = dServer*chainMgr.SimpleBMnumDperServer + rand.Intn(chainMgr.SimpleBMnumDperServer)
		eServer := cCluster*10 + rand.Intn(10)
		eid = eServer*chainMgr.SimpleBMnumEperServer + rand.Intn(chainMgr.SimpleBMnumEperServer)
	}

	ids := map[string]string{
		"Aid": strconv.FormatInt(int64(aid), 10),
		"Bid": strconv.FormatInt(int64(bid), 10),
		"Cid": strconv.FormatInt(int64(cid), 10),
		"Did": strconv.FormatInt(int64(did), 10),
		"Eid": strconv.FormatInt(int64(eid), 10),
	}

	return ids
}
