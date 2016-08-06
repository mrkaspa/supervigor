package supervigor

import (
	"fmt"
	"testing"
	"github.com/stretchr/testify/assert"
)

type Runner struct{
	name string
	mustPanic bool
	doneChan chan bool
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
	sup := NewSupervigor()
	run := &Runner{"Michel", true, make(chan bool)}
	sup.SuperviseChan <- RunnableWithName{
		Name: "demo",
		MaxRestarts: 1,
		Runnable: run,
	}
	assert.False(t, <- run.doneChan)
	assert.True(t, <- run.doneChan)
}

func TestNewSupervigorMustFail(t *testing.T) {
	sup := NewSupervigor()
	run := &Runner{"Michel", true, make(chan bool)}
	sup.SuperviseChan <- RunnableWithName{
		Name: "demo",
		MaxRestarts: 0,
		Runnable: run,
	}
	assert.False(t, <- run.doneChan)
}
