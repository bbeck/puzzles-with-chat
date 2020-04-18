package pubsub

import (
	"fmt"
	"sync"
	"time"
)

// Writer whose Write method always returns an error.
type ErrWriter struct{}

func (ErrWriter) Write(p []byte) (n int, err error) {
	return 0, fmt.Errorf("unwritable")
}

// CountDownLatch is a synchronization aid that allows one or more goroutines to
// wait until a set of operations completes.
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
	l.Lock()

	if l.done {
		l.Unlock()
		return true
	}

	l.Unlock()

	select {
	case <-l.zero:
		return true
	case <-time.After(duration):
		return false
	}
}
