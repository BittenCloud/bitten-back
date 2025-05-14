package interfaces

import "net/http"

type HttpRouter interface {
	CollectRoutes() error
	GetHandler() http.Handler
}
