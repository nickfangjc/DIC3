package main

import (
	chainMgr "ShardDB/chainManager"
	"ShardDB/server"
	"flag"
	"math/rand"
	"strconv"
)

func main() {
	warehouses := flag.Int("w", 1, "Number of warehouses")

	rand.Seed(2068)
	//rand.Seed(time.Now().UnixNano())

	client := server.SetupClient(3)
	client.ExpTime = 15
	client.WarmUpTime = 5
	client.PrintChain[0] = "NewOrder"
	client.PrintChain[1] = "Payment"
	client.PrintChain[2] = "Delivery"

	client.GenRequest = func(threadId int, count int, isReady bool) (int, map[string]string, int) {
		if !isReady {
			// Warmup using NewOrder transaction
			return genNewOrder(*warehouses)
		}

		txnType := rand.Intn(100)

		if txnType < 48 {
			return genNewOrder(*warehouses)
		} else if txnType < 96 {
			return genPayment(*warehouses)
		} else {
			return genDelivery(*warehouses)
		}
	}

	client.StartExperiment()
}

func genNewOrder(warehouses int) (int, map[string]string, int) {
	params := map[string]string{
		"wid": strconv.FormatInt(int64(rand.Intn(warehouses)), 10),
		"did": strconv.FormatInt(int64(rand.Intn(chainMgr.NumDPerW)), 10),
		"cid": strconv.FormatInt(int64(rand.Intn(chainMgr.NumCPerD)), 10)}
	for i := 1; i < 11; i++ {
		hop := strconv.FormatUint(uint64(i), 10)

		wid := params["wid"]
		if warehouses > 1 && rand.Intn(100) < 1 {
			for wid == params["wid"] {
				wid = strconv.FormatInt(int64(rand.Intn(warehouses)), 10)
			}
		}

		params["iwid"+hop] = wid
		params["iid"+hop] = strconv.FormatInt(int64(rand.Intn(chainMgr.NumItem)), 10)
	}

	shardId, _ := strconv.ParseUint(params["wid"], 10, 64)
	shardId = shardId * chainMgr.NumDPerW
	return 0, params, int(shardId)
}

func genPayment(warehouses int) (int, map[string]string, int) {
	wid := int64(rand.Intn(warehouses))
	cwid := wid
	if warehouses > 1 && rand.Intn(100) < 15 {
		for wid == cwid {
			cwid = int64(rand.Intn(warehouses))
		}
	}

	params := map[string]string{
		"wid":    strconv.FormatInt(wid, 10),
		"did":    strconv.FormatInt(int64(rand.Intn(chainMgr.NumDPerW)), 10),
		"cwid":   strconv.FormatInt(cwid, 10),
		"cdid":   strconv.FormatInt(int64(rand.Intn(chainMgr.NumDPerW)), 10),
		"cid":    strconv.FormatInt(int64(rand.Intn(chainMgr.NumCPerD)), 10),
		"amount": strconv.FormatInt(int64(rand.Intn(100)), 10)}

	shardId, _ := strconv.ParseUint(params["wid"], 10, 64)
	shardId = shardId * chainMgr.NumDPerW
	return 1, params, int(shardId)
}

func genDelivery(warehouses int) (int, map[string]string, int) {
	params := map[string]string{
		"wid":     strconv.FormatInt(int64(rand.Intn(warehouses)), 10),
		"carrier": strconv.FormatInt(int64(10), 10)}

	shardId, _ := strconv.ParseUint(params["wid"], 10, 64)
	shardId = shardId * chainMgr.NumDPerW
	return 2, params, int(shardId)
}
