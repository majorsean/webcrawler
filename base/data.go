/*
* @Author: wang
* @Date:   2017-04-01 17:58:47
* @Last Modified by:   wang
* @Last Modified time: 2017-04-05 11:29:51
 */

package base

import (
	"net/http"
)

type Request struct {
	httpReq *http.Request
	depth   uint32
}

func NewRequest(httpReq *http.Request, depth uint32) *Request {
	return &Request{httpReq: httpReq, depth: depth}
}

func (req *Request) HttpReq() *http.Request {
	return req.httpReq
}

func (req *Request) Depth() uint32 {
	return req.depth
}

func (req *Request) Valid() bool {
	return req.httpReq != nil && req.httpReq.URL != nil
}

type Response struct {
	httpResp *http.Response
	depth    uint32
}

func NewResponse(httpResp *http.Response, depth uint32) *Response {
	return &Response{httpResp: httpResp, depth: depth}
}

func (resp *Response) HttpResp() *http.Response {
	return resp.httpResp
}

func (resp *Response) Depth() uint32 {
	return resp.depth
}

func (resp *Response) Valid() bool {
	return resp.httpResp != nil && resp.httpResp.Body != nil
}

type Item map[string]interface{}

// mark
func (item Item) Valid() bool {
	return item != nil
}

type Data interface {
	Valid() bool // 数据是否有效
}
