/*
* @Author: wang
* @Date:   2017-04-05 14:10:51
* @Last Modified by:   wang
* @Last Modified time: 2017-04-05 14:23:40
 */

package downloader

type PageDownloader interface {
	Take() (PageDownloader, error)
	Return(dl PageDownloader) error
	Total() uint32
	Used() uint32
}
