/*
* @Author: wangshuo
* @Date:   2017-04-05 16:21:55
* @Last Modified by:   wangshuo
* @Last Modified time: 2017-04-05 16:25:04
 */

package middleware

type Entity interface {
	Id() uint32
}

type Pool interface {
	Take() Entity
	Return(entity Entity) error
	Total() uint32
	Used() uint32
}
