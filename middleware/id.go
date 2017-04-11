/*
* @Author: wang
* @Date:   2017-04-05 11:57:32
* @Last Modified by:   wangshuo
* @Last Modified time: 2017-04-07 14:46:10
 */

package middleware

import (
	"math"
	"sync"
)

type IdGenerator interface {
	GetUint32() uint32
}

type myIdGenerator struct {
	sn    uint32
	ended bool
	lock  sync.Mutex
}

func NewIdGenerator() IdGenerator {
	return &myIdGenerator{}
}

func (gen *myIdGenerator) GetUint32() uint32 {
	gen.lock.Lock()
	defer gen.lock.Unlock()
	if gen.ended {
		defer func() { gen.ended = false }()
		gen.sn = 0
		return gen.sn
	}
	id := gen.sn
	if id < math.MaxUint32 {
		gen.sn++
	} else {
		gen.ended = true
	}
	return id
}
