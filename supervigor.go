package supervigor

import (
	"fmt"
	"sync"
	"time"
)

// A Supervigor starts and supervises goroutines
type Supervigor struct {
	runnables map[string]*runnableWithChan
	mapMutex  *sync.Mutex
}

type runnableWithChan struct {
	runnable    Runnable
	rchan       chan bool
	restarts    int
	maxRestarts int
	maxTime     int
	restartTime time.Time
}

// Runnable is the main thread for the goroutine that will be
// started and restarted
type Runnable interface {
	Run()
}

// NewSupervigor returns and runs in a goroutine the Supervigor
func NewSupervigor() Supervigor {
	s := Supervigor{
		runnables: map[string]*runnableWithChan{},
		mapMutex:  &sync.Mutex{},
	}
	return s
}

// Supervise a Runnable
func (s *Supervigor) Supervise(name string, maxRestarts int, maxTime int, r Runnable) {
	s.mapMutex.Lock()
	rwc, ok := s.runnables[name]
	if !ok {
		rwc = &runnableWithChan{
			rchan:       make(chan bool),
			runnable:    r,
			maxTime:     maxTime,
			maxRestarts: maxRestarts,
			restarts:    0,
			restartTime: time.Time{},
		}
		s.runnables[name] = rwc
	}
	s.mapMutex.Unlock()

	go s.supervise(name, rwc)
	go run(rwc)
}

func (s *Supervigor) supervise(name string, rwc *runnableWithChan) {
	if rwc.restartTime.IsZero() {
		<-rwc.rchan
		if rwc.restarts >= rwc.maxRestarts {
			return
		}
		rwc.restarts = 1
		rwc.restartTime = time.Now()
	} else {
		select {
		case <-rwc.rchan:
			rwc.restarts++
			res := time.Now().Sub(rwc.restartTime)

			if int(res.Seconds()) >= rwc.maxTime || rwc.restarts >= rwc.maxRestarts {
				fmt.Printf("removing %s from the supervisor\n", name)
				s.mapMutex.Lock()
				delete(s.runnables, name)
				s.mapMutex.Unlock()
				return
			}
		case <-time.NewTimer(3 * time.Second).C:
			rwc.restartTime = time.Time{}
			go s.supervise(name, rwc)
			return
		}
	}

	fmt.Printf("restarting #%d %s \n", rwc.restarts, name)
	go run(rwc)
	go s.supervise(name, rwc)
}

// run the go routine if it panics notify to the supervigor
func run(rwc *runnableWithChan) {
	defer func() {
		if r := recover(); r != nil {
			rwc.rchan <- true
		}
	}()
	rwc.runnable.Run()
}
