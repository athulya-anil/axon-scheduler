# Axon Scheduler Dashboard

A real-time web dashboard for monitoring and managing the Axon distributed job scheduler cluster.

## Features

### 1. **Overview Page** (`/dashboard`)
The main dashboard provides a comprehensive view of your cluster:

- **Cluster Status Panel**
  - Current node ID
  - Leader/Follower status with color-coded badges
  - Total jobs in queue
  - Healthy workers count

- **Job Submission Form**
  - Submit jobs directly from the web UI
  - Configure job type, priority, max retries, and payload
  - Inline success/error feedback

- **Recent Jobs List**
  - Quick view of recently submitted jobs
  - Job status with color coding
  - Priority and creation time

- **Active Workers Grid**
  - Live worker health status
  - Current load per worker
  - Last heartbeat timestamps

### 2. **Jobs Page** (`/dashboard/jobs`)
Detailed job queue monitoring:

- Live job list with auto-refresh (every 2 seconds)
- Job statistics breakdown (pending, running, completed, failed)
- Job details: ID, type, priority, status, creation time
- Color-coded status badges for easy visual scanning

### 3. **Workers Page** (`/dashboard/workers`)
Worker pool management:

- Worker health monitoring grid
- Detailed worker information:
  - Worker ID and gRPC address
  - Current load (active jobs / capacity)
  - Last heartbeat timing
  - Health status (Healthy/Unhealthy)
- Worker statistics summary
- Registration instructions

## Technology Stack

- **HTMX** - Dynamic content updates without JavaScript
- **Tailwind CSS** - Modern, responsive styling via CDN
- **Server-Sent Events (SSE)** - Real-time push updates (optional)
- **Go html/template** - Server-side template rendering
- **Gin Framework** - HTTP routing and handling

## Architecture

### HTMX Polling
The dashboard uses HTMX's polling mechanism for automatic updates:

```html
<div hx-get="/api/dashboard/status"
     hx-trigger="load, every 2s"
     hx-swap="innerHTML">
</div>
```

This makes HTTP requests every 2 seconds to fetch updated content, providing near real-time monitoring without complex WebSocket management.

### API Endpoints

#### Dashboard Pages (Full HTML)
- `GET /dashboard` - Overview page
- `GET /dashboard/jobs` - Jobs page
- `GET /dashboard/workers` - Workers page

#### Partial Endpoints (HTML Fragments for HTMX)
- `GET /api/dashboard/status` - Cluster status panel
- `GET /api/dashboard/jobs` - Jobs list
- `GET /api/dashboard/workers` - Workers grid

#### SSE Endpoints (Optional Real-time Streams)
- `GET /api/events/status` - Status updates stream
- `GET /api/events/jobs` - Jobs updates stream
- `GET /api/events/workers` - Workers updates stream

## Quick Start

### 1. Start etcd
```bash
etcd
```

### 2. Start the scheduler with dashboard
```bash
./scheduler
# Dashboard available at: http://localhost:8080/dashboard
```

### 3. Start workers
```bash
# Terminal 1
WORKER_ID=worker-1 WORKER_PORT=50051 ./worker

# Terminal 2
WORKER_ID=worker-2 WORKER_PORT=50052 ./worker
```

### 4. Open dashboard
Navigate to http://localhost:8080/dashboard in your browser

### 5. Test the dashboard
```bash
./scripts/test-dashboard.sh
```

## Development

### Project Structure
```
pkg/dashboard/
├── handlers.go          # HTTP handlers and HTML generation
├── sse.go              # Server-Sent Events implementation
└── templates/
    ├── layout.html     # Base layout with navigation
    ├── overview.html   # Main dashboard page
    ├── jobs.html       # Jobs monitoring page
    └── workers.html    # Workers monitoring page
```

### Adding New Features

#### 1. Add a new page
Create template in `pkg/dashboard/templates/`:
```html
{{define "content"}}
<div class="space-y-6">
    <!-- Your content here -->
</div>
{{end}}
```

#### 2. Add handler
In `pkg/dashboard/handlers.go`:
```go
func (d *Dashboard) myNewPage(c *gin.Context) {
    data := d.getMyData()
    c.Header("Content-Type", "text/html; charset=utf-8")
    d.templates.ExecuteTemplate(c.Writer, "mypage.html", data)
}
```

#### 3. Register route
In `SetupRoutes()`:
```go
router.GET("/dashboard/mypage", d.myNewPage)
```

### Customizing Update Intervals

Change polling frequency in templates:
```html
<!-- 5 second updates -->
<div hx-get="/api/dashboard/status"
     hx-trigger="load, every 5s">
</div>

<!-- 1 second updates (more aggressive) -->
<div hx-get="/api/dashboard/jobs"
     hx-trigger="load, every 1s">
</div>
```

## Features in Detail

### Real-time Updates
- **Automatic Polling**: HTMX polls endpoints every 2-3 seconds
- **Visual Indicators**: Pulsing green dot shows live monitoring
- **Smooth Transitions**: CSS animations for loading states
- **Efficient Updates**: Only updated sections are replaced

### Responsive Design
- Mobile-friendly grid layouts
- Adaptive column counts (1/2/3 columns based on screen size)
- Touch-friendly interface

### Visual Feedback
- Color-coded status badges:
  - 🟡 Yellow: Pending
  - 🔵 Blue: Running
  - 🟢 Green: Completed/Healthy
  - 🔴 Red: Failed/Unhealthy
- Loading states with skeleton screens
- Hover effects on interactive elements

### Job Submission
Submit jobs directly from the dashboard with:
- Job type specification
- Priority setting (0-10)
- Max retries configuration
- JSON payload editor
- Instant feedback on submission

## Browser Compatibility

Tested and working on:
- Chrome 90+
- Firefox 88+
- Safari 14+
- Edge 90+

## Performance Considerations

- **Template Caching**: Templates are parsed once at startup
- **Efficient Rendering**: Only sends HTML for changed content
- **Connection Pooling**: HTTP/1.1 keep-alive for polling requests
- **Minimal JavaScript**: HTMX library (~14KB gzipped)
- **CDN Resources**: Tailwind CSS loaded from CDN (cached by browser)

## Troubleshooting

### Dashboard not loading
1. Check scheduler is running: `curl http://localhost:8080/health`
2. Verify port 8080 is accessible
3. Check logs for template parsing errors

### No real-time updates
1. Check HTMX is loading (browser dev tools → Network)
2. Verify API endpoints return data: `curl http://localhost:8080/api/dashboard/status`
3. Check browser console for errors

### Workers not showing
1. Ensure workers are started and sending heartbeats
2. Check worker registration: `curl http://localhost:8080/workers`
3. Verify etcd is running and accessible

## Future Enhancements

Potential improvements for Phase 4:

- **Advanced Filtering**: Filter jobs by status, priority, type
- **Search**: Search jobs by ID or payload content
- **Pagination**: Handle large job queues efficiently
- **Historical Data**: Charts showing job throughput over time
- **Cache Metrics**: Display semantic cache hit rates and savings
- **Log Streaming**: View job execution logs in real-time
- **Manual Actions**: Cancel jobs, retry failed jobs, pause workers
- **Alerts**: Visual alerts for unhealthy workers or failed jobs
- **Dark Mode**: Toggle between light and dark themes
- **Export**: Download job data as CSV/JSON

## Credits

Built with:
- [HTMX](https://htmx.org/) - High power tools for HTML
- [Tailwind CSS](https://tailwindcss.com/) - Utility-first CSS framework
- [Gin](https://gin-gonic.com/) - Go web framework
