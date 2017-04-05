/*
* @Author: wang
* @Date:   2017-04-05 15:07:42
* @Last Modified by:   wang
* @Last Modified time: 2017-04-05 15:32:11
 */

package scheduler

import (
	"net/http"
	"webcrawler/analyzer"
	"webcrawler/base"
	itemproc "webcrawler/itempipeline"
)

type GenhttpClient func() *http.Client

type Scheduler interface {
	Start(channelLen uint,
		poolSize uint32,
		crawlDepth uint32,
		httpClientGenerator GenhttpClient,
		respParsers []analyzer.ParseResponse,
		itemProcessors []itemproc.ProcessItem,
		firstHttpReq *http.Request) (err error)
	Stop() bool
	Running() bool
	ErrorChan() <-chan error
	Idle() bool
	Summary(prefix string) SchedSummary
}
