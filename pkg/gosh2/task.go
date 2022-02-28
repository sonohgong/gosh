package gosh2

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

const (
	MaxConcurrent = 1000
)

type TaskID string

type Task struct {
	id     TaskID
	name   string
	status TaskStatus
	data   interface{}
}

type taskManager struct {
	lock    sync.Mutex
	tasks   []*Task
	handler func(data interface{}) error
	running int32
	done    chan bool
}

func NewTaskManager(handler func(data interface{}) error) taskManager {
	return taskManager{
		tasks:   make([]*Task, 0),
		handler: handler,
		done:    make(chan bool),
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
		self.tasks = append(self.tasks, &Task{
			id:     taskID,
			name:   fmt.Sprintf("task %d", id),
			status: Ready,
		})
	}
}

// StartTasks runs all tasks that are ready
func (self *taskManager) StartTasks() {
	self.lock.Lock()
	defer self.lock.Unlock()

	fmt.Println(len(self.tasks))
	if len(self.tasks) == 0 {
		self.done <- true
		return
	}

	trim := 0
	for _, task := range self.tasks {
		if task.status == Success || task.status == Failure {
			trim++
		} else {
			break
		}
	}
	self.tasks = self.tasks[trim:]

	free := MaxConcurrent - int(atomic.LoadInt32(&self.running))
	for _, task := range self.tasks {
		if free <= 0 {
			break
		}
		if task.status == Ready {
			go self.handleTask(task)
			free--
		}
	}
}

// handleTask runs a single task and updates its status
func (self *taskManager) handleTask(task *Task) {
	atomic.AddInt32(&self.running, 1)
	task.status = Progressing
	err := self.handler(task.data)
	if err != nil {
		task.status = Failure
	} else {
		task.status = Success
	}
	atomic.AddInt32(&self.running, -1)
}

func (self *taskManager) Wait() {
	<-self.done
}
