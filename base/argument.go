/*
* @Author: wangshuo
* @Date:   2017-04-17 11:43:14
* @Last Modified by:   wangshuo
* @Last Modified time: 2017-04-18 14:54:58
 */

package base

import (
	"errors"
	"fmt"
)

type Args interface {
	Check() error
	String() string
}

type ChannelArgs struct {
	reqChanLen   uint
	respChanLen  uint
	itemChanLen  uint
	errorChanLen uint
	description  string
}

type PoolBaseArgs struct {
	pageDownloaderPoolSize uint32
	analyzerPoolSize       uint32
	description            string
}

func NewChannelArgs(reqChanLen uint, respChanLen uint, itemChanLen uint, errorChanLen uint) ChannelArgs {
	return ChannelArgs{reqChanLen: reqChanLen, respChanLen: respChanLen, itemChanLen: itemChanLen, errorChanLen: errorChanLen}
}

func (args *ChannelArgs) Check() error {
	if args.reqChanLen == 0 {
		return errors.New("The request channel max length (capacity) can not be 0!\n")
	}
	if args.respChanLen == 0 {
		return errors.New("The response channel max length (capacity) can not be 0!\n")
	}
	if args.itemChanLen == 0 {
		return errors.New("The item channel max length (capacity) can not be 0!\n")
	}
	if args.errorChanLen == 0 {
		return errors.New("The error channel max length (capacity) can not be 0!\n")
	}
	return nil
}

var channelArgsTemplate string = "{ reqChanLen: %d, respChanLen: %d," +
	" itemChanLen: %d, errorChanLen: %d }"

func (args *ChannelArgs) String() string {
	if args.description == "" {
		args.description = fmt.Sprintf(channelArgsTemplate, args.reqChanLen, args.respChanLen, args.itemChanLen, args.errorChanLen)
	}
	return args.description
}

func (args *ChannelArgs) ReqChanLen() uint {
	return args.reqChanLen
}

func (args *ChannelArgs) RespChanLen() uint {
	return args.respChanLen
}

func (args *ChannelArgs) ItemChanLen() uint {
	return args.itemChanLen
}

func (args *ChannelArgs) ErrorChanLen() uint {
	return args.errorChanLen
}

func NewPoolBaseArgs(pageDownloaderPoolSize uint32, analyzerPoolSize uint32) PoolBaseArgs {
	return PoolBaseArgs{pageDownloaderPoolSize: pageDownloaderPoolSize, analyzerPoolSize: analyzerPoolSize}
}

func (args *PoolBaseArgs) Check() error {
	if args.pageDownloaderPoolSize == 0 {
		return errors.New("The page downloader pool size can not be 0!\n")
	}
	if args.analyzerPoolSize == 0 {
		return errors.New("The analyzer pool size can not be 0!\n")
	}
	return nil
}

var poolBaseArgsTemplate string = "{ pageDownloaderPoolSize: %d," +
	" analyzerPoolSize: %d }"

func (args *PoolBaseArgs) String() string {
	if args.description == "" {
		args.description =
			fmt.Sprintf(poolBaseArgsTemplate,
				args.pageDownloaderPoolSize,
				args.analyzerPoolSize)
	}
	return args.description
}

func (args *PoolBaseArgs) PageDownloaderPoolSize() uint32 {
	return args.pageDownloaderPoolSize
}

func (args *PoolBaseArgs) AnalyzerPoolSize() uint32 {
	return args.analyzerPoolSize
}
