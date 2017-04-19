/*
* @Author: wangshuo
* @Date:   2017-04-19 09:49:56
* @Last Modified by:   wangshuo
* @Last Modified time: 2017-04-19 13:50:56
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
	"strings"
	"time"
	"webcrawler/analyzer"
	base "webcrawler/base"
	pipeline "webcrawler/itempipeline"
	sched "webcrawler/scheduler"
	"webcrawler/tool"
)

var logger logging.Logger = logging.NewSimpleLogger()

func genHttpClient() *http.Client {
	return &http.Client{}
}

func main() {
	channelArgs := base.NewChannelArgs(10, 10, 10, 10)
	poolBaseArgs := base.NewPoolBaseArgs(3, 3)
	crawlDepth := uint32(1)
	httpClientGenerator := genHttpClient
	respParsers := getResponseParsers()
	itemProcessors := getItemProcessors()
	startUrl := "http://www.qq.com"
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
		parseForATag,
	}
	return parsers
}

func getItemProcessors() []pipeline.ProcessItem {
	itemProcessors := []pipeline.ProcessItem{
		processItem,
	}
	return itemProcessors
}

func processItem(item base.Item) (result base.Item, err error) {
	if item == nil {
		return nil, errors.New("Invalid item!")
	}

	result = make(map[string]interface{})
	for k, v := range item {
		result[k] = v
	}
	if _, ok := result["number"]; !ok {
		result["number"] = len(result)
	}
	time.Sleep(10 * time.Millisecond)
	return result, nil
}

func parseForATag(httpResp *http.Response, respDepth uint32) ([]base.Data, []error) {
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

	doc.Find("a").Each(func(index int, sel *goquery.Selection) {
		href, exists := sel.Attr("href")
		if !exists || href == "" || href == "#" || href == "/" {
			return
		}
		href = strings.TrimSpace(href)
		lowerHref := strings.ToLower(href)
		if href != "" && !strings.HasPrefix(lowerHref, "javascript") {
			aUrl, err := url.Parse(href)
			if err != nil {
				errs = append(errs, err)
				return
			}
			if !aUrl.IsAbs() {
				aUrl = reqUrl.ResolveReference(aUrl)
			}
			httpReq, err := http.NewRequest("GET", aUrl.String(), nil)
			if err != nil {
				errs = append(errs, err)
			} else {
				req := base.NewRequest(httpReq, respDepth)
				dataList = append(dataList, req)
			}
		}
		text := strings.TrimSpace(sel.Text())
		if text != "" {
			imap := make(map[string]interface{})
			imap["parent_url"] = reqUrl
			imap["a.text"] = text
			imap["a.index"] = index
			item := base.Item(imap)
			dataList = append(dataList, &item)
		}
	})
	return dataList, errs
}
