package web

import (
	"context"
	"github.com/google/uuid"
)

type ctxKey int

const (
	writerKey ctxKey = iota + 1
	traceIDKey
)

// SetTraceID 为请求设置一个唯一的 traceID
func SetTraceID(ctx context.Context, traceID uuid.UUID) context.Context {
	return context.WithValue(ctx, traceIDKey, traceID)
}

// GetTraceID 返回请求的 traceID
func GetTraceID(ctx context.Context) uuid.UUID {
	v, ok := ctx.Value(traceIDKey).(uuid.UUID)
	if !ok {
		return uuid.UUID{}
	}

	return v
}
