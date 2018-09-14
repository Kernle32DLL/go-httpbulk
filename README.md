# go-httpbulk

go-httpbulk is a small wrapper lib, intended to ease the parallel loading of multiple http resources.

This is particularly useful, if you are aggregating an object from multiple resources.

Download:

```
go get github.com/Kernle32DLL/go-httpbulk
```

### Usage

First, you have to instantiate a `bulk.Executor`. This can be either done via `bulk.NewExecutor` (which takes option style parameters), or via `

```go
bulk.NewExecutor(bulk.Client(http.DefaultClient), bulk.ConcurrencyLimit(10))

// or...

bulk.NewSimpleExecutor(http.DefaultClient, 10)
```

With the executor instantiated, you can issue asynchronous requests via the `AddRequests` method:

```go
urls := []string{
    "https://www.google.com",
    "https://www.bing.com",
    "https://www.yahoo.com",
    "https://www.tarent.de",
}

executor.AddRequests(context.Background(), urls...)
```

For more control, you can use the `AddRequestsWithInterceptor` method, which allows you to both modify the request prior to sending,
as well as inspecting the request result. The former is useful for setting headers and changing the request type, the latter for
counting finished results.

```go
// Initialize a wait group with the amount of urls to call
wg := &sync.WaitGroup{}
wg.Add(len(urls))

executor.AddRequestsWithInterceptor(context.Background(), func(r *http.Request) error {
    // Change the request method to HEAD
    r.Method = "HEAD"
    return nil
}, func(r *bulk.Result) {
    defer wg.Done()

    localHash := r.Res().Header.Get("etag")
    t.Logf("%s hash %s", r.Url(), localHash)
}, urls...)
```

Two caveats for using `AddRequestsWithInterceptor`: Both the request modified and result inspector are called for EACH url. If you need to
execute some action after all requests have finished, synchronize via a `sync.WaitGroup`. Also, the result inspector is always called, and
thus useful for error handling.

If you don't need any synchronization at all, you can also use the `Results` method, which exposes the result channel.

```go
for {
    select {
    case result := <-executor.Results():
        wg.Done()
        log.Printf("%s responded with %s", result.Url(), result.Res().Status)
    }
}
```

### Status

Not yet battle tested, use with caution. On the other hand - the code is quite simple. So pick your poison.

### Thanks

This lib has been derived from the following code gist. All kudos to Montana Flynn (montanaflynn)

https://gist.github.com/montanaflynn/ea4b92ed640f790c4b9cee36046a5383

