package handler

import (
	"github.com/gaohao-creator/go-rag/api/dto"
	webmiddleware "github.com/gaohao-creator/go-rag/internal/web/middleware"
	"github.com/gin-gonic/gin"
)

func (h *Handler) ListDocuments(c *gin.Context) {
	if h.document == nil {
		h.writeDependencyMissing(c)
		return
	}
	var req dto.DocumentListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		h.writeBindError(c, err)
		return
	}
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.Size <= 0 {
		req.Size = 10
	}
	documents, total, err := h.document.ListByKnowledgeBase(c.Request.Context(), req.KnowledgeName, req.Page, req.Size)
	if err != nil {
		h.writeServiceError(c, err)
		return
	}
	webmiddleware.WriteOK(c, dto.DocumentListResponse{Data: toDocumentViews(documents), Total: total, Page: req.Page, Size: req.Size})
}

func (h *Handler) DeleteDocument(c *gin.Context) {
	if h.document == nil {
		h.writeDependencyMissing(c)
		return
	}
	var req dto.DocumentDeleteRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		h.writeBindError(c, err)
		return
	}
	if err := h.document.Delete(c.Request.Context(), req.DocumentID); err != nil {
		h.writeServiceError(c, err)
		return
	}
	webmiddleware.WriteOK(c, gin.H{})
}
