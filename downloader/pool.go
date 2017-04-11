/*
* @Author: wang
* @Date:   2017-04-05 14:10:51
* @Last Modified by:   wangshuo
* @Last Modified time: 2017-04-11 10:21:46
 */

package downloader

import (
	"reflect"
	"fmt"
	"errors"
	mdw "webcrawler/middleware"
)

type GenPageDownloader() func() PageDownloader

type PageDownloaderPool interface {
	Take() (PageDownloader, error)
	Return(dl PageDownloader) error
	Total() uint32
	Used() uint32
}

type myDownloaderPool struct {
	pool  mdw.Pool
	etype reflect.Type
}

func NewPageDownloaderPool(total uint32,gen GenPageDownloader) (PageDownloaderPool,error) {
	etype := reflect.TypeOf(gen())
	genEntity := func() mdw.Entity{
		return gen()
	}
	pool,err := mdw.NewPool(total, etype, genEntity)
	if err != nil {
		return nil,err
	}
	dlpool := &myDownloaderPool{pool:pool,etype:etype}
	return dlpool,nil
}

func (dp *myDownloaderPool) Take() (PageDownloader, error) {
	entity ,err := dp.pool.Take()
	if err != nil {
		return nil,err
	}
	dl,ok := entity.(PageDownloader)
	if !ok {
		errMsg := fmt.Sprintf("The type of entity is NOT %s!\n", dp.etype)
		panic(errors.New(errMsg))
	}
	return dl,nil
}

func (dp *myDownloaderPool) Return(dl PageDownloader) error {
	return dp.pool.Return(dl)
}

func (dp *myDownloaderPool) Total() uint32 {
	return dp.pool.Total()
}

func (dp *myDownloaderPool) Used() uint32 {
	return dp.pool.Used()
}
