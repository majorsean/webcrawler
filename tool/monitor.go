/*
* @Author: wangshuo
* @Date:   2017-04-19 11:35:03
* @Last Modified by:   wangshuo
* @Last Modified time: 2017-04-24 12:12:14
 */

package tool

import (
	"errors"
	"fmt"
	"runtime"
	"time"
	sched "webcrawler/scheduler"
)

type Record func(level byte, content string)

func Monitoring(
	scheduler sched.Scheduler,
	intervalNs time.Duration,
	maxIdleCount uint,
	autoStop bool,
	detailSummary bool,
	record Record) <-chan uint64 {
	if scheduler == nil {
		panic(errors.New("The scheduler is invalid!"))
	}

	if intervalNs < time.Millisecond {
		intervalNs = time.Millisecond
	}

	if maxIdleCount < 1000 {
		maxIdleCount = 1000
	}

	stopNotifier := make(chan byte, 1)

	reportError(scheduler, record, stopNotifier)
	recordSummary(scheduler, detailSummary, record, stopNotifier)

	checkCountChan := make(chan uint64, 2)

	checkStatus(scheduler, intervalNs, maxIdleCount, autoStop, checkCountChan, record, stopNotifier)

	return checkCountChan
}

func reportError(scheduler sched.Scheduler, record Record, stopNotifier <-chan byte) {
	go func() {
		waitForSchedulerStart(scheduler)

		for {
			select {
			case <-stopNotifier:
				return
			default:
			}

			errorChan := scheduler.ErrorChan()
			if errorChan == nil {
				return
			}
			err := <-errorChan
			if err != nil {
				errMsg := fmt.Sprintf("Error (received from error channel): %s", err)
				record(2, errMsg)
			}
			time.Sleep(time.Millisecond)
		}
	}()
}

var summaryForMonitoring = "Monitor - Collected information[%d]:\n" +
	"  Goroutine number: %d\n" +
	"  Scheduler:\n%s" +
	"  Escaped time: %s\n"

func recordSummary(scheduler sched.Scheduler, detailSummary bool, record Record, stopNotifier <-chan byte) {
	var recordCount uint64 = 1
	startTime := time.Now()
	var prevSchedSummary sched.SchedSummary
	var prevNumGoroutine int
	go func() {
		waitForSchedulerStart(scheduler)
		for {
			select {
			case <-stopNotifier:
				return
			default:
			}

			currNumGoroutine := runtime.NumGoroutine()
			currSchedSummary := scheduler.Summary("	")

			if currNumGoroutine != prevNumGoroutine ||
				!currSchedSummary.Same(prevSchedSummary) {
				schedSummaryStr := func() string {
					if detailSummary {
						return currSchedSummary.Detail()
					} else {
						return currSchedSummary.String()
					}
				}()
				info := fmt.Sprintf(summaryForMonitoring, recordCount, currNumGoroutine, schedSummaryStr, time.Since(startTime).String())
				record(0, info)
				prevNumGoroutine = currNumGoroutine
				prevSchedSummary = currSchedSummary
				recordCount++
			}
			time.Sleep(time.Millisecond)
		}
	}()
}

var msgReachMaxIdleCount = "The scheduler has been idle for a period of time" +
	" (about %s)." +
	" Now consider what stop it."

var msgStopScheduler = "Stop scheduler...%s."

func checkStatus(scheduler sched.Scheduler, intervalNs time.Duration, maxIdleCount uint, autoStop bool, checkCountChan chan<- uint64, record Record, stopNotifier chan<- byte) {
	var checkCount uint64
	go func() {
		defer func() {
			stopNotifier <- 1
			stopNotifier <- 2
			checkCountChan <- checkCount
		}()
		waitForSchedulerStart(scheduler)

		var idleCount uint
		var firstIdleTime time.Time
		for {
			if scheduler.Idle() {
				idleCount++
				if idleCount == 1 {
					firstIdleTime = time.Now()
				}
				if idleCount > maxIdleCount {
					msg := fmt.Sprintf(msgReachMaxIdleCount, time.Since(firstIdleTime).String())
					record(0, msg)
					if scheduler.Idle() {
						if autoStop {
							var result string
							if scheduler.Stop() {
								result = "success"
							} else {
								result = "failing"
							}
							msg = fmt.Sprintf(msgStopScheduler, result)
							record(0, msg)
						}
						break
					} else {
						if idleCount > 0 {
							idleCount = 0
						}
					}
				}
			} else {
				if idleCount > 0 {
					idleCount = 0
				}
			}
			checkCount++
			time.Sleep(time.Millisecond)
		}
	}()
}

func waitForSchedulerStart(scheduler sched.Scheduler) {
	for !scheduler.Running() {
		time.Sleep(time.Millisecond)
	}
}
