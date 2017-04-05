/*
* @Author: wang
* @Date:   2017-04-05 15:31:11
* @Last Modified by:   wang
* @Last Modified time: 2017-04-05 15:32:03
 */

package scheduler

type SchedSummary interface {
	String() string
	Detail() string
	Same(other SchedSummary) bool
}
