package leader

import (
    "context"
    "log"
    "time"

    clientv3 "go.etcd.io/etcd/client/v3"
    "go.etcd.io/etcd/client/v3/concurrency"
)

// ElectLeader connects to etcd and participates in leader election
func ElectLeader(id string) {
    cli, err := clientv3.New(clientv3.Config{
        Endpoints:   []string{"127.0.0.1:4001"},
        DialTimeout: 5 * time.Second,
    })
    if err != nil {
        log.Fatalf("❌ failed to connect to etcd: %v", err)
    }
    defer cli.Close()

    // Create a new session for election (with TTL 10s)
    session, err := concurrency.NewSession(cli, concurrency.WithTTL(10))
    if err != nil {
        log.Fatalf("❌ failed to create session: %v", err)
    }
    defer session.Close()

    election := concurrency.NewElection(session, "/axon-leader-election")

    ctx := context.Background()

    // Campaign to become leader
    if err := election.Campaign(ctx, id); err != nil {
        log.Fatalf("❌ election campaign failed: %v", err)
    }

    log.Printf("🏆 Node %s has become the leader!", id)

    // Keep leadership alive while session is active
    for {
        select {
        case <-session.Done():
            log.Println("⚠️ leadership lost — session expired")
            return
        default:
            time.Sleep(2 * time.Second)
        }
    }
}

