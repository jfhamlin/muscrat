package graph

import (
	"context"
	"runtime"
	"sync"
	"sync/atomic"
)

type (
	Job func(ctx context.Context)

	Queue struct {
		numWorkers int
		wg         sync.WaitGroup
		stop       func()

		remainingItems atomic.Int32
		// sent when all jobs are done
		jobsDone chan struct{}

		numItems int32
		runnable *lockFreeStack

		initiallyRunnable []*QueueItem

		cond *sync.Cond
	}

	QueueItem struct {
		job             Job
		successors      []*QueueItem
		activationCount atomic.Int32
		activationLimit int32

		// used in the lock-free stack
		next *QueueItem
	}

	lockFreeStack struct {
		top atomic.Value
	}
)

////////////////////////////////////////////////////////////////////////////////
// lockFreeStack

func newLockFreeStack() *lockFreeStack {
	s := &lockFreeStack{}
	s.top.Store((*QueueItem)(nil))
	return s
}

func (s *lockFreeStack) Push(item *QueueItem) {
	for {
		oldTop := s.top.Load()
		item.next = oldTop.(*QueueItem)
		if s.top.CompareAndSwap(oldTop, item) {
			return
		}
	}
}

func (s *lockFreeStack) Pop() *QueueItem {
	for {
		oldTop, _ := s.top.Load().(*QueueItem)
		if oldTop == nil {
			return nil
		}
		newTop := oldTop.next
		if s.top.CompareAndSwap(oldTop, newTop) {
			oldTop.next = nil
			return oldTop
		}
	}
}

func (s *lockFreeStack) IsEmpty() bool {
	top, _ := s.top.Load().(*QueueItem)
	return top == nil
}

////////////////////////////////////////////////////////////////////////////////
// Queue

func NewQueue(numWorkers int) *Queue {
	return &Queue{
		numWorkers: numWorkers,
		cond:       sync.NewCond(&sync.Mutex{}),
		// buffered to avoid blocking if a run is canceled.
		jobsDone: make(chan struct{}, 1),
	}
}

func (q *Queue) AddItem(job Job, activationLimit int) *QueueItem {
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

func (q *Queue) Start(ctx context.Context) {
	if q.runnable != nil {
		panic("queue already started")
	}

	ctx, cancel := context.WithCancel(ctx)
	q.stop = cancel

	q.runnable = newLockFreeStack()
	q.wg.Add(q.numWorkers)
	for i := 0; i < q.numWorkers; i++ {
		i := i
		go func() {
			<-ctx.Done()
		}()
		go q.runWorker(ctx, i)
	}
}

func (q *Queue) Stop() {
	if q.stop == nil {
		return
	}
	q.stop()
	q.stop = nil
	q.signalAll()
	q.wg.Wait()
}

func (q *Queue) RunJobs(ctx context.Context) error {
	q.remainingItems.Store(q.numItems)

	for _, item := range q.initiallyRunnable {
		q.runnable.Push(item)
	}
	q.signalAll()

	select {
	case <-ctx.Done():
		// TODO: cancel all jobs
		return ctx.Err()
	case <-q.jobsDone:
		return nil
	}
}

func (q *Queue) runWorker(ctx context.Context, id int) {
	defer func() {
		q.signalAll() // wake up all other workers
		q.wg.Done()
	}()

	const maxBackoff = 10

	backoff := maxBackoff
	for {
		select {
		case <-ctx.Done():
			return
		default:
		}
		item := q.runnable.Pop()
		if item == nil {
			if backoff > 0 {
				runtime.Gosched()
				backoff--
				continue
			}
			////////////////////////////////////////////////////////////////////////////////
			// wait
			q.cond.L.Lock()
			// double check that there are no items
			if !q.runnable.IsEmpty() {
				q.cond.L.Unlock()
				continue
			}
			q.cond.Wait()
			q.cond.L.Unlock()
			backoff = maxBackoff
			////////////////////////////////////////////////////////////////////////////////
			continue
		}
		backoff = maxBackoff
		item.Run(ctx, q)
		if newVal := q.remainingItems.Add(-1); newVal == 0 {
			q.jobsDone <- struct{}{}
		}
	}
}

func (q *Queue) signal() {
	q.cond.L.Lock()
	q.cond.Signal()
	q.cond.L.Unlock()
}

func (q *Queue) signalAll() {
	q.cond.L.Lock()
	q.cond.Broadcast()
	q.cond.L.Unlock()
}

////////////////////////////////////////////////////////////////////////////////
// QueueItem

func (qi *QueueItem) AddSuccessor(successor *QueueItem) {
	qi.successors = append(qi.successors, successor)
}

func (qi *QueueItem) Run(ctx context.Context, q *Queue) {
	qi.job(ctx)
	qi.updateSuccessors(q)
	qi.reset()
}

func (qi *QueueItem) updateSuccessors(q *Queue) {
	for _, succ := range qi.successors {
		if newVal := succ.activationCount.Add(-1); newVal == 0 {
			q.runnable.Push(succ)
			q.signal()
		}
	}
}

func (qi *QueueItem) reset() {
	qi.activationCount.Store(qi.activationLimit)
}
