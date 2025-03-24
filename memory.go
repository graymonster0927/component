package component

import (
	"context"
	"time"
)

type MemoryInterface interface {
	Del(ctx context.Context, key string) error
	Get(ctx context.Context, key string) (string, error)
	Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error
}

type MemoryCacheDefault struct{}

func (m *MemoryCacheDefault) Del(ctx context.Context, key string) error {
	return nil
}

func (m *MemoryCacheDefault) Get(ctx context.Context, key string) (string, error) {
	return "", nil
}

func (m *MemoryCacheDefault) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	return nil
}


