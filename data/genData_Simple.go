package main

import (
	chainMgr "ShardDB/chainManager"
	"flag"
	"log"
	"os"
	"strconv"

	"github.com/boltdb/bolt"
)

func main() {
	numShards := flag.Int("shards", 8, "number of shards")

	countA, countB, countC, countD, countE := 0, 0, 0, 0, 0
	for i := 0; i < *numShards; i++ {
		for d := 0; d < 10; d++ {
			dbname := "District" + strconv.FormatInt(int64(i*10+d), 10) + ".db"
			os.Remove(dbname)
			db, err := bolt.Open(dbname, 0600, nil)
			if err != nil {
				log.Fatal(err)
			}

			err = db.Update(func(tx *bolt.Tx) error {
				countA = buildTable(tx, "A", chainMgr.SimpleBMnumAperServer, countA)
				countB = buildTable(tx, "B", chainMgr.SimpleBMnumBperServer, countB)
				countC = buildTable(tx, "C", chainMgr.SimpleBMnumCperServer, countC)
				countD = buildTable(tx, "D", chainMgr.SimpleBMnumDperServer, countD)
				countE = buildTable(tx, "E", chainMgr.SimpleBMnumEperServer, countE)
				return nil
			})
		}
	}
}

func buildTable(tx *bolt.Tx, tableName string, num int, count int) int {
	b, _ := tx.CreateBucketIfNotExists([]byte("t" + tableName))
	for i := 0; i < num; i++ {
		key := []byte(strconv.FormatInt(int64(count), 10))
		b.Put(key, []byte(tableName))
		count++
	}

	return count
}
