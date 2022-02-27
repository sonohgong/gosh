package gosh

import (
	"sync"
	time "time"
)

type scheduler struct {
	interrupt time.Duration
	f         func()
	mu        sync.Mutex
	timer     *time.Timer
	done      chan bool
}

func NewScheduler(interrupt time.Duration, f func()) scheduler {
	return scheduler{done: make(chan bool, 1), interrupt: interrupt, f: f}
}

func (self *scheduler) Run() {
	self.mu.Lock()
	defer self.mu.Unlock()

	if self.timer != nil {
		self.timer.Stop()
	}
	self.f()
	self.timer = time.AfterFunc(self.interrupt, self.Run)
}

func (self *scheduler) Stop() {
	self.mu.Lock()
	defer self.mu.Unlock()

	if self.timer != nil {
		self.timer.Stop()
	}

}

func (self *scheduler) Done() {
	self.Stop()
	self.done <- true

}

func (self *scheduler) Wait() {
	<-self.done
}
