/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package worker

// State is the state of a Worker
type State uint8

const (
	// READY indicates that a worker is ready to accept a new task
	READY State = iota

	// STOPPED indicates that the worker has terminated
	STOPPED
)

// Task is the task that the Worker invokes
type Task interface {
	Invoke()
}

// Events receives event notifications from the worker
type Events interface {
	// StateChange indicates the new state of the worker
	StateChange(w *Worker, state State)

	// TaskStarted indicates that the given worker has started executing the given task
	TaskStarted(w *Worker, task Task)

	// TaskCompleted indicates that the given worker has completed the given task
	TaskCompleted(w *Worker, task Task)
}

// Worker invokes a Task
type Worker struct {
	name   string
	events Events
	task   chan Task
	done   chan bool
}

func newWorker(name string, events Events) *Worker {
	return &Worker{
		name:   name,
		events: events,
		task:   make(chan Task),
		done:   make(chan bool),
	}
}

// Name returns the name of the worker (useful for debugging)
func (w *Worker) Name() string {
	return w.name
}

// Submit submits a task
func (w *Worker) Submit(task Task) {
	w.task <- task
}

func (w *Worker) invoke(task Task) {
	defer w.events.TaskCompleted(w, task)

	w.events.TaskStarted(w, task)
	task.Invoke()
}

// Start starts the worker
func (w *Worker) Start() {
	go func() {
		for {
			// Inform the events that I'm available
			w.events.StateChange(w, READY)

			select {
			case task := <-w.task:
				w.invoke(task)

			case <-w.done:
				w.events.StateChange(w, STOPPED)
				return
			}
		}
	}()
}

// Stop stops the worker
func (w *Worker) Stop() {
	w.done <- true
}
