/*
* @Author: wangshuo
* @Date:   2017-04-11 11:04:07
* @Last Modified by:   wangshuo
* @Last Modified time: 2017-04-11 11:04:35
 */

package base

import "logging"

func NewLogger() logging.Logger {
	return logging.NewSimpleLogger()
}
