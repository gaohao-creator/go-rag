package service

import (
	"context"

	domainmodel "github.com/gaohao-creator/go-rag/internal/domain/model"
)

type ChatGenerateInput struct {
	ConvID     string
	Question   string
	History    []domainmodel.Message
	References []domainmodel.RetrievedChunk
}

type ChatModel interface {
	Generate(ctx context.Context, in ChatGenerateInput) (string, error)
}

type ChatStreamModel interface {
	// todo 不是常规的
	GenerateStream(ctx context.Context, in ChatGenerateInput) ([]string, error)
}
