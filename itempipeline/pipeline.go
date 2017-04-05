/*
* @Author: wang
* @Date:   2017-04-05 14:51:08
* @Last Modified by:   wang
* @Last Modified time: 2017-04-05 15:06:14
 */

package itemproc

import (
	"webcrawler/base"
)

type ItemPipeline interface {
	Send(item base.Item) []error
	FailFast() bool
	SetFailFast(failFast bool)
	Count() []uint64
	ProccessingNumber() uint64
	Summary() string
}
