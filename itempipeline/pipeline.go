/*
* @Author: wang
* @Date:   2017-04-05 14:51:08
* @Last Modified by:   wangshuo
* @Last Modified time: 2017-04-21 15:20:33
 */

package itemproc

import (
	"errors"
	"fmt"
	"sync/atomic"
	"webcrawler/base"
)

type ItemPipeline interface {
	Send(item base.Item) []error
	FailFast() bool
	SetFailFast(failFast bool)
	Count() []uint64
	ProccessingNumber() uint64
	Summary() string
}

type myItemPipeline struct {
	itemProcessors   []ProcessItem
	failFast         bool
	sent             uint64
	accepted         uint64
	processed        uint64
	processingNumber uint64
}

func NewItemPipeline(itemProcessors []ProcessItem) ItemPipeline {
	if itemProcessors == nil {
		panic(errors.New(fmt.Sprintln("Invalid item process list!")))
	}
	innerProcessors := make([]ProcessItem, 0)
	for i, ip := range itemProcessors {
		if ip == nil {
			panic(errors.New(fmt.Sprintf("Invalid item process [%d]!\n", i)))
		}
		innerProcessors = append(innerProcessors, ip)
	}
	return &myItemPipeline{itemProcessors: innerProcessors}
}

func (ip *myItemPipeline) Send(item base.Item) []error {
	atomic.AddUint64(&ip.processingNumber, 1)
	defer atomic.AddUint64(&ip.processingNumber, ^uint64(0))
	atomic.AddUint64(&ip.sent, 1)
	errs := make([]error, 0)
	if item == nil {
		errs = append(errs, errors.New("The item is invalid!"))
		return errs
	}
	atomic.AddUint64(&ip.accepted, 1)

	var currentItem base.Item = item
	for _, itemProcessor := range ip.itemProcessors {
		processedItem, err := itemProcessor(currentItem)
		if err != nil {
			errs = append(errs, err)
			if ip.failFast {
				break
			}
		}
		if processedItem != nil {
			currentItem = processedItem
		}
	}
	atomic.AddUint64(&ip.processed, 1)
	return errs
}

func (ip *myItemPipeline) FailFast() bool {
	return ip.failFast
}

func (ip *myItemPipeline) SetFailFast(failFast bool) {
	ip.failFast = failFast
}

func (ip *myItemPipeline) Count() []uint64 {
	count := make([]uint64, 3)
	count[0] = atomic.LoadUint64(&ip.sent)
	count[1] = atomic.LoadUint64(&ip.accepted)
	count[2] = atomic.LoadUint64(&ip.processed)
	return count
}

func (ip *myItemPipeline) ProccessingNumber() uint64 {
	return atomic.LoadUint64(&ip.processingNumber)
}

var summaryTemplate = "falFast: %v, processorNumber: %d, sent: %d, accepted: %d, processed: %d, processingNumer: %d"

func (ip *myItemPipeline) Summary() string {
	counts := ip.Count()
	summary := fmt.Sprintf(summaryTemplate, ip.failFast, len(ip.itemProcessors), counts[0], counts[1], counts[2], ip.ProccessingNumber())
	return summary
}
