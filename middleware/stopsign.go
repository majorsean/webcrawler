/*
* @Author: wangshuo
* @Date:   2017-04-05 16:32:24
* @Last Modified by:   wangshuo
* @Last Modified time: 2017-04-05 16:33:37
 */

package middleware

type StopSign interface {
	Sign() bool
	Signed() bool
	Reset()
	Deal(code string)
	DealCount(code string) uint32
	DealTotal() uint32
	Summary() string
}
