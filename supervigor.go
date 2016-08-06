package supervigor

import (
	"fmt"
	"sync"
)

type Supervigor struct {
	SuperviseChan chan RunnableWithName
	runnables     map[string]*runnableWithChan
	mapMutex *sync.Mutex
}

type runnableWithChan struct {
	rchan    chan bool
	restarts int
	runnable Runnable
}

type RunnableWithName struct {
	Name        string
	MaxRestarts int
	Runnable    Runnable
}

type Runnable interface {
	Run()
}

func NewSupervigor() Supervigor {
	s := Supervigor{
		SuperviseChan: make(chan RunnableWithName),
		runnables: map[string]*runnableWithChan{},
		mapMutex: &sync.Mutex{},
	}
	go s.Run()
	return s
}

func (s *Supervigor) Run() {
	for {
		select {
		case rwn := <-s.SuperviseChan:
			fmt.Printf("supervising %s \n", rwn.Name)
			s.supervise(rwn.Name, rwn.MaxRestarts, rwn.Runnable)
		}
	}
}

func (s *Supervigor) supervise(name string, maxRestarts int, r Runnable) {
	s.mapMutex.Lock()
	rwc, ok := s.runnables[name]
	if !ok{
		rwc = &runnableWithChan{
			rchan: make(chan bool),
			runnable: r,
		}
		s.runnables[name] = rwc
	}
	s.mapMutex.Unlock()

	//supervise the goroutine
	go func() {
		<-rwc.rchan
		s.mapMutex.Lock()
		s.runnables[name].restarts++
		if s.runnables[name].restarts <= maxRestarts {
			fmt.Printf("restarting #%d %s \n", s.runnables[name].restarts, name)
			s.SuperviseChan <- RunnableWithName{name, maxRestarts, r}
		}
		s.mapMutex.Unlock()
	}()

	go run(rwc)
}

func run(rwc *runnableWithChan) {
	defer func() {
		if r := recover(); r != nil {
			rwc.rchan <- true
		}
	}()
	rwc.runnable.Run()
}

