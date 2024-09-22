package worker

import (
	"context"
	"fmt"
	"runtime/debug"
)

type ExecutionFn func(interface{}, context.Context) JobRs

type JobRs struct {
	Value interface{}
	Err   error
}

type Job struct {
	ctx    context.Context
	Data   interface{}
	ExecFn ExecutionFn
}

func NewJob(data interface{}, ExecFn ExecutionFn) Job {
	j := Job{
		Data:   data,
		ExecFn: ExecFn,
		ctx:    context.Background(),
	}
	return j
}

func (j Job) Execute() JobRs {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println(fmt.Sprintf("panic: %s - %s", r, string(debug.Stack())))
			fmt.Println("worker err: ", r)
		}
	}()
	return j.ExecFn(j.Data, j.ctx)
}
