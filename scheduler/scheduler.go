/*
* @Author: wang
* @Date:   2017-04-05 15:07:42
* @Last Modified by:   wangshuo
* @Last Modified time: 2017-04-19 13:48:46
 */

package scheduler

import (
	"errors"
	"fmt"
	"logging"
	"net/http"
	"strings"
	"sync/atomic"
	"time"
	"webcrawler/analyzer"
	anlz "webcrawler/analyzer"
	base "webcrawler/base"
	dl "webcrawler/downloader"
	ipl "webcrawler/itempipeline"
	mdw "webcrawler/middleware"
)

const (
	DOWNLOADER_CODE   = "downloader"
	ANALYZER_CODE     = "analyzer"
	ITEMPIPELINE_CODE = "item_pipeline"
	SCHEDULER_CODE    = "scheduler"
)

var logger logging.Logger = logging.NewSimpleLogger()

type GenhttpClient func() *http.Client

type Scheduler interface {
	Start(channelArgs base.ChannelArgs,
		poolBaseArgs base.PoolBaseArgs,
		crawlDepth uint32,
		httpClientGenerator GenhttpClient,
		respParsers []analyzer.ParseResponse,
		itemProcessors []ipl.ProcessItem,
		firstHttpReq *http.Request) (err error)
	Stop() bool
	Running() bool
	ErrorChan() <-chan error
	Idle() bool
	Summary(prefix string) SchedSummary
}

type myScheduler struct {
	channelArgs   base.ChannelArgs
	poolBaseArgs  base.PoolBaseArgs
	crawlDepth    uint32
	primaryDomain string
	chanman       mdw.ChannelManager
	stopSign      mdw.StopSign
	dlpool        dl.PageDownloaderPool
	analyzerPool  anlz.AnalyzerPool
	itemPipeline  ipl.ItemPipeline
	running       uint32
	reqCache      requestCache
	urlMap        map[string]bool
}

func NewScheduler() Scheduler {
	return &myScheduler{}
}

func (sched *myScheduler) Start(channelArgs base.ChannelArgs,
	poolBaseArgs base.PoolBaseArgs,
	crawlDepth uint32,
	httpClientGenerator GenhttpClient,
	respParsers []analyzer.ParseResponse,
	itemProcessors []ipl.ProcessItem,
	firstHttpReq *http.Request) (err error) {

	defer func() {
		if p := recover(); p != nil {
			errMsg := fmt.Sprintf("Fatal Scheduler Error: %s\n", p)
			logger.Fatal(errMsg)
			err = errors.New(errMsg)
		}
	}()
	if atomic.LoadUint32(&sched.running) == 1 {
		return errors.New("The scheduler has been started!\n")
	}
	atomic.StoreUint32(&sched.running, 1)

	if err := channelArgs.Check(); err != nil {
		return err
	}
	sched.channelArgs = channelArgs

	if err := poolBaseArgs.Check(); err != nil {
		return err
	}
	sched.poolBaseArgs = poolBaseArgs
	sched.crawlDepth = crawlDepth

	sched.chanman = generateChannelManager(sched.channelArgs)
	if httpClientGenerator == nil {
		return errors.New("The Http Client generator list is invalid!\n")
	}

	dlpool, err := generatePageDownloaderPool(sched.poolBaseArgs.PageDownloaderPoolSize(), httpClientGenerator)
	if err != nil {
		errMsg := fmt.Sprintf("Occur error when get page downloader pool: %s\n", err)
		return errors.New(errMsg)
	}
	sched.dlpool = dlpool

	analyzerPool, err := generateAnalyzerPool(sched.poolBaseArgs.AnalyzerPoolSize())
	if err != nil {
		errMsg := fmt.Sprintf("Occur error when get analyzer pool: %s\n", err)
		return errors.New(errMsg)
	}
	sched.analyzerPool = analyzerPool

	if itemProcessors == nil {
		return errors.New("The item processor list is invalid!\n")
	}
	for i, item := range itemProcessors {
		if item == nil {
			return errors.New(fmt.Sprintf("The %dth item processor is invalid!\n", i))
		}
	}
	sched.itemPipeline = generateItemProcessors(itemProcessors)

	if sched.stopSign == nil {
		sched.stopSign = mdw.NewStopSign()
	} else {
		sched.stopSign.Reset()
	}

	sched.reqCache = newRequestCache()
	sched.urlMap = make(map[string]bool)

	sched.startDownloading()
	sched.activateAnalyzers(respParsers)
	sched.openItemPipeline()
	sched.schedule(10 * time.Millisecond)

	if firstHttpReq == nil {
		return errors.New("The first Http request is invalid!\n")
	}
	pd, err := getPrimaryDomain(firstHttpReq.Host)
	if err != nil {
		return err
	}
	sched.primaryDomain = pd
	firstReq := base.NewRequest(firstHttpReq, 0)
	sched.reqCache.put(firstReq)
	return nil
}

func (sched *myScheduler) schedule(interval time.Duration) {
	go func() {
		for {
			if sched.stopSign.Signed() {
				sched.stopSign.Deal(SCHEDULER_CODE)
				return
			}
			remainder := cap(sched.getReqChan()) - len(sched.getReqChan())
			var temp *base.Request
			for remainder > 0 {
				temp = sched.reqCache.get()
				if temp == nil {
					break
				}
				if sched.stopSign.Signed() {
					sched.stopSign.Deal(SCHEDULER_CODE)
					return
				}
				sched.getReqChan() <- *temp
				remainder--
			}
			time.Sleep(interval)
		}
	}()
}

func (sched *myScheduler) openItemPipeline() {
	go func() {
		sched.itemPipeline.SetFailFast(true)
		code := ITEMPIPELINE_CODE
		for item := range sched.getItemChan() {
			go func(item base.Item) {
				defer func() {
					if p := recover(); p != nil {
						errMsg := fmt.Sprintf("Fatal Item Processing Error:%s", p)
						logger.Fatal(errMsg)
					}
				}()
				errs := sched.itemPipeline.Send(item)
				if errs != nil {
					for _, err := range errs {
						sched.sendError(err, code)
					}
				}
			}(item)
		}
	}()
}

func (sched *myScheduler) activateAnalyzers(respParsers []analyzer.ParseResponse) {
	go func() {
		for {
			resp, ok := <-sched.getRespChan()
			if !ok {
				break
			}
			go sched.analyze(respParsers, resp)
		}
	}()
}

func (sched *myScheduler) analyze(respParsers []analyzer.ParseResponse, resp base.Response) {
	defer func() {
		if p := recover(); p != nil {
			errMsg := fmt.Sprintf("Fatal Analysis Error: %s\n", p)
			logger.Fatal(errMsg)
		}
	}()
	analyzer, err := sched.analyzerPool.Take()
	if err != nil {
		errMsg := fmt.Sprintf("Analyzer pool error:%s\n", err)
		sched.sendError(errors.New(errMsg), SCHEDULER_CODE)
	}
	defer func() {
		err := sched.analyzerPool.Return(analyzer)
		if err != nil {
			errMsg := fmt.Sprintf("Analyzer pool error:%s\n", err)
			sched.sendError(errors.New(errMsg), SCHEDULER_CODE)
		}
	}()

	code := generateCode(SCHEDULER_CODE, analyzer.Id())
	dataList, errs := analyzer.Analyze(respParsers, resp)
	if dataList != nil {
		for _, data := range dataList {
			if data == nil {
				continue
			}
			switch d := data.(type) {
			case *base.Request:
				sched.saveReqToCache(*d, code)
			case *base.Item:
				sched.sendItem(*d, code)
			default:
				errMsg := fmt.Sprintf("Unsupported data type '%T'! (value=%v)\n", d, d)
				sched.sendError(errors.New(errMsg), code)
			}
		}
	}
	if errs != nil {
		for _, err := range errs {
			sched.sendError(err, code)
		}
	}
}

func (sched *myScheduler) saveReqToCache(req base.Request, code string) bool {
	httpReq := req.HttpReq()
	if httpReq == nil {
		logger.Warnln("Ignore the request! It's HTTP request is invalid!")
		return false
	}

	reqUrl := httpReq.URL
	if reqUrl == nil {
		logger.Warnln("Ignore the request! It's url is invalid!")
		return false
	}

	if strings.ToLower(reqUrl.Scheme) != "http" {
		logger.Warnf("Ignore the request! It's url scheme '%s', but should be 'http'!\n", reqUrl.Scheme)
		return false
	}

	if _, ok := sched.urlMap[reqUrl.String()]; ok {
		logger.Warnf("Ignore the request! It's url is repeated. (requestUrl=%s)\n", reqUrl)
		return false
	}

	if pd, _ := getPrimaryDomain(reqUrl.Host); pd != sched.primaryDomain {
		logger.Warnf("Ignore the request! It's host '%s' not in primary domain '%s' . (reuqestUrl=%s)\n", httpReq.Host, sched.primaryDomain, reqUrl)
		return false
	}

	if req.Depth() > sched.crawlDepth {
		logger.Warnf("Igone the request! It's depth %d depth greater than %d. (requestUrl=%s)\n", req.Depth(), sched.crawlDepth, reqUrl)
		return false
	}

	if sched.stopSign.Signed() {
		sched.stopSign.Deal(code)
	}
	sched.reqCache.put(&req)
	sched.urlMap[reqUrl.String()] = true
	return true

}

func (sched *myScheduler) sendItem(item base.Item, code string) bool {
	if sched.stopSign.Signed() {
		sched.stopSign.Deal(code)
		return false
	}
	sched.getItemChan() <- item
	return true
}

func (sched *myScheduler) getItemChan() chan base.Item {
	itemChan, err := sched.chanman.ItemChan()
	if err != nil {
		panic(err)
	}
	return itemChan
}

func (sched *myScheduler) startDownloading() {
	go func() {
		for {
			req, ok := <-sched.getReqChan()
			if !ok {
				break
			}
			go sched.download(req)
		}
	}()
}

func (sched *myScheduler) getReqChan() chan base.Request {
	reqChan, err := sched.chanman.ReqChan()
	if err != nil {
		panic(err)
	}
	return reqChan
}

func (sched *myScheduler) getErrChan() chan error {
	errChan, err := sched.chanman.ErrorChan()
	if err != nil {
		panic(err)
	}
	return errChan
}

func (sched *myScheduler) download(req base.Request) {
	downloader, err := sched.dlpool.Take()
	if err != nil {
		errMsg := fmt.Sprintf("download pool error:%s\n", err)
		sched.sendError(errors.New(errMsg), SCHEDULER_CODE)
	}
	defer func() {
		err := sched.dlpool.Return(downloader)
		if err != nil {
			errMsg := fmt.Sprintf("download pool error:%s\n", err)
			sched.sendError(errors.New(errMsg), SCHEDULER_CODE)
		}
	}()

	code := generateCode(DOWNLOADER_CODE, downloader.Id())
	respp, err := downloader.Download(req)
	if respp != nil {
		sched.sendResp(*respp, code)
	}
	if err != nil {
		sched.sendError(err, code)
	}

	defer func() {
		if p := recover(); p != nil {
			errMsg := fmt.Sprintf("Fatal Download Error:%s\n", p)
			logger.Fatal(errMsg)
		}
	}()
}

func (sched *myScheduler) sendResp(resp base.Response, code string) bool {
	if sched.stopSign.Signed() {
		sched.stopSign.Deal(code)
		return false
	}
	sched.getRespChan() <- resp
	return true
}

func (sched *myScheduler) getRespChan() chan base.Response {
	respChan, err := sched.chanman.RespChan()
	if err != nil {
		panic(err)
	}
	return respChan
}

func (sched *myScheduler) sendError(err error, code string) bool {
	if err == nil {
		return false
	}
	codePrefix := parseCode(code)[0]
	var errType base.ErrorType
	switch codePrefix {
	case DOWNLOADER_CODE:
		errType = base.DOWNLOADER_ERROR
	case ANALYZER_CODE:
		errType = base.ANALYZER_ERROR
	case ITEMPIPELINE_CODE:
		errType = base.ITEM_PROCCESSOR_ERROR
	}
	cError := base.NewCrawlerError(errType, err.Error())
	if sched.stopSign.Signed() {
		sched.stopSign.Deal(code)
		return false
	}
	go func() {
		sched.getErrChan() <- cError
	}()
	return true
}

func (sched *myScheduler) Stop() bool {
	if atomic.LoadUint32(&sched.running) != 1 {
		return false
	}
	sched.stopSign.Sign()
	sched.reqCache.close()
	sched.chanman.Close()
	atomic.StoreUint32(&sched.running, 2)
	return true
}

func (sched *myScheduler) Running() bool {
	return atomic.LoadUint32(&sched.running) == 1
}

func (sched *myScheduler) ErrorChan() <-chan error {
	if sched.chanman.Status() != mdw.CHANNEL_MANAGET_STATUS_INITIALIZED {
		return nil
	}
	return sched.getErrChan()
}

func (sched *myScheduler) Idle() bool {
	idleDlPool := sched.dlpool.Used() == 0
	idleAnalyzerPool := sched.analyzerPool.Used() == 0
	idleItemPipeline := sched.itemPipeline.ProccessingNumber() == 0
	if idleDlPool && idleAnalyzerPool && idleItemPipeline {
		return true
	}
	return false
}

func (sched *myScheduler) Summary(prefix string) SchedSummary {
	return NewSchedSummary(sched, prefix)
}
