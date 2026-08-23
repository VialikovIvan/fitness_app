package http

import (
	"net/http"
)

func NewServer(port string, router http.Handler) *http.Server {
	server := &http.Server{
		Addr: ":"+port,
		Handler: router,
	}

	return server
}