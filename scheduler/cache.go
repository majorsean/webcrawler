/*
* @Author: wangshuo
* @Date:   2017-04-13 15:36:57
* @Last Modified by:   wangshuo
* @Last Modified time: 2017-04-14 17:48:42
 */

package scheduler

import (
	"fmt"
	"sync"
	base "webcrawler/base"
)

var statusMap = map[byte]string{
	0: "running",
	1: "closed",
}

type requestCache interface {
	put(req *base.Request) bool
	get() *base.Request
	capacity() int
	length() int
	close()
	summary() string
}

type reqCacheBySlice struct {
	cache  []*base.Request
	mutex  sync.Mutex
	status byte
}

func newRequestCache() requestCache {
	rc := &reqCacheBySlice{
		cache: make([]*base.Request, 0),
	}
	return rc
}

func (rcache *reqCacheBySlice) put(req *base.Request) bool {
	if req == nil {
		return false
	}
	if rcache.status == 1 {
		return false
	}
	rcache.mutex.Lock()
	defer rcache.mutex.Unlock()
	rcache.cache = append(rcache.cache, req)
	return true
}

func (rcache *reqCacheBySlice) get() *base.Request {
	if rcache.length() == 0 {
		return nil
	}
	if rcache.status == 1 {
		return nil
	}
	rcache.mutex.Lock()
	defer rcache.mutex.Unlock()
	req := rcache.cache[0]
	rcache.cache = rcache.cache[1:]
	return req
}

func (rcache *reqCacheBySlice) capacity() int {
	return cap(rcache.cache)
}

func (rcache *reqCacheBySlice) length() int {
	return len(rcache.cache)
}

func (rcache *reqCacheBySlice) close() {
	if rcache.status == 1 {
		return
	}
	rcache.status = 1
}

var summaryTemplate = "status: %s, " + "length:%d, " + "capacity:%d"

func (rcache *reqCacheBySlice) summary() string {
	return fmt.Sprintf(summaryTemplate, statusMap[rcache.status], rcache.length(), rcache.capacity())
}
