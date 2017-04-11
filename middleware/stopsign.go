/*
* @Author: wangshuo
* @Date:   2017-04-05 16:32:24
* @Last Modified by:   wangshuo
* @Last Modified time: 2017-04-07 12:04:24
 */

package middleware

import (
	"fmt"
	"sync"
)

type StopSign interface {
	Sign() bool
	Signed() bool
	Reset()
	Deal(code string)
	DealCount(code string) uint32
	DealTotal() uint32
	Summary() string
}

type myStopSign struct {
	signed       bool
	dealCountMap map[string]uint32
	rwmutex      sync.RWMutex
}

func NewStopSign() StopSign {
	sign := &myStopSign{
		dealCountMap: make(map[string]uint32)}
	return sign
}

func (ss *myStopSign) Sign() bool {
	ss.rwmutex.Lock()
	defer ss.rwmutex.Unlock()
	if !ss.signed {
		return false
	}
	ss.signed = true
	return true
}

func (ss *myStopSign) Signed() bool {
	return ss.signed
}

func (ss *myStopSign) Deal(code string) {
	ss.rwmutex.Lock()
	defer ss.rwmutex.Unlock()
	if !ss.signed {
		return
	}
	if _, ok := ss.dealCountMap[code]; !ok {
		ss.dealCountMap[code] = 1
	} else {
		ss.dealCountMap[code] += 1
	}
}

func (ss *myStopSign) Reset() {
	ss.rwmutex.Lock()
	defer ss.rwmutex.Unlock()
	ss.signed = false
	ss.dealCountMap = make(map[string]uint32)
}

func (ss *myStopSign) DealCount(code string) uint32 {
	ss.rwmutex.RLock()
	defer ss.rwmutex.RUnlock()
	return ss.dealCountMap[code]
}

func (ss *myStopSign) DealTotal() uint32 {
	ss.rwmutex.RLock()
	defer ss.rwmutex.RUnlock()
	var total uint32
	for _, n := range ss.dealCountMap {
		total += n
	}
	return total
}

func (ss *myStopSign) Summary() string {
	if ss.signed {
		return fmt.Sprintf("signed: true, dealCount:%v", ss.dealCountMap)
	} else {
		return "signed:false"
	}
}
