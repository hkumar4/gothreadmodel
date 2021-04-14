package main

import (
	"fmt"
	"os"
	"runtime"
	"strconv"
	"sync"
	"time"
)

/* ----------------------------------------------
 ---- Cache
-----------------------------------------*/

type SkuCache struct {
	cache map[string]string
}

func (skuCache *SkuCache) initSkuCache() {
	skuCache.cache = make(map[string]string)
}

func (skuCache *SkuCache) match(sku string) string {
	//Simulating complex search and match by doing search n times
	var str string

	for i := 0; i < 50; i++ {
		str = skuCache.cache[sku]
	}

	return str
}

/*--------------------------------------
 				Target Matcher
*-----------------------------------------------------------*/

type TargetMatcher interface {
	updateCache(skuCache *SkuCache)
	findMatch(sku string) string
}

/**--------------------------------------
 Cache Loader
*-----------------------------------------------------------*/

type TargetCacheLoader struct {
	skuCacheA        SkuCache
	skuCacheB        SkuCache
	skuCacheActive   *SkuCache
	skuCacheInactive *SkuCache
	targetMatcher    TargetMatcher
	isLoading        bool
	ticker           *time.Ticker
}

func loadTestData(skuCache *SkuCache) {
	skuCache.initSkuCache()
	//Load 10M entries
	for i := 0; i < 10000000; i++ {
		istr := strconv.Itoa(i)
		//Just load some new numbers
		skuCache.cache[istr] = strconv.Itoa(int(time.Now().UnixNano() / 1000))
	}
}

func (targetCacheLoader *TargetCacheLoader) loadAndNotify() {
	loadTestData(targetCacheLoader.skuCacheInactive)

	tempSkuCache := targetCacheLoader.skuCacheActive
	targetCacheLoader.skuCacheActive = targetCacheLoader.skuCacheInactive
	targetCacheLoader.skuCacheInactive = tempSkuCache
	targetCacheLoader.targetMatcher.updateCache(targetCacheLoader.skuCacheActive)

	targetCacheLoader.skuCacheInactive.initSkuCache()
	for k, v := range targetCacheLoader.skuCacheActive.cache {
		targetCacheLoader.skuCacheInactive.cache[k] = v
	}
}

func (targetCacheLoader *TargetCacheLoader) start() {

	targetCacheLoader.skuCacheActive = &targetCacheLoader.skuCacheA
	targetCacheLoader.skuCacheInactive = &targetCacheLoader.skuCacheB

	targetCacheLoader.loadAndNotify()

	targetCacheLoader.ticker = time.NewTicker(10 * time.Second)

	go func() {
		for {
			<-targetCacheLoader.ticker.C
			fmt.Println("Cache Loader Timer fired")
			if targetCacheLoader.isLoading {
				fmt.Println("Warning: Cache Loader missed tick")
				continue
			}
			targetCacheLoader.isLoading = true
			targetCacheLoader.loadAndNotify()
			targetCacheLoader.isLoading = false
		}
	}()

}

func (targetCacheLoader *TargetCacheLoader) stop() {
	targetCacheLoader.ticker.Stop()
}

//routine to query data
func queryMatch(targetMatcher TargetMatcher, sku string) string {
	return targetMatcher.findMatch(sku)
}

func sendRequests(targetMatcher TargetMatcher, wg *sync.WaitGroup, num int) {
	for i := 0; i < num; i++ {
		wg.Add(1)
		if i%10 == 0 {
			time.Sleep(10 * time.Microsecond)
			fmt.Printf("%d\n", runtime.NumGoroutine())
		}
		go func(i int) {
			//fmt.Printf("Found: %s for %d\n", queryMatch(targetMatcherPool, strconv.Itoa(i)), i)
			queryMatch(targetMatcher, strconv.Itoa(i))
			wg.Done()
		}(i)
		//if i%10000 == 0 {
		//After every 10K submissions
		//	fmt.Printf("%d\n", runtime.NumGoroutine())
		//	}
	}
	wg.Done()
}

func main() {

	var targetMatcher TargetMatcher
	if len(os.Args) > 1 && os.Args[1] == "passive" {
		fmt.Println("Passive Model")
		targetMatcher = NewPassiveTargetMatcher()
	} else {
		//default active
		fmt.Println("Active Model")
		targetMatcher = NewActiveTargetMatcher(10)
	}

	targetCacheLoader := TargetCacheLoader{targetMatcher: targetMatcher}

	var wg sync.WaitGroup
	targetCacheLoader.start()

	tstart := time.Now()

	for s := 0; s < 2; s++ {
		wg.Add(1)
		go sendRequests(targetMatcher, &wg, 500000)
	}

	wg.Wait()
	fmt.Printf("Took: %dms\n", time.Since(tstart).Milliseconds())

	targetCacheLoader.stop()
}
