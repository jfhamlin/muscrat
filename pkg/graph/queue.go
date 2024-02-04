package graph

import (
	"context"
	"sync"
	"sync/atomic"
)

type (
	Queue struct {
		numWorkers int
		wg         sync.WaitGroup

		remainingItems atomic.Int32
		numItems       int32
		runChan        chan *QueueItem

		initiallyRunnable []*QueueItem
	}

	QueueItem struct {
		job             func()
		successors      []*QueueItem
		activationCount atomic.Int32
		activationLimit int32
	}
)

////////////////////////////////////////////////////////////////////////////////
// Queue

func NewQueue(numWorkers int) *Queue {
	return &Queue{
		numWorkers: numWorkers,
	}
}

func (q *Queue) AddItem(job func(), activationLimit int) *QueueItem {
	q.numItems++
	item := &QueueItem{
		job:             job,
		activationLimit: int32(activationLimit),
	}
	item.activationCount.Store(item.activationLimit)
	if activationLimit == 0 {
		q.addInitialItem(item)
	}
	return item
}

func (q *Queue) addInitialItem(item *QueueItem) {
	q.initiallyRunnable = append(q.initiallyRunnable, item)
}

func (q *Queue) Run(ctx context.Context) {
	// room for all items to be runnable at once, so workers don't
	// block.
	q.runChan = make(chan *QueueItem, q.numItems)
	q.wg.Add(q.numWorkers)
	for i := 0; i < q.numWorkers; i++ {
		go q.runWorker(ctx)
	}
	for _, item := range q.initiallyRunnable {
		q.runChan <- item
	}

	q.wg.Wait()
	q.reset()
}

func (q *Queue) runWorker(ctx context.Context) {
	defer q.wg.Done()

	for {
		select {
		case <-ctx.Done():
			return
		case item, ok := <-q.runChan:
			if !ok {
				return
			}
			item.Run(q)
			if newVal := q.remainingItems.Add(-1); newVal == 0 {
				close(q.runChan)
			}
		}
	}
}

func (q *Queue) reset() {
	q.remainingItems.Store(q.numItems)
	q.runChan = nil
}

////////////////////////////////////////////////////////////////////////////////
// QueueItem

func (qi *QueueItem) AddSuccessor(successor *QueueItem) {
	qi.successors = append(qi.successors, successor)
}

func (qi *QueueItem) Run(q *Queue) {
	qi.job()
	qi.updateSuccessors(q)
	qi.resetActivationCount()
}

func (qi *QueueItem) updateSuccessors(q *Queue) {
	for _, succ := range qi.successors {
		if newVal := succ.activationCount.Add(-1); newVal == 0 {
			q.runChan <- succ
		}
	}
}

func (qi *QueueItem) resetActivationCount() {
	qi.activationCount.Store(qi.activationLimit)
}
