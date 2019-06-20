package taskq

import (
	"sync"
)

type TaskFunc func(args ...interface{})

type Task struct {
	tfn TaskFunc
	//args chan interface{}
	args []interface{}
}

type TaskQueue struct {
	q     chan *Task
	start chan struct{}
	done  chan struct{}
	wg    sync.WaitGroup
}

func executeTask(t *Task) {
	t.tfn(t.args...)
}

func (tq *TaskQueue) worker() {
	<-tq.start
	defer tq.wg.Done()

	for t := range tq.q {
		executeTask(t)
	}
}

func (tq *TaskQueue) Start() {
	close(tq.start)
}

func (tq *TaskQueue) Stop() {
	close(tq.q)
	close(tq.done)
	tq.wg.Wait()
}

func (tq *TaskQueue) Enqueue(t TaskFunc, args ...interface{}) {
	select {
	case <-tq.done:
		return
	default:
	}

	tk := &Task{
		tfn: t,
		//args: make(chan interface{}, 10),
		args: make([]interface{}, len(args)),
	}

	for k, arg := range args {
		tk.args[k] = arg
	}
	//close(tk.args)

	tq.q <- tk
}

func NewTaskQueue(wkrcnt int) *TaskQueue {
	tq := &TaskQueue{
		q:     make(chan *Task, 100),
		start: make(chan struct{}),
		done:  make(chan struct{}),
	}

	tq.wg.Add(wkrcnt)

	for i := 0; i < wkrcnt; i++ {
		go tq.worker()
	}

	return tq
}
