/*
Copyright © 2025 Chanhyeok Seo chanhyeok.seo2@gmail.com
*/
package common

import (
	"context"
	"fmt"
	"time"
)

type StatusChecker interface {
	CheckStatus(ctx context.Context) (string, error)
	IsTargetStatus(currentStatus string) bool
	IsErrorStatus(currentStatus string) bool
	GetResourceID() string
	GetResourceType() string
}

func WaitForResourceStatus(ctx context.Context, checker StatusChecker) error {
	fmt.Printf("Waiting for %s %s to reach target status...\n", checker.GetResourceType(), checker.GetResourceID())
	startTime := time.Now()
	lastStatusUpdate := time.Now()

	for i := 0; i < MaxRetries; i++ {
		status, err := checker.CheckStatus(ctx)
		if err != nil {
			return fmt.Errorf("error checking status: %w", err)
		}

		if checker.IsTargetStatus(status) {
			elapsed := time.Since(startTime)
			fmt.Printf("%s %s has reached target status (waited %s)\n",
				checker.GetResourceType(),
				checker.GetResourceID(),
				elapsed.Round(time.Second))
			return nil
		}

		if checker.IsErrorStatus(status) {
			elapsed := time.Since(startTime)
			return fmt.Errorf("%s %s is in error state %s (waited %s)",
				checker.GetResourceType(),
				checker.GetResourceID(),
				status,
				elapsed.Round(time.Second))
		}

		if time.Since(lastStatusUpdate) >= StatusUpdateInterval*time.Second {
			elapsed := time.Since(startTime)
			fmt.Printf("Current status: %s, waiting... (elapsed: %s)\n",
				status,
				elapsed.Round(time.Second))
			lastStatusUpdate = time.Now()
		}

		time.Sleep(RetryDelay)
	}

	elapsed := time.Since(startTime)
	return fmt.Errorf("timeout waiting for %s %s to reach target status (waited %s)",
		checker.GetResourceType(),
		checker.GetResourceID(),
		elapsed.Round(time.Second))
}
