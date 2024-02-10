package graph

import (
	"context"
	"sync/atomic"
	"testing"
	"time"
)

func TestQueueBasic(t *testing.T) {
	// # Job dependencies
	// 0 -> 1 -> 2
	// 0 -> 3
	// 3 -> 4
	// 2 -> 4

	// Invariants
	// 1. A job can only be started if all its dependencies are done
	// 2. All jobs eventually finish

	successors := [][]int{
		{1, 3},
		{2},
		{4},
		{4},
		{},
	}

	predecessors := make([][]int, len(successors))
	for i := range predecessors {
		var pred []int
		for j, succs := range successors {
			if contains(succs, i) {
				pred = append(pred, j)
			}
		}
		predecessors[i] = pred
	}

	numJobs := 5
	done := make([]atomic.Bool, numJobs)

	makeJob := func(i int) Job {
		return func(ctx context.Context) {
			// ensure all predecessors are done
			for _, pred := range predecessors[i] {
				if !done[pred].Load() {
					t.Fatalf("job %d cannot start because predecessor %d is not done", i, pred)
				}
			}
			// set done and ensure it was not already done
			if done[i].Swap(true) {
				t.Fatalf("job %d was already done", i)
			}
		}
	}

	const numWorkers = 1
	q := NewQueue(numWorkers)
	items := make([]*QueueItem, numJobs)
	for i := range items {
		items[i] = q.AddItem(makeJob(i), len(predecessors[i]))
	}
	for i := range items {
		// add successor dependencies
		for _, succ := range successors[i] {
			items[i].AddSuccessor(items[succ])
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	q.Start(ctx)
	q.RunJobs(ctx)
	q.Stop()
	// check that all jobs are done
	for i := range done {
		if !done[i].Load() {
			t.Fatalf("job %d was not done", i)
		}
	}
}

func contains(s []int, e int) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}
