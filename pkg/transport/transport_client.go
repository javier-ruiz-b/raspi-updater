package transport

import (
	"net/http"
)

type transportClient interface {
	Close()
	Get(url string) (*http.Response, error)
}
