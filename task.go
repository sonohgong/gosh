package gosh

import (
	"fmt"
	"sync"
	"sync/atomic"

	"github.com/google/uuid"
)

type TaskStatus int

const (
	Ready TaskStatus = iota
	Progressing
	Success
	Failure
)

type TaskID string

type Task struct {
	id     TaskID
	name   string
	status TaskStatus
	data   interface{}
}

type taskManager struct {
	lock          sync.Mutex
	tasks         sync.Map
	ready         map[TaskID]bool
	handler       func(data interface{}) error
	concurrent    int64
	maxConcurrent uint64
}

func NewTaskManager(handler func(data interface{}) error) taskManager {
	return taskManager{
		ready:         make(map[TaskID]bool),
		tasks:         sync.Map{},
		handler:       handler,
		maxConcurrent: 1000,
	}
}

// NewTasks adds n new tasks to our task queue
func (self *taskManager) NewTasks(n int) {
	self.lock.Lock()
	defer self.lock.Unlock()

	for i := 0; i < n; i += 1 {
		id, err := uuid.NewRandom()
		if err != nil {
			panic("failed to assign task ID")
		}
		taskID := TaskID(id.String())
		self.tasks.Store(taskID, &Task{
			id:     taskID,
			name:   fmt.Sprintf("task %d", id),
			status: Ready,
		})
		self.ready[taskID] = true
	}
}

// StartTasks runs all tasks that are ready
func (self *taskManager) StartTasks() {
	self.lock.Lock()
	defer self.lock.Unlock()

	free := int64(self.maxConcurrent) - atomic.LoadInt64(&self.concurrent)
	for taskID := range self.ready {
		if free <= 0 {
			break
		}
		delete(self.ready, taskID)
		go self.handleTask(taskID)
		free--
	}
}

// handleTask runs a single task and updates its status
func (self *taskManager) handleTask(id TaskID) {
	atomic.AddInt64(&self.concurrent, 1)
	taskInterface, ok := self.tasks.Load(id)
	if !ok {
		panic("missing task")
	}
	task := taskInterface.(*Task)
	task.status = Progressing
	err := self.handler(task.data)
	if err != nil {
		task.status = Failure
	} else {
		task.status = Success
	}
	atomic.AddInt64(&self.concurrent, -1)
}
