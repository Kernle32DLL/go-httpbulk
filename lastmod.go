package bulk

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"sync"
	"time"
)

// FetchLastModDatesForUrls fetches the last modification date for multiple urls at once.
func FetchLastModDatesForUrls(options []Option, modifyRequest func(r *http.Request) error, urls ...string) ([]time.Time, error) {
	if urls != nil && len(urls) > 0 {
		executor := NewExecutor(options...)

		initialContext, cancel := context.WithCancel(context.Background())

		wg := &sync.WaitGroup{}
		times := make([]time.Time, len(urls))

		for i, url := range urls {
			wg.Add(1)

			executor.AddRequestsWithInterceptor(initialContext, modifyRequest, func(r *Result) {
				lastModified, err := handleResponse(*r)
				if err != nil {
					r.SetErr(err)
				}

				times[i] = lastModified
			}, url)
		}

		go func() {
			wg.Wait()
			executor.Close()
		}()

		for {
			select {
			case <-initialContext.Done():
				return nil, initialContext.Err()
			case result, more := <-executor.Results():
				if !more {
					return times, nil
				}

				wg.Done()

				if result.Err() != nil {
					cancel()
					return nil, result.Err()
				}
			}
		}
	}

	return []time.Time{}, nil
}

func handleResponse(r Result) (time.Time, error) {
	if r.Err() != nil {
		return time.Time{}, r.Err()
	}

	defer r.Res().Body.Close()
	if _, err := ioutil.ReadAll(r.Res().Body); err != nil {
		return time.Time{}, err
	}

	if r.Res().StatusCode == 404 {
		return time.Unix(0, 0), nil
	}

	if r.Res().StatusCode != 200 && r.Res().StatusCode != 304 {
		return time.Time{}, fmt.Errorf("failed to get last-modified date for %s", r.Url())
	}

	return time.Parse(time.RFC1123, r.Res().Header.Get("last-modified"))
}
