package batch

import (
	"context"
	"sync/atomic"
	"testing"
	"time"
)

func TestRunWorkerLimit(t *testing.T) {
	items := make([]string, 20)
	for i := range items {
		items[i] = "item"
	}

	var inFlight int64
	var maxInFlight int64

	worker := func(_ context.Context, _ string) ItemResult {
		current := atomic.AddInt64(&inFlight, 1)
		for {
			observed := atomic.LoadInt64(&maxInFlight)
			if current <= observed {
				break
			}
			if atomic.CompareAndSwapInt64(&maxInFlight, observed, current) {
				break
			}
		}
		time.Sleep(10 * time.Millisecond)
		atomic.AddInt64(&inFlight, -1)
		return ItemResult{}
	}

	Run(context.Background(), items, Config{Workers: 3, QueueSize: 4}, worker, nil)

	if maxInFlight > 3 {
		t.Fatalf("expected max in-flight <= 3, got %d", maxInFlight)
	}
}

func TestRunCancellation(t *testing.T) {
	items := make([]string, 50)
	for i := range items {
		items[i] = "item"
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	worker := func(ctx context.Context, _ string) ItemResult {
		select {
		case <-ctx.Done():
			return ItemResult{Err: ctx.Err()}
		case <-time.After(30 * time.Millisecond):
			return ItemResult{}
		}
	}

	go func() {
		time.Sleep(60 * time.Millisecond)
		cancel()
	}()

	result := Run(ctx, items, Config{Workers: 4, QueueSize: 4}, worker, nil)

	if len(result.Items) >= len(items) {
		t.Fatalf("expected cancellation to stop early, got %d items", len(result.Items))
	}
}
