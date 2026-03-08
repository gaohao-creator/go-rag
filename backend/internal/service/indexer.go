package service

import (
	"context"
	"errors"
	"fmt"
	"strings"

	domainmodel "github.com/gaohao-creator/go-rag/internal/domain/model"
	domainrepo "github.com/gaohao-creator/go-rag/internal/domain/repository"
	domainservice "github.com/gaohao-creator/go-rag/internal/domain/service"
)

const (
	DocumentStatusPending  = 0
	DocumentStatusIndexing = 1
	DocumentStatusActive   = 2
	DocumentStatusFailed   = 3
)

type IndexInput struct {
	URI           string
	KnowledgeName string
	FileName      string
}

type IndexerService struct {
	documentRepo domainrepo.DocumentRepository
	chunkRepo    domainrepo.ChunkRepository
	engine       domainservice.Indexer
}

func NewIndexerService(documentRepo domainrepo.DocumentRepository, chunkRepo domainrepo.ChunkRepository, engine domainservice.Indexer) *IndexerService {
	return &IndexerService{
		documentRepo: documentRepo,
		chunkRepo:    chunkRepo,
		engine:       engine,
	}
}

func (s *IndexerService) Index(ctx context.Context, in IndexInput) ([]string, error) {
	if strings.TrimSpace(in.URI) == "" {
		return nil, errors.New("文档地址不能为空")
	}
	if strings.TrimSpace(in.KnowledgeName) == "" {
		return nil, errors.New("知识库名称不能为空")
	}
	if strings.TrimSpace(in.FileName) == "" {
		return nil, errors.New("文件名不能为空")
	}

	documentID, err := s.documentRepo.Create(ctx, domainmodel.Document{
		KnowledgeBaseName: in.KnowledgeName,
		FileName:          in.FileName,
		Status:            DocumentStatusPending,
	})
	if err != nil {
		return nil, err
	}

	indexedChunks, err := s.engine.Index(ctx, domainservice.IndexRequest{
		URI:           in.URI,
		KnowledgeName: in.KnowledgeName,
	})
	if err != nil {
		return nil, s.markDocumentFailed(ctx, documentID, err)
	}

	chunks := make([]domainmodel.Chunk, 0, len(indexedChunks))
	ids := make([]string, 0, len(indexedChunks))
	for _, chunk := range indexedChunks {
		chunks = append(chunks, domainmodel.Chunk{
			KnowledgeDocID: documentID,
			ChunkID:        chunk.ChunkID,
			Content:        chunk.Content,
			Ext:            chunk.Ext,
			Status:         DocumentStatusActive,
		})
		ids = append(ids, chunk.ChunkID)
	}

	if err = s.chunkRepo.BatchCreate(ctx, chunks); err != nil {
		return nil, s.markDocumentFailed(ctx, documentID, err)
	}
	if err = s.documentRepo.UpdateStatus(ctx, documentID, DocumentStatusActive); err != nil {
		return nil, fmt.Errorf("更新文档状态失败: %w", err)
	}
	return ids, nil
}

func (s *IndexerService) markDocumentFailed(ctx context.Context, documentID int64, sourceErr error) error {
	if err := s.documentRepo.UpdateStatus(ctx, documentID, DocumentStatusFailed); err != nil {
		return fmt.Errorf("%w; 更新失败状态也失败: %v", sourceErr, err)
	}
	return sourceErr
}
