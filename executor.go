package bulk

import (
	"context"
	"golang.org/x/net/context/ctxhttp"
	"net/http"
)

/**
This code has been adapted from the following code gist:
https://gist.github.com/montanaflynn/ea4b92ed640f790c4b9cee36046a5383

All kudos to Montana Flynn (montanaflynn)
*/

// Executor is the central bulk request maintainer.
type Executor struct {
	client *http.Client

	semaphoreChan chan struct{}
	resultsChan   chan Result
}

// Results offers access to the result channel.
func (e *Executor) Results() chan Result {
	return e.resultsChan
}

// Close closes the internal channels, and makes the Executor unavailable for further usage.
func (e *Executor) Close() {
	close(e.resultsChan)

	if e.semaphoreChan != nil {
		close(e.semaphoreChan)
	}
}

// NewExecutor instantiates a new Executor.
func NewExecutor(setters ...Option) Executor {
	// Default Options
	args := &Options{
		ConcurrencyLimit: 10,
		Client:           http.DefaultClient,
	}

	for _, setter := range setters {
		setter(args)
	}

	return NewSimpleExecutor(args.Client, args.ConcurrencyLimit)
}

// NewExecutor instantiates a new Executor with an http client and concurrency limit.
func NewSimpleExecutor(client *http.Client, concurrencyLimit int) Executor {
	// this buffered channel will block at the concurrency limit
	var semaphoreChan chan struct{}
	if concurrencyLimit > 0 {
		semaphoreChan = make(chan struct{}, concurrencyLimit)
	}

	// this channel will not block and collect the http request results
	resultsChan := make(chan Result)

	return Executor{
		client:        client,
		semaphoreChan: semaphoreChan,
		resultsChan:   resultsChan,
	}
}

// Issues one or more urls to be called. For each call, optional hooks for modifying the request and inspecting the result are executed (if not nil).
func (e Executor) AddRequestsWithInterceptor(
	ctx context.Context,
	modifyRequest func(r *http.Request) error,
	inspectResult func(r *Result),
	urls ...string,
) {
	for _, url := range urls {
		e.addRequestInternal(ctx, modifyRequest, inspectResult, url)
	}
}

// Issues one or more urls to be called.
func (e Executor) AddRequests(
	ctx context.Context,
	urls ...string,
) {
	e.AddRequestsWithInterceptor(ctx, nil, nil, urls...)
}

func (e Executor) addRequestInternal(
	ctx context.Context,
	modifyRequest func(r *http.Request) error,
	inspectResult func(r *Result),
	url string,
) {
	// start a go routine with the index and url in a closure
	go func(url string, ctx context.Context) {
		// If concurrency limit enabled...
		if e.semaphoreChan != nil {
			// this sends an empty struct into the semaphoreChan which
			// is basically saying add one to the limit, but when the
			// limit has been reached block until there is room
			e.semaphoreChan <- struct{}{}
		}

		var result Result
		if req, err := http.NewRequest("GET", url, nil); err == nil {
			doSend := true
			if modifyRequest != nil {
				if err := modifyRequest(req); err != nil {
					result = Result{url, nil, err}
					doSend = false
				}
			}

			if doSend {
				// send the request and put the response in a result struct
				// along with the index so we can sort them later along with
				// any error that might have occurred
				res, err := ctxhttp.Do(ctx, e.client, req)

				result = Result{url, res, err}
			}
		} else {
			result = Result{url, nil, err}
		}

		if inspectResult != nil {
			inspectResult(&result)
		}

		// now we can send the result struct through the resultsChan
		e.resultsChan <- result

		// If concurrency limit enabled...
		if e.semaphoreChan != nil {
			// once we're done it's we read from the semaphoreChan which
			// has the effect of removing one from the limit and allowing
			// another goroutine to start
			<-e.semaphoreChan
		}
	}(url, ctx)
}
