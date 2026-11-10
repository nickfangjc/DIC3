package chainManager

import (
	"encoding/json"
	"github.com/boltdb/bolt"
)

const NumWarehouse = 4
const NumDPerW = 10
const NumCPerD = 3000
const NumItem = 100000

type Warehouse struct {
	// Key: W_ID
	W_ID  uint64
	W_TAX float32
	W_YTD float32
}

type DistrictKey struct {
	D_ID   uint64
	D_W_ID uint64
}

type District struct {
	D_KEY        DistrictKey
	D_TAX        float32
	D_YTD        float32
	D_NEXT_O_ID  uint64
	D_NEXT_NO_ID uint64
}

type CustomerKey struct {
	C_ID   uint64
	C_D_ID uint64
	C_W_ID uint64
}

type Customer struct {
	// Key: C_W_ID, C_D_ID, C_ID
	C_KEY        CustomerKey
	C_CREDIT     string
	C_CREDIT_LIM float32
	C_DISCOUNT   float32
	C_BALANCE    float32
	C_DATA       string
}

// History is insert only, no primary key
type History struct {
	H_C_ID   uint64
	H_C_D_ID uint64
	H_C_W_ID uint64
	H_D_ID   uint64
	H_W_ID   uint64
	H_AMOUNT float32
}

type NewOrder struct {
	// Key: NO_W_ID, NO_D_ID, NO_O_ID
	NO_O_ID uint64
	NO_D_ID uint64
	NO_W_ID uint64
}

type Order struct {
	// Key: O_W_ID, O_D_ID, O_ID
	O_KEY        NewOrder
	O_C_ID       uint64
	O_OL_CNT     uint64
	O_ALL_LOCAL  bool
	O_CARRIER_ID uint64
}

type OrderLineKey struct {
	OL_O_ID   uint64
	OL_D_ID   uint64
	OL_W_ID   uint64
	OL_NUMBER uint64
}

type OrderLine struct {
	// Key: OL_W_ID, OL_D_ID, OL_O_ID, OL_NUMBER
	OL_KEY         OrderLineKey
	OL_I_ID        uint64
	OL_SUPPLY_W_ID uint64
}

type ItemKey struct {
	W_ID uint64
	I_ID uint64
}

type Item struct {
	I_KEY   ItemKey
	I_PRICE float32
}

type Stock struct {
	S_KEY        ItemKey
	S_QUANTITY   uint64
	S_DIST       string
	S_YTD        uint64
	S_ORDER_CNT  uint64
	S_REMOTE_CNT uint64
}

func TPCCScheds() {
	LockSchedulers["Warehouse"] = NewScheduler()
	LockSchedulers["Item"] = NewScheduler()
	LockSchedulers["District"] = NewScheduler()
	LockSchedulers["Customer"] = NewScheduler()

	AccessSchedulers["District"] = NewScheduler()
	AccessSchedulers["Item"] = NewScheduler()
	AccessSchedulers["Customer"] = NewScheduler()

	AccessDSchedulers["District"] = NewScheduler()
	AccessDSchedulers["Item"] = NewScheduler()
	AccessDSchedulers["Customer"] = NewScheduler()
}

func TPCCChains(db *bolt.DB) []*Chain {
	return []*Chain{TPCCNewOrderChainImpl(db), TPCCPaymentChainImpl(db), TPCCDeliveryChainImpl(db)}
}

func getWareHouse(tx *bolt.Tx, params map[string]string) ([]byte, Warehouse) {
	warehouses := tx.Bucket([]byte("tWarehouses"))
	widBytes := []byte(params["wid"])
	var warehouse Warehouse
	json.Unmarshal(warehouses.Get(widBytes), &warehouse)
	return widBytes, warehouse
}

func getDistrict(tx *bolt.Tx, wid uint64, did uint64) ([]byte, District) {
	districts := tx.Bucket([]byte("tDistrict"))

	dKey := DistrictKey{
		D_ID:   did,
		D_W_ID: wid,
	}

	dKeyBytes, _ := json.Marshal(dKey)
	var d District
	json.Unmarshal(districts.Get(dKeyBytes), &d)
	return dKeyBytes, d
}

func getCustomer(tx *bolt.Tx, wid uint64, did uint64, cid uint64) ([]byte, Customer) {
	customers := tx.Bucket([]byte("tCustomer"))

	cKey := CustomerKey{
		C_ID:   cid,
		C_D_ID: did,
		C_W_ID: wid,
	}

	cKeyBytes, _ := json.Marshal(cKey)
	var c Customer
	json.Unmarshal(customers.Get(cKeyBytes), &c)
	return cKeyBytes, c
}

func getOrder(tx *bolt.Tx, wid uint64, did uint64, oid uint64) ([]byte, Order) {
	orders := tx.Bucket([]byte("tOrder"))

	OKey := NewOrder{
		NO_O_ID: oid,
		NO_D_ID: did,
		NO_W_ID: wid,
	}

	OKeyBytes, _ := json.Marshal(OKey)
	var o Order
	json.Unmarshal(orders.Get(OKeyBytes), &o)
	return OKeyBytes, o
}

func getItem(tx *bolt.Tx, wid uint64, iid uint64) ([]byte, Item) {
	items := tx.Bucket([]byte("tItem"))
	itemKey := ItemKey{
		I_ID: iid,
		W_ID: wid,
	}

	itemKeyBytes, _ := json.Marshal(itemKey)
	var item Item
	json.Unmarshal(items.Get(itemKeyBytes), &item)

	return itemKeyBytes, item
}
