package main

import (
	"sync"
	"time"
)

type CountDownLatch struct {
	sync.Mutex

	count int
	zero  chan struct{}
	done  bool
}

func NewCountDownLatch(count int) *CountDownLatch {
	return &CountDownLatch{
		count: count,
		zero:  make(chan struct{}),
		done:  count == 0,
	}
}

func (l *CountDownLatch) CountDown() {
	l.Lock()
	defer l.Unlock()

	l.count--
	if l.count <= 0 && !l.done {
		l.done = true
		close(l.zero)
	}
}

func (l *CountDownLatch) Wait(duration time.Duration) bool {
	if l.done {
		return true
	}

	select {
	case <-l.zero:
		return true
	case <-time.After(duration):
		return false
	}
}
