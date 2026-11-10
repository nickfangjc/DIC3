package util

import (
	"sync"
)

type AccessTuple struct {
	TxId   uint64
	IsRead bool
}

type AccessList struct {
	lock sync.Mutex
	list []AccessTuple
}

func NewAccessList() *AccessList {
	return &AccessList{
		lock: sync.Mutex{},
		list: nil,
	}
}

func (list *AccessList) GetList() []AccessTuple {
	list.lock.Lock()
	defer list.lock.Unlock()

	rList := list.list
	return rList
}

func (list *AccessList) Remove(txId uint64) {
	list.lock.Lock()
	defer list.lock.Unlock()

	id := -1
	for i, v := range list.list {
		if v.TxId == txId {
			id = i
			break
		}
	}

	list.list = RemoveElement(list.list, id)
}

func (list *AccessList) Add(txId uint64, isRead bool) []AccessTuple {
	list.lock.Lock()
	defer list.lock.Unlock()

	oldList := make([]AccessTuple, len(list.list))
	copy(oldList, list.list)
	list.list = append(list.list, AccessTuple{TxId: txId, IsRead: isRead})
	return oldList
}

type AccessDTuple struct {
	TxId   uint64
	HopId  int32
	Shards []int32
	IsRead bool
}
type AccessDList struct {
	lock sync.Mutex
	list []AccessDTuple
}

func NewAccessDList() *AccessDList {
	return &AccessDList{
		lock: sync.Mutex{},
		list: nil,
	}
}

func (list *AccessDList) Remove(txId uint64) {
	list.lock.Lock()
	defer list.lock.Unlock()

	id := -1
	for i, v := range list.list {
		if v.TxId == txId {
			id = i
			break
		}
	}

	if id != -1 {
		list.list = append(list.list[:id], list.list[id+1:]...)
	}
}

func (list *AccessDList) Add(txId uint64, hopId int32, shards []int32, isRead bool) []AccessDTuple {
	list.lock.Lock()
	defer list.lock.Unlock()

	oldList := make([]AccessDTuple, len(list.list))
	copy(oldList, list.list)
	list.list = append(list.list, AccessDTuple{TxId: txId, HopId: hopId, Shards: shards, IsRead: isRead})
	return oldList
}
