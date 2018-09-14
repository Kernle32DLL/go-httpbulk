package main

import (
	"github.com/Kernle32DLL/go-httpbulk"

	"context"
	"log"
	"sync"
	"testing"
	"time"
)

// Demonstrates, how to quickly fetch results in parallel, and synchronize with a sync.WaitGroup.
func Test(t *testing.T) {
	executor := bulk.NewExecutor()

	urls := []string{
		"https://www.google.com",
		"https://www.bing.com",
		"https://www.yahoo.com",
		"https://www.tarent.de",
	}

	wg := &sync.WaitGroup{}
	wg.Add(len(urls))

	executor.AddRequests(context.Background(), urls...)

	// Exit, when all urls have been fetched
	exitChan := make(chan struct{}, 1)
	go func() {
		wg.Wait()
		exitChan <- struct{}{}
	}()

	for {
		fin := false

		select {
		case result := <-executor.Results():
			wg.Done()
			log.Printf("%s responded with %s", result.Url(), result.Res().Status)
		case <-exitChan:
			fin = true
		case <-time.After(2 * time.Second):
			t.Fatal("Timeout...")
		}

		if fin {
			break
		}
	}
}
