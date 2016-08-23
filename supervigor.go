package supervigor

import (
  "fmt"
  "sync"
)

// A Supervigor starts and supervises goroutines
type Supervigor struct {
  superviseChan chan RunnableWithName
  runnables     map[string]*runnableWithChan
  mapMutex      *sync.Mutex
}

type runnableWithChan struct {
  rchan    chan bool
  restarts int
  runnable Runnable
}

// A RunnableWithName packs the name of the goroutine,
// the max amount of restarts and the Runnble object to run
type RunnableWithName struct {
  Name        string
  MaxRestarts int
  Runnable    Runnable
}

// Runnable is the main thread for the goroutine that will be
// started and restarted
type Runnable interface {
  Run()
}

// NewSupervigor returns and runs in a goroutine the Supervigor
func NewSupervigor() Supervigor {
  s := Supervigor{
    superviseChan: make(chan RunnableWithName),
    runnables:     map[string]*runnableWithChan{},
    mapMutex:      &sync.Mutex{},
  }
  go s.run()
  return s
}

// Supervise a runnable
func (s *Supervigor) Supervise(rwn RunnableWithName) {
  s.superviseChan <- rwn
}

func (s *Supervigor) run() {
  for {
    select {
    case rwn := <-s.superviseChan:
      fmt.Printf("supervising %s \n", rwn.Name)
      s.runAndSupervise(rwn.Name, rwn.MaxRestarts, rwn.Runnable)
    }
  }
}

func (s *Supervigor) runAndSupervise(name string, maxRestarts int, r Runnable) {
  s.mapMutex.Lock()
  rwc, ok := s.runnables[name]
  if !ok {
    rwc = &runnableWithChan{
      rchan:    make(chan bool),
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
      s.superviseChan <- RunnableWithName{name, maxRestarts, r}
    }
    s.mapMutex.Unlock()
  }()

  go run(rwc)
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
