package service

import "context"

// Attempt tells which processing attempt of a job runs now.
type Attempt struct {
	// Number is the number of this attempt. The first attempt is 1.
	Number int
	// Max is the maximum number of attempts. 0 means no limit.
	Max int
	// IsFinal is true if the queue will not retry the job when this attempt fails.
	IsFinal bool
}

type attemptCtxKey struct{}

// ContextWithAttempt returns a copy of ctx that holds attempt.
// The jobs queue uses it to tell the service which attempt runs.
func ContextWithAttempt(ctx context.Context, attempt Attempt) context.Context {
	return context.WithValue(ctx, attemptCtxKey{}, attempt)
}

// AttemptFromContext returns the attempt from ctx. If ctx has no attempt,
// it returns attempt 1 of an unknown maximum that is not final.
func AttemptFromContext(ctx context.Context) Attempt {
	if a, ok := ctx.Value(attemptCtxKey{}).(Attempt); ok {
		return a
	}
	return Attempt{Number: 1}
}
