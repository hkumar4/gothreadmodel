package main

import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

/**----------------------------------------------------
 				Active Target Matcher
*-----------------------------------------------------------*/

type QueryMessage struct {
	sku      string
	wg       sync.WaitGroup
	strategy string //reply
}

type TargetMatcherWorker struct {
	queryChan       chan *QueryMessage
	updateCacheChan chan *SkuCache
	storeroomChan   chan *QueryMessage
	skuCache        *SkuCache
}

//Factory method for constructing a TargetMatcherWorker
func NewTargetMatcherWorker() *TargetMatcherWorker {
	targetMatcherWorker := new(TargetMatcherWorker)
	targetMatcherWorker.queryChan = make(chan *QueryMessage, 1000)
	targetMatcherWorker.storeroomChan = make(chan *QueryMessage, 1000)
	targetMatcherWorker.updateCacheChan = make(chan *SkuCache)
	return targetMatcherWorker
}

func (targetMatcherWorker *TargetMatcherWorker) run() {
	//runtime.LockOSThread() //<--- Not Good
	for {
		select {
		case queryMessage := <-targetMatcherWorker.queryChan:
			if targetMatcherWorker.skuCache != nil {
				targetMatcherWorker.skuCache.match(queryMessage.sku) //Use CPU
				//Simulate Async IO
				go func() {
					time.Sleep(50 * time.Microsecond)
					targetMatcherWorker.storeroomChan <- queryMessage
				}()
			} else {
				//Should never happen
				queryMessage.wg.Done()
			}
		case queryMessage := <-targetMatcherWorker.storeroomChan:
			queryMessage.strategy = targetMatcherWorker.skuCache.match(queryMessage.sku) //Use CPU
			queryMessage.wg.Done()
		case skuCache := <-targetMatcherWorker.updateCacheChan:
			fmt.Println("Updated skuCache")
			targetMatcherWorker.skuCache = skuCache
		}
	}
}

func (targetMatcherWorker *TargetMatcherWorker) doQuery(sku string) string {
	queryMessage := QueryMessage{sku: sku}
	queryMessage.wg.Add(1)
	targetMatcherWorker.queryChan <- &queryMessage //Send a query to the worker goroutine
	queryMessage.wg.Wait()                         //wait for the result
	return queryMessage.strategy
}

type ActiveTargetMatcher struct {
	pool                    []*TargetMatcherWorker
	nextTargetMatcherWorker int64
}

//Factory to create TargetMatcher
func NewActiveTargetMatcher(numTargetMatcherWorkers int) *ActiveTargetMatcher {
	activeTargetMatcher := new(ActiveTargetMatcher)
	activeTargetMatcher.pool = make([]*TargetMatcherWorker, numTargetMatcherWorkers)

	for i := 0; i < numTargetMatcherWorkers; i++ {
		targetMatcherWorker := NewTargetMatcherWorker()
		activeTargetMatcher.pool[i] = targetMatcherWorker
		go targetMatcherWorker.run()
	}

	return activeTargetMatcher
}

func (activeTargetMatcher *ActiveTargetMatcher) getNextTargetMatcherWorker() *TargetMatcherWorker {
	//Do round-robin
	var nextIndex = int(atomic.AddInt64(&activeTargetMatcher.nextTargetMatcherWorker, 1))
	return activeTargetMatcher.pool[nextIndex%len(activeTargetMatcher.pool)]
}

func (activeTargetMatcher *ActiveTargetMatcher) updateCache(skuCache *SkuCache) {
	for i := range activeTargetMatcher.pool {
		activeTargetMatcher.pool[i].updateCacheChan <- skuCache
	}
}

func (activeTargetMatcher *ActiveTargetMatcher) findMatch(sku string) string {
	targetMatcherWorker := activeTargetMatcher.getNextTargetMatcherWorker()
	return targetMatcherWorker.doQuery(sku)
}
