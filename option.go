package bulk

import "net/http"

// Option is the option-wrapper for defining the workings of an Executor
type Options struct {
	ConcurrencyLimit int
	Client           *http.Client
}

type Option func(*Options)

// ConcurrencyLimit regulates how many requests will be done at the same time
// Note, you need a http client, which is actually capable of doing multiple requests
// at the same time, for this to work properly. You can use -1 to indicate no limit -
// use with care!
func ConcurrencyLimit(limit int) Option {
	return func(args *Options) {
		args.ConcurrencyLimit = limit
	}
}

// Client sets the http client, which is used for issuing requests. Per default,
// the default http client is used.
func Client(client *http.Client) Option {
	return func(args *Options) {
		args.Client = client
	}
}
