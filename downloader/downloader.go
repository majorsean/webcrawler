/*
* @Author: wang
* @Date:   2017-04-05 11:53:24
* @Last Modified by:   wang
* @Last Modified time: 2017-04-05 11:59:59
 */

package downloader

import (
	"webcrawler/base"
	mdw "webcrawler/middleware"
)

var downloaderIdGenerator mdw.IdGenerator = mdw.NewIdGenerator()

type PageDownloader interface {
	Id() uint32
	Download(req base.Request) (*base.Response, error)
}

func Download() {

}
