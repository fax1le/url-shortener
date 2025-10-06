package storage

import (
	"context"
	"time"
)

type Database interface {
	StoreUrl(context.Context, string, string, time.Duration) error
	StoreClicks(context.Context, string, ...any) error
	SlugExists(context.Context, string) (bool, error)
	GetUrl(context.Context, string) (string, error)
	CleanUp(context.Context) error
	ExpireUrls(context.Context) (int64, error)
	Close() error
}

type Cache interface {
	GetUrl(context.Context, string) (string, error)
	HashGetAll(context.Context, string) (map[string]string, error)
	StoreUrl(context.Context, string, string) error
	CleanUp(context.Context) error
	GetIP(context.Context, string) (map[string]string, error)
	StoreIPLimit(context.Context, string, float64, float64) error
	Increment(context.Context, string, int64) error
	IncrementBatch(context.Context, string, map[string]int64, int) error
	Delete(context.Context, string) error
	RenameSetTTL(context.Context, string, string, time.Duration) error
	Close() error
}
