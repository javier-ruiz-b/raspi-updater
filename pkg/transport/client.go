package transport

import "net/http"

type Client interface {
	Close()
	Get(url string) (*http.Response, error)
}
