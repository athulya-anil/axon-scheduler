package main

import (
    "log"
    "os"

    "github.com/athulya-anil/axon-scheduler/pkg/leader"
)

func main() {
    nodeID, _ := os.Hostname()
    log.Printf("🚀 Axon Scheduler starting on node %s...", nodeID)

    leader.ElectLeader(nodeID)
}

