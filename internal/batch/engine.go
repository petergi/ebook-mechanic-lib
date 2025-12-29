package batch

import (
	"context"
	"sync"
	"time"
)

type ItemResult struct {
	Path     string
	Value    interface{}
	Err      error
	Duration time.Duration
}

type ProgressUpdate struct {
	Path      string
	Completed int
	Total     int
	Err       error
	Value     interface{}
}

type ProgressFunc func(ProgressUpdate)

type Config struct {
	Workers   int
	QueueSize int
}

type Result struct {
	Items []ItemResult
}

func Run(ctx context.Context, items []string, cfg Config, worker func(context.Context, string) ItemResult, progress ProgressFunc) Result {
	workers := cfg.Workers
	if workers <= 0 {
		workers = 1
	}
	queueSize := cfg.QueueSize
	if queueSize <= 0 {
		queueSize = workers
	}

	jobs := make(chan string, queueSize)
	results := make(chan ItemResult, queueSize)

	var wg sync.WaitGroup
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				select {
				case <-ctx.Done():
					return
				case path, ok := <-jobs:
					if !ok {
						return
					}
					results <- worker(ctx, path)
				}
			}
		}()
	}

	go func() {
		defer close(jobs)
		for _, item := range items {
			select {
			case <-ctx.Done():
				return
			case jobs <- item:
			}
		}
	}()

	go func() {
		wg.Wait()
		close(results)
	}()

	result := Result{Items: make([]ItemResult, 0, len(items))}
	completed := 0

	for res := range results {
		completed++
		result.Items = append(result.Items, res)
		if progress != nil {
			progress(ProgressUpdate{
				Path:      res.Path,
				Completed: completed,
				Total:     len(items),
				Err:       res.Err,
				Value:     res.Value,
			})
		}
	}

	return result
}
