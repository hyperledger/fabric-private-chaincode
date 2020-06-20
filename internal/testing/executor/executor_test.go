/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package executor

import (
	"crypto/rand"
	"fmt"
	"sync"
	"testing"
	"time"
)

type Task struct {
	name     string
	taskID   int
	callback func(err error)
}

func (t *Task) Invoke() {
	if err := t.doInvoke(); err != nil {
		t.callback(err)
	} else {
		t.callback(nil)
	}
}

func (t *Task) doInvoke() error {
	b := make([]byte, 128)
	_, err := rand.Read(b)
	return err
}

func Test_CreateExecutor(t *testing.T) {

	numWorkers := uint16(8)
	totalTx := 100000

	executor := NewConcurrent("Client", numWorkers)
	executor.Start()
	defer executor.Stop(true)

	var wg sync.WaitGroup
	var mutex sync.RWMutex

	var errs []error
	success := 0

	// create tasks
	var tasks []*Task
	for i := 0; i < totalTx; i++ {
		myTask := &Task{name: "Rudi",
			taskID: i,
			callback: func(err error) {
				defer wg.Done()
				mutex.Lock()
				defer mutex.Unlock()
				if err != nil {
					errs = append(errs, err)
				} else {
					success++
				}
			}}
		tasks = append(tasks, myTask)
	}

	numInvocations := len(tasks)
	wg.Add(numInvocations)

	startTime := time.Now()
	// execute tasks
	for _, task := range tasks {
		if err := executor.Submit(task); err != nil {
			panic(fmt.Sprintf("error submitting task: %s", err))
		}
	}

	// Wait for all tasks to complete
	wg.Wait()

	duration := time.Now().Sub(startTime)

	if numInvocations > 1 {
		fmt.Printf("\n")
		fmt.Printf("*** ---------- Summary: ----------\n")
		fmt.Printf("***   - Workers:         %d\n", numWorkers)
		fmt.Printf("***   - Invocations:     %d\n", numInvocations)
		fmt.Printf("***   - Successfull:     %d\n", success)
		fmt.Printf("***   - Duration:        %s\n", duration)
		fmt.Printf("***   - Rate:            %2.2f/s\n", float64(numInvocations)/duration.Seconds())
		fmt.Printf("*** ------------------------------\n")
	}

	for _, err := range errs {
		fmt.Printf("err: %s", err)
	}
}
