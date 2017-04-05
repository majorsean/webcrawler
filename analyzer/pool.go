/*
* @Author: wang
* @Date:   2017-04-05 14:45:39
* @Last Modified by:   wang
* @Last Modified time: 2017-04-05 14:47:09
 */

package analyzer

type AnalyzerPool interface {
	Take() (Analyzer, error)
	Return(analyzer Analyzer) error
	Total() uint32
	Used() uint32
}
