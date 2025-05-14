package interfaces

import (
	"context"
	"github.com/google/uuid"
)

type KeyManagerRepo interface {
	GetRandomKey(ctx context.Context) string
	GetKeyByID(ctx context.Context, id uuid.UUID) string
}
