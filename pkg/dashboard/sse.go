package dashboard

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
)

// jobsSSE streams job updates via Server-Sent Events
func (d *Dashboard) jobsSSE(c *gin.Context) {
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("Access-Control-Allow-Origin", "*")

	// Create a channel for client disconnection
	clientGone := c.Request.Context().Done()

	// Send updates every 2 seconds
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-clientGone:
			return
		case <-ticker.C:
			jobs := d.scheduler.ListJobs()

			// Convert jobs to JSON
			jobsJSON, err := json.Marshal(jobs)
			if err != nil {
				continue
			}

			// Send SSE event
			fmt.Fprintf(c.Writer, "event: jobs\n")
			fmt.Fprintf(c.Writer, "data: %s\n\n", jobsJSON)
			c.Writer.Flush()
		}
	}
}

// workersSSE streams worker updates via Server-Sent Events
func (d *Dashboard) workersSSE(c *gin.Context) {
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("Access-Control-Allow-Origin", "*")

	// Create a channel for client disconnection
	clientGone := c.Request.Context().Done()

	// Send updates every 2 seconds
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-clientGone:
			return
		case <-ticker.C:
			workers := d.scheduler.ListWorkers()

			// Convert workers to JSON
			workersJSON, err := json.Marshal(workers)
			if err != nil {
				continue
			}

			// Send SSE event
			fmt.Fprintf(c.Writer, "event: workers\n")
			fmt.Fprintf(c.Writer, "data: %s\n\n", workersJSON)
			c.Writer.Flush()
		}
	}
}

// statusSSE streams cluster status updates via Server-Sent Events
func (d *Dashboard) statusSSE(c *gin.Context) {
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("Access-Control-Allow-Origin", "*")

	// Create a channel for client disconnection
	clientGone := c.Request.Context().Done()

	// Send updates every 2 seconds
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-clientGone:
			return
		case <-ticker.C:
			status := d.getStatusData()

			// Convert status to JSON
			statusJSON, err := json.Marshal(status)
			if err != nil {
				continue
			}

			// Send SSE event
			fmt.Fprintf(c.Writer, "event: status\n")
			fmt.Fprintf(c.Writer, "data: %s\n\n", statusJSON)
			c.Writer.Flush()
		}
	}
}
