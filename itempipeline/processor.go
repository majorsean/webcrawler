/*
* @Author: wang
* @Date:   2017-04-05 15:04:25
* @Last Modified by:   wang
* @Last Modified time: 2017-04-05 15:06:16
 */

package itemproc

import (
	"webcrawler/base"
)

type ProcessItem func(item base.Item) (result base.Item, err error)
