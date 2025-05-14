package services

import (
	"bitback/internal/interfaces"
	"context"
)

type KeyManagerService struct {
	keyManagerRepo interfaces.KeyManagerRepo
}

func NewKeyManager(repo interfaces.KeyManagerRepo) *KeyManagerService {
	return &KeyManagerService{
		keyManagerRepo: repo,
	}
}

func (k *KeyManagerService) GetRandomKey(ctx context.Context) string {
	return ""
}
