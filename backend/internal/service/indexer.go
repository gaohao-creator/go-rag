package service

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"

	domainmodel "github.com/gaohao-creator/go-rag/internal/domain/model"
	domainrepo "github.com/gaohao-creator/go-rag/internal/domain/repository"
	domainservice "github.com/gaohao-creator/go-rag/internal/domain/service"
	apprag "github.com/gaohao-creator/go-rag/internal/rag"
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
	rag          *apprag.RAG
}

func NewIndexerService(
	documentRepo domainrepo.DocumentRepository,
	chunkRepo domainrepo.ChunkRepository,
	rag *apprag.RAG,
) *IndexerService {
	return &IndexerService{
		documentRepo: documentRepo,
		chunkRepo:    chunkRepo,
		rag:          rag,
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

	indexedChunks, err := s.rag.Index(ctx, domainservice.IndexRequest{
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
	if err = s.writeVectorChunks(ctx, in.KnowledgeName, chunks); err != nil {
		log.Printf("vector write skipped after failure: %v", err)
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

func (s *IndexerService) writeVectorChunks(
	ctx context.Context,
	knowledgeName string,
	chunks []domainmodel.Chunk,
) error {
	if len(chunks) == 0 {
		return nil
	}
	return s.rag.Store(ctx, domainservice.ChunkStoreRequest{
		KnowledgeName: knowledgeName,
		Chunks:        chunks,
	})
}
