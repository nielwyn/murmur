package feedfetch

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/nielwyn/murmur/internal/database"
)

const (
	defaultWorkers      = 5
	defaultInterval     = time.Minute
	defaultBatchSize    = 10
	defaultFetchTimeout = 10 * time.Second
)

// Config controls the scheduler's behavior. Zero values fall back to defaults.
type Config struct {
	Workers int
	// Interval is how often the scheduler checks for due feeds. Individual
	// feeds are still only fetched per their own fetch_interval_seconds.
	Interval     time.Duration
	BatchSize    int32
	FetchTimeout time.Duration
}

func (c Config) withDefaults() Config {
	if c.Workers <= 0 {
		c.Workers = defaultWorkers
	}
	if c.Interval <= 0 {
		c.Interval = defaultInterval
	}
	if c.BatchSize <= 0 {
		c.BatchSize = defaultBatchSize
	}
	if c.FetchTimeout <= 0 {
		c.FetchTimeout = defaultFetchTimeout
	}
	return c
}

// Scheduler periodically fetches due feeds using a bounded worker pool
// (fan-out) and a single collector goroutine (fan-in).
type Scheduler struct {
	db  *database.Queries
	cfg Config
}

func NewScheduler(db *database.Queries, cfg Config) *Scheduler {
	return &Scheduler{db: db, cfg: cfg.withDefaults()}
}

// Run starts the scheduler and blocks until ctx is canceled. On each tick it
// queries due feeds and fans them out to s.cfg.Workers workers; results are
// fanned back in and logged by a single collector goroutine.
func (s *Scheduler) Run(ctx context.Context) {
	jobs := make(chan FetchJob)
	results := make(chan FetchResult)

	var workers sync.WaitGroup
	for range s.cfg.Workers {
		workers.Go(func() {
			s.worker(ctx, jobs, results)
		})
	}

	collectorDone := make(chan struct{})
	go func() {
		defer close(collectorDone)
		for result := range results {
			logResult(result)
		}
	}()

	ticker := time.NewTicker(s.cfg.Interval)
	defer ticker.Stop()

	// Run an immediate tick on startup so feeds don't wait a full interval
	// before the first fetch.
	s.tick(ctx, jobs)

	for {
		select {
		case <-ctx.Done():
			close(jobs)
			workers.Wait()
			close(results)
			<-collectorDone
			return
		case <-ticker.C:
			s.tick(ctx, jobs)
		}
	}
}

// tick queries due feeds and sends one job per feed into jobs.
func (s *Scheduler) tick(ctx context.Context, jobs chan<- FetchJob) {
	feeds, err := s.db.GetFeedsDueForFetch(ctx, s.cfg.BatchSize)
	if err != nil {
		log.Printf("feedfetch: querying due feeds: %v", err)
		return
	}
	if len(feeds) == 0 {
		return
	}

	log.Printf("feedfetch: %d feed(s) due", len(feeds))
	for _, feed := range feeds {
		select {
		case jobs <- FetchJob{Feed: feed}:
		case <-ctx.Done():
			return
		}
	}
}

func logResult(r FetchResult) {
	if r.Err != nil {
		log.Printf("feedfetch: %s: error: %v (%s)", r.Feed.Name, r.Err, r.Duration)
		return
	}
	log.Printf("feedfetch: %s: %d new post(s) (%s)", r.Feed.Name, r.NewPosts, r.Duration)
}
