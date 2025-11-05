package queue

import (
	"container/heap"
	"sync"
	"time"
)

// ------------------------------
// Job definition
// ------------------------------

type JobStatus string

const (
	PENDING   JobStatus = "PENDING"
	RUNNING   JobStatus = "RUNNING"
	COMPLETED JobStatus = "COMPLETED"
	FAILED    JobStatus = "FAILED"
)

type Job struct {
	ID          string
	Type        string
	Payload     []byte
	Priority    int
	Status      JobStatus
	CreatedAt   time.Time
	StartedAt   *time.Time
	CompletedAt *time.Time
	Retries     int
	MaxRetries  int
	DependsOn   []string
	WorkerID    string
}

// ------------------------------
// Internal heap implementation
// ------------------------------

type jobHeap []*Job

func (jh jobHeap) Len() int { return len(jh) }

func (jh jobHeap) Less(i, j int) bool {
	if jh[i].Priority == jh[j].Priority {
		return jh[i].CreatedAt.Before(jh[j].CreatedAt)
	}
	return jh[i].Priority < jh[j].Priority
}

func (jh jobHeap) Swap(i, j int) { jh[i], jh[j] = jh[j], jh[i] }

func (jh *jobHeap) Push(x interface{}) {
	*jh = append(*jh, x.(*Job))
}

func (jh *jobHeap) Pop() interface{} {
	old := *jh
	n := len(old)
	item := old[n-1]
	*jh = old[0 : n-1]
	return item
}

// ------------------------------
// Thread-safe Priority Queue
// ------------------------------

type PriorityQueue struct {
	mu    sync.RWMutex
	heap  jobHeap
	index map[string]*Job
}

func NewPriorityQueue() *PriorityQueue {
	pq := &PriorityQueue{
		index: make(map[string]*Job),
	}
	heap.Init(&pq.heap)
	return pq
}

func (pq *PriorityQueue) Push(job *Job) {
	pq.mu.Lock()
	defer pq.mu.Unlock()

	if _, exists := pq.index[job.ID]; exists {
		return
	}

	heap.Push(&pq.heap, job)
	pq.index[job.ID] = job
}

func (pq *PriorityQueue) Pop() *Job {
	pq.mu.Lock()
	defer pq.mu.Unlock()

	if pq.heap.Len() == 0 {
		return nil
	}

	job := heap.Pop(&pq.heap).(*Job)
	delete(pq.index, job.ID)
	return job
}

func (pq *PriorityQueue) Peek() *Job {
	pq.mu.RLock()
	defer pq.mu.RUnlock()

	if pq.heap.Len() == 0 {
		return nil
	}
	return pq.heap[0]
}

func (pq *PriorityQueue) Get(jobID string) *Job {
	pq.mu.RLock()
	defer pq.mu.RUnlock()
	return pq.index[jobID]
}

func (pq *PriorityQueue) UpdateStatus(jobID string, status JobStatus) {
	pq.mu.Lock()
	defer pq.mu.Unlock()

	if job, exists := pq.index[jobID]; exists {
		job.Status = status
	}
}

func (pq *PriorityQueue) Len() int {
	pq.mu.RLock()
	defer pq.mu.RUnlock()
	return pq.heap.Len()
}

