/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package executor

import (
	"fmt"
	"math"
	"sync"
	"time"

	"github.com/hyperledger-labs/fabric-private-chaincode/internal/testing/executor/worker"
)

// State is the state of the executor
type State uint8

const (
	// NEW indicates that the executor is new and has not been started yet
	NEW State = iota

	// STARTED indicates that the executor is up and running
	STARTED

	// TERMINATED indicates that the executor has been shut down
	TERMINATED
)

// Executor maintains a pool of workers that execute Tasks in separate Go routines. The caller submits
// a Task to the Executor, which is submitted to an available worker. If a worker is not available then
// the task is queue until one becomes available. The Executor is useful for throttling requests in order
// not to overload the application.
type Executor struct {
	name        string
	state       State
	tasks       chan worker.Task
	terminating chan bool
	pool        *worker.Pool
	wg          sync.WaitGroup
}

// NewConcurrent creates a new, concurrent executor with the given concurrency.
// As tasks are submitted they are queued while waiting for a worker for execution.
//
// - name: The name of the executor (useful for debugging)
// - concurrency: The concurrency, i.e. the number of workers executing concurrently
func NewConcurrent(name string, concurrency uint16) *Executor {
	return New(name, math.MaxInt16, worker.NewPool(name, concurrency))

}

// NewBoundedConcurrent creates a new, concurrent executor with the given concurrency
// and queue length. As tasks are submitted they are queued while waiting for a worker for execution. Once the
// number of queued tasks reaches the given queue length, the Submit operation will block until a worker becomes
// available.
//
// - name: The name of the executor (useful for debugging)
// - concurrency: The concurrency, i.e. the number of workers executing concurrently
// - queueLength: The maximum number of tasks allowed to be queued while waiting for a worker
func NewBoundedConcurrent(name string, concurrency uint16, queueLength uint16) *Executor {
	return New(name, queueLength, worker.NewPool(name, concurrency))

}

// New creates a new, multi-threaded executor with the given options:
// - name: The name of the executor (useful for debugging)
// - queueLength: The maximum number of tasks allowed to be queued while waiting for a worker
//		If queueSize == concurrency then the Submit() function will block until a worker becomes free,
//		otherwise the task will be added to the queue and Submit() will not block.
// pool - The worker pool
func New(name string, queueLength uint16, pool *worker.Pool) *Executor {
	return &Executor{
		name:        name,
		state:       NEW,
		terminating: make(chan bool),
		tasks:       make(chan worker.Task, queueLength),
		pool:        pool,
	}
}

// Start starts the Executor
func (e *Executor) Start() bool {
	if e.state != NEW {
		return false
	}

	// Start the worker pool
	e.pool.Start()

	// Start the task dispatcher
	go e.dispatch()
	e.state = STARTED

	return true
}

// Submit submits a new task
func (e *Executor) Submit(task worker.Task) error {
	if e.state != STARTED {
		return fmt.Errorf("executor [%s] is not started", e.name)

	}

	// Submit the task to the dispatcher
	e.wg.Add(1)
	e.tasks <- task

	return nil
}

// SubmitDelayed submits a new task in the future
func (e *Executor) SubmitDelayed(task worker.Task, delay time.Duration) error {
	if e.state != STARTED {
		return fmt.Errorf("executor [%s] is not started", e.name)
	}

	e.wg.Add(1)

	go func() {
		<-time.After(delay)
		e.Submit(task)
		e.wg.Done()
	}()

	return nil
}

// Wait waits for all outstanding tasks to complete
func (e *Executor) Wait() {
	e.wg.Wait()
}

// Stop stops the executor.
// - wait: If true then the call will block until all outstanding
//   tasks have completed; otherwise the executor will shut down immediately
func (e *Executor) Stop(wait bool) bool {
	if e.state != STARTED {
		return false
	}

	e.state = TERMINATED

	// Wait for the dispatcher to purge its queue
	e.wg.Wait()

	// Stop the dispatcher
	e.terminating <- true

	// Stop the worker pool
	e.pool.Stop(wait)

	return true
}

func (e *Executor) dispatch() {
	for {
		select {
		// Wait for a task
		case task := <-e.tasks:
			e.pool.Submit(task)
			e.wg.Done()

		case <-e.terminating:
			return
		}
	}
}
