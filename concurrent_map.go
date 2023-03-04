package auto_blacklist

import "sync"

type Hashcode interface {
	Hash() int
}

type ConcurrentHashMap struct {
	shardLen int
	shardMap []*MapShard
}

func NewConcurrentHashMap(shardLen int) *ConcurrentHashMap {
	c := &ConcurrentHashMap{
		shardLen: shardLen,
		shardMap: make([]*MapShard, shardLen),
	}
	for i := 0; i < shardLen; i++ {
		c.shardMap[i] = NewMapShard()
	}
	return c
}

type MapShard struct {
	sync.RWMutex
	entries map[interface{}]interface{}
}

func NewMapShard() *MapShard {
	return &MapShard{
		entries: make(map[interface{}]interface{}),
	}
}

func (ms *MapShard) put(k, v interface{}) {
	ms.Lock()
	defer ms.Unlock()
	ms.entries[k] = v
}

func (ms *MapShard) putIfNotExists(k, v interface{}) interface{} {
	ms.Lock()
	defer ms.Unlock()
	if o, exists := ms.entries[k]; exists {
		return o
	}
	ms.entries[k] = v
	return v
}

func (ms *MapShard) get(k interface{}) interface{} {
	ms.RLock()
	defer ms.RUnlock()
	return ms.entries[k]
}

func (chm *ConcurrentHashMap) Put(key Hashcode, val interface{}) {
	chm.getShard(key).put(key, val)
}

func (chm *ConcurrentHashMap) PutIfNotExists(key Hashcode, val interface{}) interface{} {
	return chm.getShard(key).putIfNotExists(key, val)
}

func (chm *ConcurrentHashMap) Get(key Hashcode) interface{} {
	return chm.getShard(key).get(key)
}

func (chm *ConcurrentHashMap) getShard(key Hashcode) *MapShard {
	return chm.shardMap[key.Hash()%chm.shardLen]
}
