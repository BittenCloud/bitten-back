package interfaces

import "context"

type ApiServer interface {
	CreateAndPrepare() ApiServer
	Run() error
	Shutdown(ctx context.Context) error
}
