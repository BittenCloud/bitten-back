package interfaces

import "context"

type KeyManagerService interface {
	GetRandomKey(ctx context.Context) string
}
