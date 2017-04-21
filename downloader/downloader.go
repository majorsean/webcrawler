/*
* @Author: wang
* @Date:   2017-04-05 11:53:24
* @Last Modified by:   wangshuo
* @Last Modified time: 2017-04-21 11:22:18
 */

package downloader

import (
	"net/http"
	"webcrawler/base"
	mdw "webcrawler/middleware"
)

var downloaderIdGenerator mdw.IdGenerator = mdw.NewIdGenerator()

type PageDownloader interface {
	Id() uint32
	Download(req base.Request) (*base.Response, error)
}

type myPageDownloader struct {
	id         uint32
	httpClient http.Client
}

func genDownloaderId() uint32 {
	return downloaderIdGenerator.GetUint32()
}

func NewPageDownloader(client *http.Client) PageDownloader {
	id := genDownloaderId()
	if client == nil {
		client = &http.Client{}
	}
	return &myPageDownloader{id: id, httpClient: *client}
}

func (dl *myPageDownloader) Id() uint32 {
	return dl.id
}

func (dl *myPageDownloader) Download(req base.Request) (*base.Response, error) {
	httpReq := req.HttpReq()
	httpResp, err := dl.httpClient.Do(httpReq)
	if err != nil {
		return nil, err
	}
	return base.NewResponse(httpResp, req.Depth()), nil
}
