# Supervigor

Supervisor for gouroutines.

## Check the doc

[link](https://godoc.org/github.com/mrkaspa/supervigor)

## How to use

The routine to supervise must implement the Runnable interface:

```go
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
	r.doneChan <- true
}
```

Now you must start a supervigor:

```go
sup := NewSupervigor()
run := &Runner{"Michel", true, make(chan bool)}
```

And you can start the runnable like:

```go
sup.SuperviseChan <- RunnableWithName{
    Name: "demo",
    MaxRestarts: 1,
    Runnable: run,
}

```

The first time should fail and be restarted but the second time should work

```go
<- run.doneChan // false
<- run.doneChan // true
```
