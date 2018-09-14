package main

import (
	"github.com/Kernle32DLL/go-httpbulk"

	"context"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"sync"
	"testing"
)

// Demonstrates, how to quickly calculate the hash of multiple resources.
func Test(t *testing.T) {
	executor := bulk.NewExecutor()

	urls := []string{
		"https://gobyexample.com",
		"https://reactjs.org",
		"https://www.tarent.de",
	}

	wg := &sync.WaitGroup{}
	wg.Add(len(urls))

	hashes := make([]string, len(urls))

	executor.AddRequestsWithInterceptor(context.Background(), func(r *http.Request) error {
		// Change the request method to head - we don't need the body
		r.Method = "HEAD"
		return nil
	}, func(r *bulk.Result) {
		localHash := r.Res().Header.Get("etag")

		t.Logf("%s hash %s", r.Url(), localHash)

		for index, url := range urls {
			if url == r.Url() {
				hashes[index] = localHash
			}
		}
	}, urls...)

	closer := make(chan struct{})
	defer close(closer)
	go func() {
		for {
			select {
			case <-executor.Results():
				wg.Done()
			case <-closer:
				return
			}
		}
	}()

	// Wait for the results
	wg.Wait()
	closer <- struct{}{}

	h := sha256.New()
	for _, value := range hashes {
		h.Write([]byte(value))
	}
	hash := h.Sum(nil)

	t.Logf("Final hash: %s", hex.EncodeToString(hash[:]))
}
