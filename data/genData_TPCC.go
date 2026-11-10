package main

import (
	chainMgr "ShardDB/chainManager"
	"encoding/json"
	"flag"
	"github.com/boltdb/bolt"
	"log"
	"math/rand"
	"os"
	"strconv"
	"time"
)

var tables map[string]*bolt.Bucket

func main() {
	numWhse := flag.Int("w", 4, "number of warehouses")
	numDperW := chainMgr.NumDPerW
	rand.Seed(time.Now().UnixNano())

	flag.Parse()

	// Each wareHouse
	for w := 0; w < *numWhse; w++ {
		for d := 0; d < numDperW; d++ {
			dbname := "District" + strconv.FormatInt(int64(w*numDperW+d), 10) + ".db"
			os.Remove(dbname)
			db, err := bolt.Open(dbname, 0600, nil)
			if err != nil {
				log.Fatal(err)
			}

			err = db.Update(func(tx *bolt.Tx) error {
				if d == 0 {
					LoadWareHouse(tx, w)
				}
				CreateTables(tx)
				LoadDistrict(w, d)
				LoadCustomer(w, d)
				LoadStockItem(w, d)

				return nil
			})
		}
	}
}

func LoadWareHouse(tx *bolt.Tx, w int) {
	warehouses, _ := tx.CreateBucketIfNotExists([]byte("tWarehouses"))

	warehouse := chainMgr.Warehouse{
		W_ID:  uint64(w),
		W_TAX: rand.Float32() / 10,
		W_YTD: 0,
	}

	key := []byte(strconv.FormatInt(int64(w), 10))
	wBytes, _ := json.Marshal(warehouse)

	warehouses.Put(key, wBytes)
}

func LoadDistrict(w int, d int) {
	dKey := chainMgr.DistrictKey{
		D_ID:   uint64(d),
		D_W_ID: uint64(w),
	}
	district := chainMgr.District{
		D_KEY:       dKey,
		D_TAX:       rand.Float32() / 10,
		D_YTD:       0,
		D_NEXT_O_ID: 0,
	}

	dKeyBytes, _ := json.Marshal(dKey)
	districtBytes, _ := json.Marshal(district)

	tables["district"].Put(dKeyBytes, districtBytes)
}

func LoadCustomer(w int, d int) {
	cKey := chainMgr.CustomerKey{
		C_ID:   0,
		C_D_ID: uint64(d),
		C_W_ID: uint64(w),
	}
	customer := chainMgr.Customer{
		C_KEY:        cKey,
		C_CREDIT:     "",
		C_CREDIT_LIM: 0,
		C_DISCOUNT:   0,
		C_BALANCE:    0,
	}

	for c := 0; c < chainMgr.NumCPerD; c++ {
		cKey.C_ID = uint64(c)
		customer.C_KEY.C_ID = uint64(c)
		cKeyBytes, _ := json.Marshal(cKey)
		customerBytes, _ := json.Marshal(customer)

		tables["customer"].Put(cKeyBytes, customerBytes)
	}
}

func LoadStockItem(w int, d int) {
	itemKey := chainMgr.ItemKey{
		W_ID: uint64(w),
		I_ID: 0,
	}
	item := chainMgr.Item{
		I_KEY:   itemKey,
		I_PRICE: rand.Float32() * 100,
	}
	stock := chainMgr.Stock{
		S_KEY:        itemKey,
		S_QUANTITY:   uint64(rand.Intn(90) + 10),
		S_DIST:       "",
		S_YTD:        0,
		S_ORDER_CNT:  0,
		S_REMOTE_CNT: 0,
	}
	numIPerD := chainMgr.NumItem / chainMgr.NumDPerW

	for i := 0; i < numIPerD; i++ {
		itemKey.I_ID = uint64(i*chainMgr.NumDPerW + d)
		item.I_KEY = itemKey
		stock.S_KEY = itemKey

		iKeyBytes, _ := json.Marshal(itemKey)
		itemBytes, _ := json.Marshal(item)
		stockBytes, _ := json.Marshal(stock)

		tables["item"].Put(iKeyBytes, itemBytes)
		tables["stock"].Put(iKeyBytes, stockBytes)
	}
}

func CreateTables(tx *bolt.Tx) {
	tables = make(map[string]*bolt.Bucket)

	tables["district"], _ = tx.CreateBucketIfNotExists([]byte("tDistrict"))
	tables["customer"], _ = tx.CreateBucketIfNotExists([]byte("tCustomer"))
	tables["stock"], _ = tx.CreateBucketIfNotExists([]byte("tStock"))
	tables["item"], _ = tx.CreateBucketIfNotExists([]byte("tItem"))
	tables["newOrder"], _ = tx.CreateBucketIfNotExists([]byte("tNewOrder"))
	tables["order"], _ = tx.CreateBucketIfNotExists([]byte("tOrder"))
	tables["orderLine"], _ = tx.CreateBucketIfNotExists([]byte("tOrderLine"))
	tables["history"], _ = tx.CreateBucketIfNotExists([]byte("tHistory"))
}
