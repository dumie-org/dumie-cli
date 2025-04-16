/*
Copyright © 2025 Chanhyeok Seo chanhyeok.seo2@gmail.com
*/
package common

import "time"

const (
	// MaxRetries is the maximum number of retries for AWS operations
	MaxRetries = 300

	// RetryDelay is the delay between retries
	RetryDelay = 1 * time.Second

	// StatusUpdateInterval is the interval for showing status updates
	StatusUpdateInterval = 5
)
