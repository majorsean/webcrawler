/*
* @Author: wangshuo
* @Date:   2017-04-05 15:57:30
* @Last Modified by:   wangshuo
* @Last Modified time: 2017-04-19 09:45:44
 */

package middleware

import (
	"errors"
	"fmt"
	"sync"
	"webcrawler/base"
)

type ChannelManagerStatus uint8

const (
	CHANNEL_MANAGET_STATUS_UNINITIALIZED ChannelManagerStatus = 0
	CHANNEL_MANAGET_STATUS_INITIALIZED   ChannelManagerStatus = 1
	CHANNEL_MANAGET_STATUS_CLOSED        ChannelManagerStatus = 2
)

var statusMap = map[ChannelManagerStatus]string{
	CHANNEL_MANAGET_STATUS_CLOSED:        "closed.",
	CHANNEL_MANAGET_STATUS_INITIALIZED:   "initialized.",
	CHANNEL_MANAGET_STATUS_UNINITIALIZED: "uninitialized.",
}

type ChannelManager interface {
	Init(channelArgs base.ChannelArgs, reset bool) bool
	Close() bool
	ReqChan() (chan base.Request, error)
	RespChan() (chan base.Response, error)
	ItemChan() (chan base.Item, error)
	ErrorChan() (chan error, error)
	Status() ChannelManagerStatus
	Summary() string
}

type myChannelManager struct {
	channalArgs base.ChannelArgs
	reqCh       chan base.Request
	respCh      chan base.Response
	itemCh      chan base.Item
	errorCh     chan error
	status      ChannelManagerStatus
	rwmutex     sync.RWMutex
}

func NewChannelManager(channelArgs base.ChannelArgs) ChannelManager {
	chanm := &myChannelManager{}
	chanm.Init(channelArgs, true)
	return chanm
}

func (chanman *myChannelManager) Init(channalArgs base.ChannelArgs, reset bool) bool {
	if err := channalArgs.Check(); err != nil {
		panic(err)
	}
	chanman.rwmutex.Lock()
	defer chanman.rwmutex.Unlock()
	if chanman.status == CHANNEL_MANAGET_STATUS_INITIALIZED && !reset {
		return false
	}

	chanman.channalArgs = channalArgs
	chanman.reqCh = make(chan base.Request, channalArgs.ReqChanLen())
	chanman.respCh = make(chan base.Response, channalArgs.RespChanLen())
	chanman.itemCh = make(chan base.Item, channalArgs.ItemChanLen())
	chanman.errorCh = make(chan error, channalArgs.ErrorChanLen())
	chanman.status = CHANNEL_MANAGET_STATUS_INITIALIZED

	return true
}

func (chanman *myChannelManager) Close() bool {
	chanman.rwmutex.Lock()
	defer chanman.rwmutex.Unlock()
	if chanman.status != CHANNEL_MANAGET_STATUS_INITIALIZED {
		return false
	}

	close(chanman.reqCh)
	close(chanman.respCh)
	close(chanman.itemCh)
	close(chanman.errorCh)
	chanman.status = CHANNEL_MANAGET_STATUS_CLOSED
	return true
}

func (chanman *myChannelManager) checkStatus() error {
	if chanman.status == CHANNEL_MANAGET_STATUS_INITIALIZED {
		return nil
	}

	statusName, ok := statusMap[chanman.status]
	if !ok {
		statusName = fmt.Sprintf("%d", chanman.status)
	}
	errMsg := fmt.Sprintf("The undesirable status of channel manager: %s\n", statusName)
	return errors.New(errMsg)
}

func (chanman *myChannelManager) ReqChan() (chan base.Request, error) {
	chanman.rwmutex.Lock()
	defer chanman.rwmutex.Unlock()
	if err := chanman.checkStatus(); err != nil {
		return nil, err
	}
	return chanman.reqCh, nil
}

func (chanman *myChannelManager) RespChan() (chan base.Response, error) {
	chanman.rwmutex.Lock()
	defer chanman.rwmutex.Unlock()
	if err := chanman.checkStatus(); err != nil {
		return nil, err
	}
	return chanman.respCh, nil
}

func (chanman *myChannelManager) ItemChan() (chan base.Item, error) {
	chanman.rwmutex.Lock()
	defer chanman.rwmutex.Unlock()
	if err := chanman.checkStatus(); err != nil {
		return nil, err
	}
	return chanman.itemCh, nil
}

func (chanman *myChannelManager) ErrorChan() (chan error, error) {
	chanman.rwmutex.Lock()
	defer chanman.rwmutex.Unlock()
	if err := chanman.checkStatus(); err != nil {
		return nil, err
	}
	return chanman.errorCh, nil
}

var chanmanSummaryTemplate = "status: %s, " +
	"requestChannel: %d/%d, " +
	"responseChannel: %d/%d, " +
	"itemChannel: %d/%d, " +
	"errorChannel: %d/%d"

func (chanman *myChannelManager) Summary() string {
	summary := fmt.Sprintf(chanmanSummaryTemplate, statusMap[chanman.status],
		len(chanman.reqCh), cap(chanman.reqCh),
		len(chanman.respCh), cap(chanman.respCh),
		len(chanman.itemCh), cap(chanman.itemCh),
		len(chanman.errorCh), cap(chanman.errorCh))
	return summary
}

func (chanman *myChannelManager) Status() ChannelManagerStatus {
	return chanman.status
}
