/*
* @Author: wangshuo
* @Date:   2017-04-19 09:49:56
* @Last Modified by:   wangshuo
* @Last Modified time: 2017-04-24 17:08:50
 */

package main

import (
	"errors"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"io"
	"logging"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync/atomic"
	"time"
	"webcrawler/analyzer"
	base "webcrawler/base"
	pipeline "webcrawler/itempipeline"
	sched "webcrawler/scheduler"
	"webcrawler/tool"
)

var (
	logger  logging.Logger = logging.NewSimpleLogger()
	gotPage bool           = false
	count   uint32
)

func genHttpClient() *http.Client {
	return &http.Client{}
}

func main() {
	channelArgs := base.NewChannelArgs(10, 10, 10, 10)
	poolBaseArgs := base.NewPoolBaseArgs(8, 3)
	crawlDepth := uint32(3)
	httpClientGenerator := genHttpClient
	respParsers := getResponseParsers()
	itemProcessors := getItemProcessors()
	startUrl := "https://www.zhihu.com/collection/20615676"
	// startUrl := "https://www.zhihu.com/collection/139296034"
	// startUrl := "https://www.zhihu.com/collection/75387977"
	// startUrl := "https://www.zhihu.com/question/58433345/answer/158035178"
	firstHttpReq, err := http.NewRequest("GET", startUrl, nil)
	if err != nil {
		logger.Errorln(err)
		return
	}

	scheduler := sched.NewScheduler()

	intervalNs := 10 * time.Millisecond
	maxIdleCount := uint(1000)
	checkCountChan := tool.Monitoring(scheduler, intervalNs, maxIdleCount, true, false, record)

	scheduler.Start(channelArgs, poolBaseArgs, crawlDepth, httpClientGenerator, respParsers, itemProcessors, firstHttpReq)

	<-checkCountChan
	fmt.Printf("count:%d\n", count)
}

func record(level byte, content string) {
	if content == "" {
		return
	}
	switch level {
	case 0:
		logger.Infoln(content)
	case 1:
		logger.Warnln(content)
	case 2:
		logger.Fatalln(content)
	}
}

func getResponseParsers() []analyzer.ParseResponse {
	parsers := []analyzer.ParseResponse{
		parseForRequest,
		// parseForAnswer,
	}
	return parsers
}

func parseForAnswer(httpResp *http.Response, respDepth uint32) ([]base.Data, []error) {
	if httpResp.StatusCode != 200 {
		err := errors.New(fmt.Sprintf("Unsupported status code %d. (httpResponse=%v)", httpResp))
		return nil, []error{err}
	}
	var reqUrl *url.URL = httpResp.Request.URL
	if !strings.Contains(reqUrl.String(), "answer") {
		return nil, []error{}
	}
	var httpRespBody io.ReadCloser = httpResp.Body
	defer func() {
		if httpRespBody != nil {
			httpRespBody.Close()
		}
	}()
	dataList := make([]base.Data, 0)
	errs := make([]error, 0)
	doc, err := goquery.NewDocumentFromReader(httpRespBody)
	if err != nil {
		errs = append(errs, err)
		return dataList, errs
	}
	imap := make(map[string]interface{})
	section := doc.Find(".QuestionAnswer-content")
	avatar, _ := section.Find(".AuthorInfo-avatar").Attr("src")
	content, _ := section.Find(".CopyrightRichText-richText").Html()
	imap["nickname"] = section.Find(".UserLink-link").Text()
	imap["authorinfo"] = section.Find(".AuthorInfo-badge").Text()
	imap["voters"] = section.Find(".Voters").Text()
	imap["content"] = content
	imap["avatar"] = avatar
	item := base.Item(imap)
	dataList = append(dataList, &item)
	return dataList, errs
}

func getItemProcessors() []pipeline.ProcessItem {
	itemProcessors := []pipeline.ProcessItem{
		processItem,
	}
	return itemProcessors
}

func parseForRequest(httpResp *http.Response, respDepth uint32) ([]base.Data, []error) {
	if httpResp.StatusCode != 200 {
		err := errors.New(fmt.Sprintf("Unsupported status code %d. (httpResponse=%v)", httpResp))
		return nil, []error{err}
	}

	// var reqUrl *url.URL = httpResp.Request.URL
	var httpRespBody io.ReadCloser = httpResp.Body
	defer func() {
		if httpRespBody != nil {
			httpRespBody.Close()
		}
	}()

	dataList := make([]base.Data, 0)
	errs := make([]error, 0)
	if !gotPage {
		dataList, errs = parseForPage(httpResp, respDepth)
	}
	doc, err := goquery.NewDocumentFromReader(httpRespBody)
	if err != nil {
		errs = append(errs, err)
		return dataList, errs
	}

	doc.Find(".zm-item").Each(func(index int, sel *goquery.Selection) {
		imap := make(map[string]interface{})
		imap["title"] = sel.Find(".zm-item-title").Text()
		name := sel.Find(".name").Text()
		if name != "" {
			imap["nickname"] = name
		} else {
			imap["nickname"] = sel.Find(".author-link").Text()
			imap["authorinfo"] = sel.Find(".bio").Text()
		}
		imap["voters"] = sel.Find(".js-voteCount").Text()
		content, _ := sel.Find(".content").Html()
		imap["content"] = content
		item := base.Item(imap)
		dataList = append(dataList, &item)
	})
	return dataList, errs
}

func parseForPage(httpResp *http.Response, respDepth uint32) ([]base.Data, []error) {
	if gotPage {
		return nil, []error{}
	}
	if httpResp.StatusCode != 200 {
		err := errors.New(fmt.Sprintf("Unsupported status code %d. (httpResponse=%v)", httpResp))
		return nil, []error{err}
	}
	var reqUrl *url.URL = httpResp.Request.URL
	var httpRespBody io.ReadCloser = httpResp.Body
	defer func() {
		if httpRespBody != nil {
			httpRespBody.Close()
		}
	}()

	dataList := make([]base.Data, 0)
	errs := make([]error, 0)
	doc, err := goquery.NewDocumentFromReader(httpRespBody)
	if err != nil {
		errs = append(errs, err)
		return dataList, errs
	}

	doc.Find(".zm-invite-pager").Each(func(index int, sel *goquery.Selection) {
		selSpan := sel.Find("span")
		lastPage := selSpan.Eq(selSpan.Size() - 2).Text()
		lp, err := strconv.Atoi(lastPage)
		if err != nil {
			errs = append(errs, err)
		} else {
			var url string
			for i := 1; i <= lp; i++ {
				url = fmt.Sprintf("%s?page=%d", reqUrl, i)
				httpReq, err := http.NewRequest("GET", url, nil)
				if err != nil {
					errs = append(errs, err)
				} else {
					req := base.NewRequest(httpReq, respDepth)
					dataList = append(dataList, req)
				}
			}
		}
	})
	gotPage = true
	return dataList, errs
}

func processItem(item base.Item) (result base.Item, err error) {
	if item == nil {
		return nil, errors.New("Invalid item!")
	}
	atomic.AddUint32(&count, 1)
	result = make(map[string]interface{})
	time.Sleep(10 * time.Millisecond)
	return result, nil
}
