/*
Copyright 2021 Spectro Cloud

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package maasclient

import (
	"context"
	"sync"
	"time"
)

// rateLimiter is a small token-bucket limiter. Each CAPMAAS instance has its
// own limiter so hundreds of instances do not synchronize their retries.
type rateLimiter struct {
	mu     sync.Mutex
	rate   float64 // tokens per second
	burst  float64
	tokens float64
	last   time.Time
}

func newRateLimiter(qps float64, burst int) *rateLimiter {
	if qps <= 0 {
		return nil
	}
	if burst <= 0 {
		burst = 1
	}
	return &rateLimiter{
		rate:   qps,
		burst:  float64(burst),
		tokens: float64(burst),
		last:   time.Now(),
	}
}

func (l *rateLimiter) wait(ctx context.Context) error {
	if l == nil {
		return nil
	}
	for {
		l.mu.Lock()
		now := time.Now()
		if elapsed := now.Sub(l.last).Seconds(); elapsed > 0 {
			l.tokens += elapsed * l.rate
			if l.tokens > l.burst {
				l.tokens = l.burst
			}
			l.last = now
		}
		if l.tokens >= 1 {
			l.tokens--
			l.mu.Unlock()
			return nil
		}
		wait := time.Duration((1 - l.tokens) / l.rate * float64(time.Second))
		if wait < time.Millisecond {
			wait = time.Millisecond
		}
		l.mu.Unlock()

		timer := time.NewTimer(wait)
		select {
		case <-ctx.Done():
			timer.Stop()
			return ctx.Err()
		case <-timer.C:
		}
	}
}
