package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	domainmodel "github.com/gaohao-creator/go-rag/internal/domain/model"
	"github.com/gaohao-creator/go-rag/internal/service"
	webmiddleware "github.com/gaohao-creator/go-rag/internal/web/middleware"
	"github.com/gin-gonic/gin"
)

type KnowledgeBaseCreator interface {
	Create(ctx context.Context, in service.CreateKnowledgeBaseInput) (int64, error)
	List(ctx context.Context, in service.ListKnowledgeBasesInput) ([]domainmodel.KnowledgeBase, error)
	GetByID(ctx context.Context, id int64) (*domainmodel.KnowledgeBase, error)
	Update(ctx context.Context, in service.UpdateKnowledgeBaseInput) error
	Delete(ctx context.Context, id int64) error
}

type DocumentLister interface {
	ListByKnowledgeBase(ctx context.Context, knowledgeBaseName string, page int, size int) ([]domainmodel.Document, int64, error)
	Delete(ctx context.Context, id int64) error
}

type ChunkLister interface {
	ListByDocumentID(ctx context.Context, documentID int64, page int, size int) ([]domainmodel.Chunk, int64, error)
	DeleteByID(ctx context.Context, id int64) error
	UpdateStatusByIDs(ctx context.Context, ids []int64, status int) error
	UpdateContentByID(ctx context.Context, id int64, content string) error
}

type Indexer interface {
	Index(ctx context.Context, in service.IndexInput) ([]string, error)
}

type Retriever interface {
	Retrieve(ctx context.Context, in service.RetrieveInput) ([]domainmodel.RetrievedChunk, error)
}

type Chatter interface {
	Chat(ctx context.Context, in service.ChatInput) (*service.ChatResult, error)
	ChatStream(ctx context.Context, in service.ChatInput) (*service.ChatStreamResult, error)
}

type Handler struct {
	knowledgeBase KnowledgeBaseCreator
	document      DocumentLister
	chunk         ChunkLister
	indexer       Indexer
	retriever     Retriever
	chat          Chatter
}

func NewHandler(knowledgeBase KnowledgeBaseCreator, document DocumentLister, chunk ChunkLister, indexer Indexer, retriever Retriever, chat Chatter) *Handler {
	return &Handler{knowledgeBase: knowledgeBase, document: document, chunk: chunk, indexer: indexer, retriever: retriever, chat: chat}
}

func (h *Handler) writeBindError(c *gin.Context, err error) {
	webmiddleware.WriteBadRequest(c, err.Error())
}

func (h *Handler) writeServiceError(c *gin.Context, err error) {
	webmiddleware.WriteInternalError(c, err.Error())
}

func (h *Handler) NotImplemented(c *gin.Context) {
	webmiddleware.WriteJSON(c, http.StatusNotImplemented, webmiddleware.CodeInternalError, "接口暂未实现", nil)
}

func (h *Handler) writeDependencyMissing(c *gin.Context) {
	webmiddleware.WriteServiceUnavailable(c, "服务未配置")
}

func writeSSEEvent(c *gin.Context, event string, data string) {
	_, _ = fmt.Fprintf(c.Writer, "event: %s\n", event)
	_, _ = fmt.Fprintf(c.Writer, "data: %s\n\n", data)
	c.Writer.Flush()
}

func encodeJSON(value any) string {
	payload, _ := json.Marshal(value)
	return string(payload)
}
