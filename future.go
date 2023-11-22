// Package atime parses dates, times and ranges without requiring a format.
package atime

import (
	"errors"
	"sync"
	"time"

	dps "github.com/markusmobius/go-dateparser"
)

// ParseFutureTime parses the time string and returns the nearest future time.
func ParseFutureTime(now time.Time, raw string) (time.Time, error) {
	// use two parsers to parse the time
	var (
		wg      sync.WaitGroup
		t1, t2  time.Time
		e1, e2  error
		errPast = errors.New("time is in the past")
	)
	wg.Add(2)

	// use anytime to parse the time and check if it's in the future
	go func() {
		defer wg.Done()

		t1, e1 = Parse(raw, now, DefaultToFuture)
		if e1 == nil && t1.Before(now) {
			e1 = errPast
		}
	}()

	// use dateparser to parse the time and check if it's in the future
	go func() {
		defer wg.Done()

		cfg := &dps.Configuration{
			CurrentTime:         now,
			PreferredDateSource: dps.Future,
		}
		if dt, err := dps.Parse(cfg, raw); err != nil {
			e2 = err
		} else if dt.Time.Before(now) {
			e2 = errPast
		} else {
			t2 = dt.Time
		}
	}()

	// check result
	wg.Wait()
	if e1 != nil && e2 != nil {
		return t1, e1
	}

	// pick the result
	if e1 == nil && e2 == nil {
		// both parser succeed, use the one that returns the nearest future
		if t1.Sub(now) < t2.Sub(now) {
			return t1, nil
		}
		return t2, nil
	} else if e1 == nil {
		// only anytime succeed
		return t1, nil
	} else {
		// only dateparser succeed
		return t2, nil
	}
}
