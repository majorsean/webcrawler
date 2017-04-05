/*
* @Author: wang
* @Date:   2017-04-05 11:08:35
* @Last Modified by:   wang
* @Last Modified time: 2017-04-05 11:28:37
 */

package base

import (
	"bytes"
	"fmt"
)

type ErrorType string

const (
	DOWNLOADER_ERROR      ErrorType = "Dowloader Error"
	ANALYZER_ERROR        ErrorType = "Analyzer Error"
	ITEM_PROCCESSOR_ERROR ErrorType = "Item Proccessor Error"
)

type CrawlerError interface {
	Type() ErrorType
	Error() string
}

type myCrawlerError struct {
	errType    ErrorType
	errMsg     string
	fullErrMsg string
}

func NewCrawlerError(errType ErrorType, errMsg string) CrawlerError {
	return &myCrawlerError{errType: errType, errMsg: errMsg}
}

func (ce *myCrawlerError) Error() string {
	if ce.fullErrMsg == "" {
		ce.genFullErrMsg()
	}
	return ce.fullErrMsg
}

func (ce *myCrawlerError) Type() ErrorType {
	return ce.errType
}

func (ce *myCrawlerError) genFullErrMsg() {
	var buffer bytes.Buffer
	buffer.WriteString("Crawler Error: ")
	if ce.errType != "" {
		buffer.WriteString(string(ce.errType))
		buffer.WriteString(": ")
	}
	buffer.WriteString(ce.errMsg)
	ce.fullErrMsg = fmt.Sprintf("%s\n", buffer.String())
	return
}
