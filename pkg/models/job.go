package models

// Job represents a unit of work to be scheduled to workers.
type Job struct {
	ID       string `json:"id"`
	Priority int    `json:"priority"`
	Status   string `json:"status"`
}

