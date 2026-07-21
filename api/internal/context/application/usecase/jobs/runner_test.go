package jobs

import (
	"context"
	"errors"
	"testing"
	"time"

	"repin/internal/context/domain"
)

type retryCall struct {
	id    int64
	runAt time.Time
	cause string
}

type buryCall struct {
	id    int64
	cause string
}

type fakeQueue struct {
	completed []int64
	retried   []retryCall
	buried    []buryCall
}

func (q *fakeQueue) Claim(context.Context, []domain.JobKind) (*domain.Job, error) {
	return nil, domain.ErrNoJobs
}

func (q *fakeQueue) Complete(_ context.Context, id int64) error {
	q.completed = append(q.completed, id)
	return nil
}

func (q *fakeQueue) Retry(_ context.Context, id int64, runAt time.Time, cause string) error {
	q.retried = append(q.retried, retryCall{id: id, runAt: runAt, cause: cause})
	return nil
}

func (q *fakeQueue) Bury(_ context.Context, id int64, cause string) error {
	q.buried = append(q.buried, buryCall{id: id, cause: cause})
	return nil
}

func (q *fakeQueue) RequeueStale(context.Context, time.Duration) (int64, error) {
	return 0, nil
}

const testKind domain.JobKind = "test.job"

func newTestRunner(q queue) *Runner {
	return NewRunner(q, RunnerConfig{
		RetryBackoff:    30 * time.Second,
		RetryBackoffMax: 30 * time.Minute,
	})
}

func TestBackoffDoublesUpToCap(t *testing.T) {
	t.Parallel()

	r := newTestRunner(&fakeQueue{})

	tests := []struct {
		attempts int
		want     time.Duration
	}{
		{0, 30 * time.Second},
		{1, 30 * time.Second},
		{2, time.Minute},
		{3, 2 * time.Minute},
		{6, 16 * time.Minute},
		{7, 30 * time.Minute},  // 32m hits the cap
		{50, 30 * time.Minute}, // stays capped, no overflow
	}

	for _, tt := range tests {
		if got := r.backoff(tt.attempts); got != tt.want {
			t.Errorf("backoff(%d) = %v, want %v", tt.attempts, got, tt.want)
		}
	}
}

func TestProcessCompletesSuccessfulJob(t *testing.T) {
	t.Parallel()

	q := &fakeQueue{}
	r := newTestRunner(q)
	r.Handle(testKind, func(context.Context, domain.Job) error { return nil })

	r.process(context.Background(), domain.Job{ID: 7, Kind: testKind, Attempts: 1, MaxAttempts: 5})

	if len(q.completed) != 1 || q.completed[0] != 7 {
		t.Fatalf("completed = %v, want [7]", q.completed)
	}

	if len(q.retried) != 0 || len(q.buried) != 0 {
		t.Fatalf("unexpected retries %v or burials %v", q.retried, q.buried)
	}
}

func TestProcessRetriesFailedJobWithBackoff(t *testing.T) {
	t.Parallel()

	q := &fakeQueue{}
	r := newTestRunner(q)
	r.Handle(testKind, func(context.Context, domain.Job) error { return errors.New("boom") })

	before := time.Now()
	r.process(context.Background(), domain.Job{ID: 7, Kind: testKind, Attempts: 2, MaxAttempts: 5})

	if len(q.retried) != 1 {
		t.Fatalf("retried = %v, want one call", q.retried)
	}

	call := q.retried[0]
	if call.id != 7 || call.cause != "boom" {
		t.Errorf("retry call = %+v, want id 7 cause 'boom'", call)
	}

	wantAt := before.Add(time.Minute) // backoff for the 2nd attempt
	if call.runAt.Before(wantAt) || call.runAt.After(wantAt.Add(5*time.Second)) {
		t.Errorf("runAt = %v, want ~%v", call.runAt, wantAt)
	}

	if len(q.buried) != 0 {
		t.Fatalf("unexpected burials %v", q.buried)
	}
}

func TestProcessBuriesJobAfterLastAttempt(t *testing.T) {
	t.Parallel()

	q := &fakeQueue{}
	r := newTestRunner(q)
	r.Handle(testKind, func(context.Context, domain.Job) error { return errors.New("boom") })

	r.process(context.Background(), domain.Job{ID: 7, Kind: testKind, Attempts: 5, MaxAttempts: 5})

	if len(q.buried) != 1 || q.buried[0].id != 7 || q.buried[0].cause != "boom" {
		t.Fatalf("buried = %v, want one call for job 7 with cause 'boom'", q.buried)
	}

	if len(q.retried) != 0 {
		t.Fatalf("unexpected retries %v", q.retried)
	}
}

func TestProcessBuriesJobWithoutHandler(t *testing.T) {
	t.Parallel()

	q := &fakeQueue{}
	r := newTestRunner(q)
	r.Handle(testKind, func(context.Context, domain.Job) error { return nil })

	r.process(context.Background(), domain.Job{ID: 7, Kind: "unknown.kind", Attempts: 1, MaxAttempts: 5})

	if len(q.buried) != 1 || q.buried[0].cause != "no handler registered" {
		t.Fatalf("buried = %v, want one 'no handler registered' call", q.buried)
	}
}

// A job interrupted by shutdown must go straight back to the queue, without
// consuming an attempt's backoff or being buried even on its last attempt.
func TestProcessRequeuesJobInterruptedByShutdown(t *testing.T) {
	t.Parallel()

	q := &fakeQueue{}
	r := newTestRunner(q)

	ctx, cancel := context.WithCancel(context.Background())
	r.Handle(testKind, func(context.Context, domain.Job) error {
		cancel()
		return context.Canceled
	})

	before := time.Now()
	r.process(ctx, domain.Job{ID: 7, Kind: testKind, Attempts: 5, MaxAttempts: 5})

	if len(q.retried) != 1 {
		t.Fatalf("retried = %v, want one call", q.retried)
	}

	call := q.retried[0]
	if call.cause != "interrupted by shutdown" {
		t.Errorf("cause = %q, want 'interrupted by shutdown'", call.cause)
	}

	if call.runAt.After(before.Add(5 * time.Second)) {
		t.Errorf("runAt = %v, want immediate", call.runAt)
	}

	if len(q.buried) != 0 {
		t.Fatalf("unexpected burials %v", q.buried)
	}
}
