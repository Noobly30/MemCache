package cache

import (
	"fmt"
	"log"
	"sync"
	"time"
)

type memCache struct{
	maxMemorySize int64
	maxMemorySizeStr string
	currMemorySize int64
	values map[string]*memCacheValue
	locker sync.RWMutex
	clearExpiredItemTimeInterval time.Duration
}

type memCacheValue struct {
	val 	interface{} 
	expireTime	time.Time
	expire time.Duration
	size int64
}

func NewMemCache() Cache {
	mc := &memCache{
		values: make(map[string]*memCacheValue),
		clearExpiredItemTimeInterval: time.Second * 10,
	}
	go mc.clearExpiredItem()
	return mc
}

func (mc *memCache) SetMaxMemory(size string) bool{
	mc.maxMemorySize, mc.maxMemorySizeStr = ParseSize(size)
	return false
}

func (mc *memCache) Set(key string, val interface{}, expire time.Duration) bool{
	mc.locker.Lock()
	defer mc.locker.Unlock()
	fmt.Println("called set")
	v := &memCacheValue{
		val: 		val,
		expireTime: 	time.Now().Add(expire),
		size:		GetValSize(val),
	}
	mc.del(key)
	mc.add(key, v)
	if mc.currMemorySize > mc.maxMemorySize {
		mc.del(key)
		log.Println(fmt.Sprintf("max memory size %s", mc.maxMemorySize))
		panic(fmt.Sprintf("max memory size %s", mc.maxMemorySize))

	}
	return true
}

func (mc *memCache) get(key string) (*memCacheValue, bool){
	val, ok := mc.values[key]
	return val, ok
}

func (mc *memCache) del(key string){
	tmp, ok := mc.get(key)
	if ok && tmp != nil {
		mc.currMemorySize -= tmp.size
		delete(mc.values, key)
	}

}

func (mc *memCache) add(key string, val *memCacheValue){
	mc.values[key] = val
	mc.currMemorySize += val.size
}

func (mc *memCache) Get(key string) (interface{},bool){
	mc.locker.RLock()
	defer mc.locker.Unlock()
	mcv, ok := mc.get(key)
	if ok{
		if mcv.expireTime.Before(time.Now()){
			mc.del(key)
			return nil, false
		}
		return mcv.val, ok
	}
	return nil, false
}

func (mc *memCache) Del(key string) bool{
	mc.locker.Lock()
	defer mc.locker.Unlock()
	mc.del(key)
	return true
}

func (mc *memCache)	Exists(key string) bool{
	mc.locker.RLock()
	defer mc.locker.RUnlock()
	_ , ok := mc.values[key]
	return ok
}

func (mc *memCache)	Flush() bool{
	mc.locker.Lock()
	defer mc.locker.Unlock()
	mc.values = make(map[string]*memCacheValue, 0)
	mc.currMemorySize = 0
	return true
}

func (mc *memCache)	Keys() int64{
	mc.locker.RLock()
	defer mc.locker.RUnlock()
	return int64(len(mc.values))
}

func (mc *memCache) clearExpiredItem() {
	timeTicker := time.NewTicker(mc.clearExpiredItemTimeInterval)
	defer timeTicker.Stop()
	for{
		select{
		case <- timeTicker.C:
			for key, item := range mc.values{
				if item.expire != 0 && time.Now().After(item.expireTime){
					mc.locker.Lock()
					mc.del(key)
					mc.locker.Unlock()
				}
			}
		}
	}
}

