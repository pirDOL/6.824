package mapreduce

import (
	"fmt"
)

//
// schedule() starts and waits for all tasks in the given phase (Map
// or Reduce). the mapFiles argument holds the names of the files that
// are the inputs to the map phase, one per map task. nReduce is the
// number of reduce tasks. the registerChan argument yields a stream
// of registered workers; each item is the worker's RPC address,
// suitable for passing to call(). registerChan will yield all
// existing registered workers (if any) and new ones as they register.
//
func schedule(jobName string, mapFiles []string, nReduce int, phase jobPhase, registerChan chan string) {
	var ntasks int
	var ntasksOfOtherPhase int // number of inputs (for reduce) or outputs (for map)
	switch phase {
	case mapPhase:
		ntasks = len(mapFiles)
		ntasksOfOtherPhase = nReduce
	case reducePhase:
		ntasks = nReduce
		ntasksOfOtherPhase = len(mapFiles)
	}

	fmt.Printf("Schedule: %v %v tasks (%d I/Os)\n", ntasks, phase, ntasksOfOtherPhase)

	// All ntasks tasks have to be scheduled on workers, and only once all of
	// them have been completed successfully should the function return.
	// Remember that workers may fail, and that any given worker may finish
	// multiple tasks.
	//
	// TODO TODO TODO TODO TODO TODO TODO TODO TODO TODO TODO TODO TODO
	//

	ts := newTaskScheduler(ntasks)
	ts.run(jobName, mapFiles, phase, ntasksOfOtherPhase, registerChan)
	fmt.Printf("Schedule: %v phase done\n", phase)
}

type TaskScheduler struct {
	idleWorkers    []string
	taskIterator   *TaskIterator
	doTaskDoneChan chan *doTaskContext
}

func newTaskScheduler(nTotalTask int) *TaskScheduler {
	return &TaskScheduler{
		idleWorkers:    make([]string, 0),
		doTaskDoneChan: make(chan *doTaskContext),
		taskIterator:   newTaskIterator(nTotalTask),
	}
}

func (s *TaskScheduler) run(jobName string, mapFiles []string, phase jobPhase, nOtherPhase int, idleWorkerChan chan string) {
	for isRunning := true; isRunning; {
		select {
		case idleWorker := <-idleWorkerChan:
			taskNumber := s.taskIterator.next()
			switch taskNumber {
			case END_OF_TASK:
				isRunning = false
			case HAS_PENDING_TASK:
				s.idleWorkers = append(s.idleWorkers, idleWorker)
			default:
				go func(tn int, wk string) {
					args := DoTaskArgs{
						JobName:       jobName,
						File:          mapFiles[tn],
						Phase:         phase,
						TaskNumber:    tn,
						NumOtherPhase: nOtherPhase,
					}
					ok := call(wk, "Worker.DoTask", args, new(struct{}))
					s.doTaskDoneChan <- &doTaskContext{ok, wk, tn}
				}(taskNumber, idleWorker)
			}
		case ctx := <-s.doTaskDoneChan:
			s.taskIterator.done(ctx.tn, ctx.ok)
			if s.taskIterator.eof() {
				isRunning = false
			} else {
				var wk string
				if ctx.ok {
					wk = ctx.wk
				} else {
					nIdleWorker := len(s.idleWorkers)
					if nIdleWorker != 0 {
						wk = s.idleWorkers[nIdleWorker-1]
						s.idleWorkers = s.idleWorkers[:nIdleWorker-1]
					}
				}
				if len(wk) != 0 {
					go func() { idleWorkerChan <- wk }()
				}
			}
			// default:
			// 	fmt.Println(s.taskIterator)
		}
	}
}

const (
	END_OF_TASK       = -1
	HAS_PENDING_TASK  = -2
	TASK_NUMBER_ERROR = -3
)

type doTaskContext struct {
	ok bool
	wk string
	tn int
}

type TaskIterator struct {
	failedTaskNumbers []int
	nTotalTask        int
	nDoingTask        int
}

func newTaskIterator(nTotalTask int) *TaskIterator {
	return &TaskIterator{
		failedTaskNumbers: make([]int, 0),
		nTotalTask:        nTotalTask,
	}
}

func (i *TaskIterator) eof() bool {
	return len(i.failedTaskNumbers) == 0 && i.nTotalTask == 0 && i.nDoingTask == 0
}

func (i *TaskIterator) next() int {
	if i.eof() {
		return END_OF_TASK
	}

	nFailedTask := len(i.failedTaskNumbers)
	if nFailedTask != 0 {
		taskNumber := i.failedTaskNumbers[nFailedTask-1]
		i.failedTaskNumbers = i.failedTaskNumbers[:nFailedTask-1]
		i.nDoingTask++
		return taskNumber
	}
	if i.nTotalTask != 0 {
		taskNumber := i.nTotalTask - 1
		i.nTotalTask--
		i.nDoingTask++
		return taskNumber
	}
	if i.nDoingTask != 0 {
		return HAS_PENDING_TASK
	}
	return TASK_NUMBER_ERROR
}

func (i *TaskIterator) done(taskNumber int, ok bool) {
	i.nDoingTask--
	if !ok {
		i.failedTaskNumbers = append(i.failedTaskNumbers, taskNumber)
	}
}
