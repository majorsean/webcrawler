/*
* @Author: wang
* @Date:   2017-04-05 14:31:42
* @Last Modified by:   wangshuo
* @Last Modified time: 2017-04-11 12:00:18
 */

package analyzer

import (
	"errors"
	"fmt"
	"logging"
	"net/http"
	"net/url"
	"webcrawler/base"
	mdw "webcrawler/middleware"
)

var analyzerIdGenerator mdw.IdGenerator = mdw.NewIdGenerator()

var logger logging.Logger = base.NewLogger()

type ParseResponse func(httpResp *http.Response, respDepth uint32) ([]base.Data, []error)

func genAnalyzerId() uint32 {
	return analyzerIdGenerator.GetUint32()
}

type Analyzer interface {
	Id() uint32
	Analyze(respParsers []ParseResponse, resp base.Response) ([]base.Data, []error)
}

type myAnalyzer struct {
	id uint32
}

func NewAnalyzer() Analyzer {
	return &myAnalyzer{id: genAnalyzerId()}
}

func (analyzer *myAnalyzer) Id() uint32 {
	return analyzer.id
}

func (analyzer *myAnalyzer) Analyze(respParsers []ParseResponse, resp base.Response) (dataList []base.Data, errorList []error) {
	if respParsers == nil {
		err := errors.New("The response list is invalid!")
		return nil, []error{err}
	}
	httpResp := resp.HttpResp()
	if httpResp == nil {
		err := errors.New("The http response is invalid!")
		return nil, []error{err}
	}

	var reqUrl *url.URL = httpResp.Request.URL

	logger.Infof("Parse the response (reqUrl=%s)...\n", reqUrl)

	respDepth := resp.Depth()

	dataList = make([]base.Data, 0)
	errorList = make([]error, 0)

	for i, respParser := range respParsers {
		if respParser == nil {
			err := errors.New(fmt.Sprintf("The document parser [%d] is invalid!\n", i))
			errorList = append(errorList, err)
			continue
		}
		pDataList, pErrorList := respParser(httpResp, respDepth)
		if pDataList != nil {
			for _, pData := range pDataList {
				dataList = appendDataList(dataList, pData, respDepth)
			}
		}
		if pErrorList != nil {
			for _, err := range errorList {
				errorList = appendErrorList(errorList, err)
			}
		}
	}
	return
}

func appendDataList(dataList []base.Data, data base.Data, respDepth uint32) []base.Data {
	if data == nil {
		return dataList
	}
	req, ok := data.(*base.Request)
	if !ok {
		return append(dataList, data)
	}

	newDepth := respDepth + 1
	if req.Depth() != newDepth {
		req = base.NewRequest(req.HttpReq(), newDepth)
	}
	return append(dataList, req)
}

func appendErrorList(errorList []error, err error) []error {
	if err == nil {
		return errorList
	}
	return append(errorList, err)
}
