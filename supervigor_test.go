package supervigor_test

import (
  "fmt"
  "testing"
  "time"

  "github.com/mrkaspa/supervigor"
  "github.com/stretchr/testify/assert"
)

type Runner struct {
  name      string
  mustPanic bool
  doneChan  chan bool
}

func (r *Runner) Run() {
  if r.mustPanic {
    r.doneChan <- false
    r.mustPanic = false
    panic("")
  }
  fmt.Printf("Hi I'm %s\n", r.name)
  r.doneChan <- true
}

func TestNewSupervigor(t *testing.T) {
  t.Run("HappyPath", whenIsRestarted)
  t.Run("UnhappyPath", whenIsNotRestarted)
}

func whenIsRestarted(t *testing.T) {
  sup := supervigor.NewSupervigor()
  run := &Runner{"Michel", true, make(chan bool)}
  sup.Supervise("demo", 1, run)
  assert.False(t, <-run.doneChan)
  assert.True(t, <-run.doneChan)
}

func whenIsNotRestarted(t *testing.T) {
  sup := supervigor.NewSupervigor()
  run := &Runner{"Michel", true, make(chan bool)}
  sup.Supervise("demo", 0, run)
  assert.False(t, <-run.doneChan)
  select {
  case <-run.doneChan:
    t.Error("Should not enter here")
  case <-time.NewTimer(1 * time.Second).C:
  }
}
