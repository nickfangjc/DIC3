package main

import (
	chainMgr "ShardDB/chainManager"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/boltdb/bolt"
	"log"
	"math/rand"
	"os"
	"strconv"
	"time"
)

func main() {
	numShards := flag.Int("shards", 4, "number of shards")
	numItems := flag.Int("items", 100, "number of items of per server")
	rand.Seed(time.Now().UnixNano())

	flag.Parse()
	numServers := *numShards * 10

	for w := 0; w < *numShards; w++ {
		for d := 0; d < 10; d++ {
			dbname := "District" + strconv.FormatInt(int64(w*10+d), 10) + ".db"
			os.Remove(dbname)
			db, err := bolt.Open(dbname, 0600, nil)
			if err != nil {
				log.Fatal(err)
			}

			err = db.Update(func(tx *bolt.Tx) error {
				items, err := tx.CreateBucketIfNotExists([]byte("Items"))
				if err != nil {
					return fmt.Errorf("could not create AItem bucket: %v", err)
				}

				iid := uint64(10*w + d)
				for i := 0; i < *numItems; i += *numShards {
					itemKey := chainMgr.ItemsKey{
						Iid: iid,
					}
					item := chainMgr.Items{
						Key:      itemKey,
						Price:    rand.Float32() * 100,
						Category: rand.Intn(10),
					}

					keybytes, _ := json.Marshal(itemKey)
					itemBytes, _ := json.Marshal(item)
					items.Put(keybytes, itemBytes)

					iid += uint64(i * numServers)
				}

				return nil
			})
		}
	}

}
