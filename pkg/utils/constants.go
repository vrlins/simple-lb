package utils

import "time"

const (
	MaxAttempts              = 3
	MaxRetry                 = 3
	HealthCheckInterval      = 2 * time.Minute
	DefaultPort              = 3000
	RetryWaitDuration        = 10 * time.Millisecond
	BackendConnectionTimeout = 2 * time.Second
)
