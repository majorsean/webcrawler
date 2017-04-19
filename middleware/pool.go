/*
* @Author: wangshuo
* @Date:   2017-04-05 16:21:55
* @Last Modified by:   wangshuo
* @Last Modified time: 2017-04-19 10:59:14
 */

package middleware

import (
	"errors"
	"fmt"
	"reflect"
	"sync"
)

type Entity interface {
	Id() uint32
}

type Pool interface {
	Take() (Entity, error)
	Return(entity Entity) error
	Total() uint32
	Used() uint32
}

type myPool struct {
	total       uint32
	entityType  reflect.Type
	genEntity   func() Entity
	container   chan Entity
	idContainer map[uint32]bool
	mutex       sync.Mutex
}

func NewPool(
	total uint32,
	entityType reflect.Type,
	genEntity func() Entity) (Pool, error) {
	if total == 0 {
		errMsg := fmt.Sprintf("The pool can not be initialized (total=%d)\n", total)
		return nil, errors.New(errMsg)
	}
	size := int(total)
	container := make(chan Entity, size)
	idContainer := make(map[uint32]bool)
	for i := 0; i < size; i++ {
		newEntity := genEntity()
		if entityType != reflect.TypeOf(newEntity) {
			errMsg := fmt.Sprintf("The type of result of function genEntity() is NOT %s\n", entityType)
			return nil, errors.New(errMsg)
		}
		container <- newEntity
		idContainer[newEntity.Id()] = true
	}
	pool := &myPool{total: total, entityType: entityType, container: container, genEntity: genEntity, idContainer: idContainer}
	return pool, nil
}

func (pool *myPool) Take() (Entity, error) {
	entity, ok := <-pool.container
	if !ok {
		return nil, errors.New("The inner container is invalid!")
	}
	pool.mutex.Lock()
	defer pool.mutex.Unlock()
	pool.idContainer[entity.Id()] = false
	return entity, nil
}

func (pool *myPool) Return(entity Entity) error {
	if entity == nil {
		return errors.New("The returning entity invalid!")
	}
	if pool.entityType != reflect.TypeOf(entity) {
		errMsg := fmt.Sprintf("The type of returning entity is NOT %s!\n", pool.entityType)
		return errors.New(errMsg)
	}

	entityId := entity.Id()
	casResult := pool.compareAndSetForIdContainer(entityId, false, true)
	if casResult == 1 {
		pool.container <- entity
		return nil
	} else if casResult == 0 {
		errMsg := fmt.Sprintf("The entity (id=%d) is alreadey in the pool!\n", entityId)
		return errors.New(errMsg)
	} else {
		errMsg := fmt.Sprintf("The entity (id=%d) is illegal!\n", entityId)
		return errors.New(errMsg)
	}
}

func (pool *myPool) compareAndSetForIdContainer(entityId uint32, oldValue bool, newValue bool) int8 {
	pool.mutex.Lock()
	defer pool.mutex.Unlock()
	v, ok := pool.idContainer[entityId]
	if !ok {
		return -1
	}
	if v != oldValue {
		return 0
	}
	pool.idContainer[entityId] = newValue
	return 1
}

func (pool *myPool) Total() uint32 {
	return pool.total
}

func (pool *myPool) Used() uint32 {
	return pool.total - uint32(len(pool.container))
}
