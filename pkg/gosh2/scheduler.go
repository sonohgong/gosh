package gosh2

import (
	time "time"
)

type scheduler struct {
	interrupt time.Duration
	f         func()
	signaler  chan struct{}
	done      chan bool
}

func NewScheduler(interrupt time.Duration, f func()) scheduler {
	return scheduler{done: make(chan bool, 1), interrupt: interrupt, f: f}
}

func (self *scheduler) Start() {
	go func() {
		for {
			select {
			case <-time.NewTimer(self.interrupt).C:
				break
			case <-self.signaler:
				break
			case <-self.done:
				return
			}
			self.f()
		}
	}()
}

func (self *scheduler) Run() {
	self.signaler <- struct{}{}
}

func (self *scheduler) Stop() {
	self.done <- true
}

func (self *scheduler) Wait() {
	<-self.done
}
