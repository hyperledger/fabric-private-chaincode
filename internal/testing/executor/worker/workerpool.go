/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package worker

import (
	"fmt"
	"sync"
)

type Pool struct {
	name            string
	workers         []*Worker
	availableWorker chan *Worker
	taskWg          sync.WaitGroup
	wg              sync.WaitGroup
}

// NewPool creates a worker Pool with the given Factory
func NewPool(name string, concurrency uint16) *Pool {
	pool := &Pool{
		name:            name,
		availableWorker: make(chan *Worker, concurrency),
		workers:         make([]*Worker, concurrency),
	}

	// Create the workers
	for i := 0; i < int(concurrency); i++ {
		pool.workers[i] = newWorker(fmt.Sprintf("%s-%d", name, i), pool)
	}

	return pool
}

// Name returns the name of the pool
func (p *Pool) Name() string {
	return p.name
}

// Start starts the pool
func (p *Pool) Start() {
	p.wg.Add(len(p.workers))

	// Start the workers
	for _, w := range p.workers {
		w.Start()
	}
}

// Stop stops the pool and optionally waits until all tasks have completed
func (p *Pool) Stop(wait bool) {

	if wait {
		// Wait for all the tasks to complete
		p.taskWg.Wait()
	} else {
	}

	// Shut down the workers
	for i := 0; i < len(p.workers); i++ {
		w := <-p.availableWorker
		w.Stop()
	}
	// Wait for all of the workers to stop
	p.wg.Wait()
}

// Submit submits a Task for execution
func (p *Pool) Submit(task Task) {
	p.taskWg.Add(1)

	// Wait for an available worker
	w := <-p.availableWorker

	// Submit the task to the worker
	w.Submit(task)
}

// StateChange is invoked when the state of the Worker changes
func (p *Pool) StateChange(w *Worker, state State) {
	switch state {
	case READY:
		p.availableWorker <- w
		break

	case STOPPED:
		p.wg.Done()
		break

	default:
		break
	}
}

// TaskStarted is invoked when the given Worker begins executing the given Task
func (p *Pool) TaskStarted(w *Worker, task Task) {
	// Nothing to do
}

// TaskCompleted is invoked when the given Worker completed executing the given Task
func (p *Pool) TaskCompleted(w *Worker, task Task) {
	p.taskWg.Done()
}
