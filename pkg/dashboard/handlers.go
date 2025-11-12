package dashboard

import (
	"fmt"
	"html/template"
	"time"

	"github.com/athulya-anil/axon-scheduler/pkg/scheduler"
	"github.com/gin-gonic/gin"
)

// Dashboard provides HTTP handlers for the web UI
type Dashboard struct {
	scheduler *scheduler.Scheduler
	templates *template.Template
}

// NewDashboard creates a new dashboard instance
func NewDashboard(s *scheduler.Scheduler) (*Dashboard, error) {
	// Parse all templates
	tmpl, err := template.ParseGlob("pkg/dashboard/templates/*.html")
	if err != nil {
		return nil, err
	}

	return &Dashboard{
		scheduler: s,
		templates: tmpl,
	}, nil
}

// SetupRoutes configures dashboard routes
func (d *Dashboard) SetupRoutes(router *gin.Engine) {
	// Main dashboard pages
	router.GET("/", d.overview)
	router.GET("/dashboard", d.overview)
	router.GET("/dashboard/jobs", d.jobsPage)
	router.GET("/dashboard/workers", d.workersPage)

	// HTMX partial endpoints (return HTML fragments)
	router.GET("/api/dashboard/status", d.statusPartial)
	router.GET("/api/dashboard/jobs", d.jobsPartial)
	router.GET("/api/dashboard/workers", d.workersPartial)

	// SSE endpoints for real-time updates
	router.GET("/api/events/jobs", d.jobsSSE)
	router.GET("/api/events/workers", d.workersSSE)
	router.GET("/api/events/status", d.statusSSE)
}

// overview renders the main dashboard page
func (d *Dashboard) overview(c *gin.Context) {
	data := d.getStatusData()
	c.Header("Content-Type", "text/html; charset=utf-8")
	d.templates.ExecuteTemplate(c.Writer, "overview.html", data)
}

// jobsPage renders the jobs page
func (d *Dashboard) jobsPage(c *gin.Context) {
	data := d.getJobsData()
	c.Header("Content-Type", "text/html; charset=utf-8")
	d.templates.ExecuteTemplate(c.Writer, "jobs.html", data)
}

// workersPage renders the workers page
func (d *Dashboard) workersPage(c *gin.Context) {
	data := d.getWorkersData()
	c.Header("Content-Type", "text/html; charset=utf-8")
	d.templates.ExecuteTemplate(c.Writer, "workers.html", data)
}

// statusPartial returns HTML fragment for status updates (used by HTMX)
func (d *Dashboard) statusPartial(c *gin.Context) {
	data := d.getStatusData()
	c.Header("Content-Type", "text/html; charset=utf-8")

	leaderBadge := `<span class="inline-flex items-center px-3 py-1 rounded-full text-sm font-medium bg-gray-100 text-gray-800">Follower</span>`
	if data["is_leader"].(bool) {
		leaderBadge = `<span class="inline-flex items-center px-3 py-1 rounded-full text-sm font-medium bg-green-100 text-green-800">Leader</span>`
	}

	html := fmt.Sprintf(`
	<div id="status-panel" class="grid grid-cols-1 md:grid-cols-4 gap-4">
		<div class="bg-white rounded-lg shadow p-6">
			<div class="text-sm font-medium text-gray-500">Node ID</div>
			<div class="mt-2 text-xl font-semibold text-gray-900">%s</div>
		</div>
		<div class="bg-white rounded-lg shadow p-6">
			<div class="text-sm font-medium text-gray-500">Leader Status</div>
			<div class="mt-2">%s</div>
		</div>
		<div class="bg-white rounded-lg shadow p-6">
			<div class="text-sm font-medium text-gray-500">Queue Length</div>
			<div class="mt-2 text-3xl font-bold text-blue-600">%d</div>
		</div>
		<div class="bg-white rounded-lg shadow p-6">
			<div class="text-sm font-medium text-gray-500">Healthy Workers</div>
			<div class="mt-2 text-3xl font-bold text-green-600">%d / %d</div>
		</div>
	</div>
	`, data["node_id"].(string), leaderBadge, data["queue_length"].(int),
	   data["workers_healthy"].(int), data["workers_total"].(int))

	c.Writer.Write([]byte(html))
}

// jobsPartial returns HTML fragment for jobs list (used by HTMX)
func (d *Dashboard) jobsPartial(c *gin.Context) {
	jobs := d.scheduler.ListJobs()

	c.Header("Content-Type", "text/html; charset=utf-8")

	html := `<div id="jobs-list" class="space-y-4">`

	if len(jobs) == 0 {
		html += `<div class="text-center py-12 text-gray-500">No jobs in queue</div>`
	} else {
		for _, job := range jobs {
			statusColor := "gray"
			switch job.Status {
			case "PENDING":
				statusColor = "yellow"
			case "RUNNING":
				statusColor = "blue"
			case "COMPLETED":
				statusColor = "green"
			case "FAILED":
				statusColor = "red"
			}

			jobIDShort := job.ID
			if len(job.ID) > 8 {
				jobIDShort = job.ID[:8] + "..."
			}

			html += fmt.Sprintf(`
			<div class="bg-white rounded-lg shadow p-4 hover:shadow-md transition-shadow">
				<div class="flex items-center justify-between">
					<div class="flex-1">
						<div class="flex items-center gap-3">
							<span class="text-sm font-mono text-gray-500">%s</span>
							<span class="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium bg-%s-100 text-%s-800">
								%s
							</span>
							<span class="text-sm text-gray-600">Priority: %d</span>
						</div>
						<div class="mt-1 text-sm text-gray-500">Type: %s</div>
					</div>
					<div class="text-sm text-gray-400">%s</div>
				</div>
			</div>
			`, jobIDShort, statusColor, statusColor, string(job.Status), job.Priority, job.Type, job.CreatedAt.Format("15:04:05"))
		}
	}

	html += `</div>`

	c.Writer.Write([]byte(html))
}

// workersPartial returns HTML fragment for workers list (used by HTMX)
func (d *Dashboard) workersPartial(c *gin.Context) {
	workers := d.scheduler.ListWorkers()

	c.Header("Content-Type", "text/html; charset=utf-8")

	html := `<div id="workers-list" class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">`

	if len(workers) == 0 {
		html += `<div class="col-span-full text-center py-12 text-gray-500">No workers registered</div>`
	} else {
		for _, worker := range workers {
			healthColor := "red"
			healthText := "Unhealthy"
			if worker.Healthy {
				healthColor = "green"
				healthText = "Healthy"
			}

			timeSinceHeartbeat := time.Since(worker.LastHeartbeat).Round(time.Second)

			html += fmt.Sprintf(`
			<div class="bg-white rounded-lg shadow p-4">
				<div class="flex items-center justify-between mb-3">
					<span class="text-sm font-medium text-gray-900">%s</span>
					<span class="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium bg-%s-100 text-%s-800">
						%s
					</span>
				</div>
				<div class="space-y-2 text-sm text-gray-600">
					<div class="flex justify-between">
						<span>Address:</span>
						<span class="font-mono text-xs">%s</span>
					</div>
					<div class="flex justify-between">
						<span>Load:</span>
						<span class="font-semibold">%d/%d</span>
					</div>
					<div class="flex justify-between">
						<span>Last heartbeat:</span>
						<span class="text-xs">%s ago</span>
					</div>
				</div>
			</div>
			`, worker.ID, healthColor, healthColor, healthText, worker.Address,
			   len(worker.CurrentJobs), worker.Capacity, timeSinceHeartbeat.String())
		}
	}

	html += `</div>`

	c.Writer.Write([]byte(html))
}

// Helper functions to get data
func (d *Dashboard) getStatusData() map[string]interface{} {
	workers := d.scheduler.ListWorkers()
	healthyCount := 0
	for _, w := range workers {
		if w.Healthy {
			healthyCount++
		}
	}

	return map[string]interface{}{
		"node_id":         d.scheduler.NodeID,
		"is_leader":       d.scheduler.IsLeader,
		"queue_length":    d.scheduler.GetQueueLength(),
		"workers_total":   len(workers),
		"workers_healthy": healthyCount,
		"timestamp":       time.Now().Format("15:04:05"),
	}
}

func (d *Dashboard) getJobsData() map[string]interface{} {
	jobs := d.scheduler.ListJobs()

	return map[string]interface{}{
		"jobs":  jobs,
		"count": len(jobs),
	}
}

func (d *Dashboard) getWorkersData() map[string]interface{} {
	workers := d.scheduler.ListWorkers()

	return map[string]interface{}{
		"workers": workers,
		"count":   len(workers),
	}
}
