package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// Scheduler Metrics
	QueueLength = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "axon_scheduler_queue_length",
		Help: "Current number of jobs in the scheduler queue",
	})

	IsLeader = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "axon_scheduler_is_leader",
		Help: "Whether this scheduler node is the current leader (1 = leader, 0 = follower)",
	})

	JobsSubmitted = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "axon_scheduler_jobs_submitted_total",
		Help: "Total number of jobs submitted to the scheduler",
	}, []string{"type"})

	JobsAssigned = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "axon_scheduler_jobs_assigned_total",
		Help: "Total number of jobs assigned to workers",
	}, []string{"worker_id"})

	JobsCompleted = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "axon_scheduler_jobs_completed_total",
		Help: "Total number of jobs completed",
	}, []string{"status"})

	JobAssignmentDuration = promauto.NewHistogram(prometheus.HistogramOpts{
		Name:    "axon_scheduler_job_assignment_duration_seconds",
		Help:    "Time taken to assign a job to a worker",
		Buckets: prometheus.DefBuckets,
	})

	JobExecutionDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "axon_scheduler_job_execution_duration_seconds",
		Help:    "Time taken to execute a job",
		Buckets: []float64{.005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5, 10, 30, 60, 120},
	}, []string{"type", "status"})

	// Worker Metrics
	WorkersRegistered = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "axon_scheduler_workers_registered",
		Help: "Number of workers registered with the scheduler",
	})

	WorkersHealthy = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "axon_scheduler_workers_healthy",
		Help: "Number of healthy workers",
	})

	WorkerHealthy = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "axon_worker_healthy",
		Help: "Health status of individual worker (1 = healthy, 0 = unhealthy)",
	}, []string{"worker_id"})

	WorkerCapacity = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "axon_worker_capacity",
		Help: "Total capacity of a worker",
	}, []string{"worker_id"})

	WorkerActiveJobs = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "axon_worker_active_jobs",
		Help: "Number of active jobs on a worker",
	}, []string{"worker_id"})

	WorkerHeartbeats = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "axon_worker_heartbeats_total",
		Help: "Total number of heartbeats received from workers",
	}, []string{"worker_id"})

	// Cache Metrics
	CacheHits = promauto.NewCounter(prometheus.CounterOpts{
		Name: "axon_cache_hits",
		Help: "Total number of cache hits",
	})

	CacheMisses = promauto.NewCounter(prometheus.CounterOpts{
		Name: "axon_cache_misses",
		Help: "Total number of cache misses",
	})

	CacheLookupDuration = promauto.NewHistogram(prometheus.HistogramOpts{
		Name:    "axon_cache_lookup_duration_seconds",
		Help:    "Time taken for cache lookup",
		Buckets: []float64{.0001, .0005, .001, .0025, .005, .01, .025, .05, .1},
	})

	// Leader Election Metrics
	LeaderElectionCampaigns = promauto.NewCounter(prometheus.CounterOpts{
		Name: "axon_leader_election_campaigns_total",
		Help: "Total number of leader election campaigns attempted",
	})

	LeaderElectionFailures = promauto.NewCounter(prometheus.CounterOpts{
		Name: "axon_leader_election_failures_total",
		Help: "Total number of leader election campaign failures",
	})

	LeadershipChanges = promauto.NewCounter(prometheus.CounterOpts{
		Name: "axon_leadership_changes_total",
		Help: "Total number of leadership changes",
	})

	// Job Retry Metrics
	JobRetries = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "axon_job_retries_total",
		Help: "Total number of job retries",
	}, []string{"job_type"})

	JobMaxRetriesExceeded = promauto.NewCounter(prometheus.CounterOpts{
		Name: "axon_job_max_retries_exceeded_total",
		Help: "Total number of jobs that exceeded max retries",
	})
)
