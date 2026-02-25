package main

import "fmt"

type Task struct {
	Err    error
	Result any
	f      func(any) (any, error)
}

func (t *Task) process() {
	fmt.Println("process")
	t.Result, t.Err = t.f(t.Result)
}
func NewTask(f func(any) (any, error)) *Task {
	return &Task{
		f: f,
	}
}
