package leader

import (
	"context"
	"log"

	clientv3 "go.etcd.io/etcd/client/v3"
)

// LeaseManager manages leadership lease renewal and expiration detection
type LeaseManager struct {
	Client     *clientv3.Client
	LeaseID    clientv3.LeaseID
	CancelFunc context.CancelFunc
}

// CreateLease creates a new lease with the given TTL
func (lm *LeaseManager) CreateLease(ttl int64) error {
	resp, err := lm.Client.Grant(context.Background(), ttl)
	if err != nil {
		return err
	}
	lm.LeaseID = resp.ID
	return nil
}

// KeepAlive maintains the lease periodically
func (lm *LeaseManager) KeepAlive() {
	ctx, cancel := context.WithCancel(context.Background())
	lm.CancelFunc = cancel

	ch, err := lm.Client.KeepAlive(ctx, lm.LeaseID)
	if err != nil {
		log.Printf("❌ Lease KeepAlive failed: %v", err)
		return
	}

	go func() {
		for {
			select {
			case ka, ok := <-ch:
				if !ok {
					log.Println("⚠️  Lease KeepAlive channel closed — leadership lost")
					return
				}
				log.Printf("💓 Lease %v renewed, TTL=%v\n", ka.ID, ka.TTL)
			case <-ctx.Done():
				log.Println("🛑 Lease renewal stopped")
				return
			}
		}
	}()
}

// Revoke cancels the lease when the leader voluntarily steps down
func (lm *LeaseManager) Revoke() {
	if lm.CancelFunc != nil {
		lm.CancelFunc()
	}
	_, err := lm.Client.Revoke(context.Background(), lm.LeaseID)
	if err != nil {
		log.Printf("⚠️  Failed to revoke lease: %v", err)
	}
}

