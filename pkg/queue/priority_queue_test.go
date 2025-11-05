package queue

import (
	"testing"
	"time"
)

func TestPriorityQueueOrdering(t *testing.T) {
	pq := NewPriorityQueue()

	job1 := &Job{ID: "1", Priority: 5, CreatedAt: time.Now()}
	job2 := &Job{ID: "2", Priority: 1, CreatedAt: time.Now()}
	job3 := &Job{ID: "3", Priority: 3, CreatedAt: time.Now()}

	pq.Push(job1)
	pq.Push(job2)
	pq.Push(job3)

	if pq.Pop().ID != "2" {
		t.Error("Expected job2 to be popped first (highest priority)")
	}
	if pq.Pop().ID != "3" {
		t.Error("Expected job3 to be popped second")
	}
	if pq.Pop().ID != "1" {
		t.Error("Expected job1 to be popped last")
	}
}

