package main

import (
	"sort"
	"strings"
)

type TestSuite struct {
	Task
}

func NewTestSuite() *TestSuite {
	return &TestSuite{}
}

func (ts *TestSuite) AddEvent(ev Event) {
	ts.Extend(ev.Time)
	if ev.Package == "" {
		ts.Events.Add(ev)
		return
	}

	elems := []string{ev.Package}
	if ev.Test != "" {
		elems = append(elems, strings.Split(ev.Test, "/")...)
	}

	ts.Add(elems, ev)
}

// Task is either a package, a Test or a Subtest
type Task struct {
	Name string
	Span
	Events Events

	ByName map[string]*Task
	Sub    []*Task
}

type Events []Event

func (evs *Events) Add(ev Event) { *evs = append(*evs, ev) }

func (task *Task) Add(path []string, ev Event) {
	task.Extend(ev.Time)
	if len(path) == 0 {
		task.Events.Add(ev)
		return
	}

	sub := task.EnsureSub(path[0])
	sub.Add(path[1:], ev)
}

func (task *Task) EnsureSub(name string) *Task {
	if task.ByName == nil {
		task.ByName = map[string]*Task{}
	}

	sub, ok := task.ByName[name]
	if ok {
		return sub
	}

	sub = &Task{
		Name: name,
	}
	task.ByName[name] = sub
	task.Sub = append(task.Sub, sub)

	return sub
}

func (task *Task) Sort() {
	sort.Slice(task.Sub, func(i, k int) bool {
		return task.Sub[i].Start.Before(task.Sub[k].Start)
	})

	for _, sub := range task.Sub {
		sub.Sort()
	}
}
