/*
* @Author: wang
* @Date:   2017-04-05 14:31:42
* @Last Modified by:   wang
* @Last Modified time: 2017-04-05 14:34:24
 */

package analyzer

import (
	"net/http"
	"webcrawler/base"
)

type ParseResponse func(httpResp *http.Response, respDepth uint32) ([]base.Data, []error)

type Analyzer interface {
	Id() uint32
	Analyze(respParsers []ParseResponse, resp base.Response) ([]base.Data, []error)
}
