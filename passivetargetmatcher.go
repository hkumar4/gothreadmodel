package main

import (
	"fmt"
	"sync"
	"time"
)

/**********************--------------------------------------
 				Passive Target Matcher
*-----------------------------------------------------------*/

type PassiveTargetMatcher struct {
	rwLock   sync.RWMutex
	skuCache *SkuCache
}

//Factory method for constructing a TargetMatcher
func NewPassiveTargetMatcher() *PassiveTargetMatcher {
	passiveTargetMatcher := new(PassiveTargetMatcher)
	return passiveTargetMatcher
}

func (passiveTargetMatcher *PassiveTargetMatcher) getCache() *SkuCache {
	var cache *SkuCache
	passiveTargetMatcher.rwLock.RLock() //Reader Lock
	cache = passiveTargetMatcher.skuCache
	passiveTargetMatcher.rwLock.RUnlock()
	return cache
}

func (passiveTargetMatcher *PassiveTargetMatcher) updateCache(cache *SkuCache) {
	passiveTargetMatcher.rwLock.Lock() //Writer lock
	passiveTargetMatcher.skuCache = cache
	fmt.Println("Updated skuCache")
	passiveTargetMatcher.rwLock.Unlock()
}

func (passiveTargetMatcher *PassiveTargetMatcher) findMatch(sku string) string {
	passiveTargetMatcher.getCache().match(sku)
	//Simulate Sync IO
	time.Sleep(10 * time.Microsecond)
	return passiveTargetMatcher.getCache().match(sku)
}
