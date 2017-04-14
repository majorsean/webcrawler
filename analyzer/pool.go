/*
* @Author: wang
* @Date:   2017-04-05 14:45:39
* @Last Modified by:   wangshuo
* @Last Modified time: 2017-04-11 12:18:47
 */

package analyzer

import (
	"errors"
	"fmt"
	"reflect"
	mdw "webcrawler/middleware"
)

type AnalyzerPool interface {
	Take() (Analyzer, error)
	Return(analyzer Analyzer) error
	Total() uint32
	Used() uint32
}

type myAnalyzerPool struct {
	pool  mdw.Pool
	etype reflect.Type
}

type GenAnalyer func() Analyzer

func NewAnalyzerPool(total uint32, gen GenAnalyer) (AnalyzerPool, error) {
	etype := reflect.TypeOf(gen())
	genEntity := func() mdw.Entity {
		return gen()
	}
	pool, err := mdw.NewPool(total, etype, genEntity)
	if err != nil {
		return nil, err
	}
	dlpool := &myAnalyzerPool{pool: pool, etype: etype}
	return dlpool, nil
}

func (ap *myAnalyzerPool) Take() (Analyzer, error) {
	entity, err := ap.pool.Take()
	if err != nil {
		return nil, err
	}
	al, ok := entity.(Analyzer)
	if !ok {
		errMsg := fmt.Sprintf("The type entity is NOT %s!\n", ap.etype)
		panic(errors.New(errMsg))
	}
	return al, nil
}

func (ap *myAnalyzerPool) Return(analyzer Analyzer) error {
	return ap.pool.Return(analyzer)
}

func (ap *myAnalyzerPool) Total() uint32 {
	return ap.pool.Total()
}

func (ap *myAnalyzerPool) Used() uint32 {
	return ap.pool.Used()
}
