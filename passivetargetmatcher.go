package main

import (
	"fmt"
	"sync"
	"time"
)

/*----------------------------------------------------
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
	//var cache *SkuCache
	cache := passiveTargetMatcher.skuCache
	return cache
}

func (passiveTargetMatcher *PassiveTargetMatcher) updateCache(cache *SkuCache) {
	passiveTargetMatcher.rwLock.Lock() //Writer lock
	passiveTargetMatcher.skuCache = cache
	fmt.Println("Updated skuCache")
	passiveTargetMatcher.rwLock.Unlock()
}

func (passiveTargetMatcher *PassiveTargetMatcher) findMatch(sku string) string {
	passiveTargetMatcher.rwLock.RLock()        //Reader Lock
	passiveTargetMatcher.getCache().match(sku) //Use CPU
	passiveTargetMatcher.rwLock.RUnlock()

	//Simulate Sync IO
	time.Sleep(50 * time.Microsecond)

	passiveTargetMatcher.rwLock.RLock()                    //Reader Lock
	strategy := passiveTargetMatcher.getCache().match(sku) //Use CPU
	passiveTargetMatcher.rwLock.RUnlock()
	return strategy
}
