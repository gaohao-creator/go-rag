package handler

import (
	"github.com/gaohao-creator/go-rag/api/dto"
	webmiddleware "github.com/gaohao-creator/go-rag/internal/web/middleware"
	"github.com/gin-gonic/gin"
)

func (h *Handler) ListChunks(c *gin.Context) {
	if h.chunk == nil {
		h.writeDependencyMissing(c)
		return
	}
	var req dto.ChunkListRequest
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
	chunks, total, err := h.chunk.ListByDocumentID(c.Request.Context(), req.KnowledgeDocID, req.Page, req.Size)
	if err != nil {
		h.writeServiceError(c, err)
		return
	}
	webmiddleware.WriteOK(c, dto.ChunkListResponse{Data: toChunkViews(chunks), Total: total, Page: req.Page, Size: req.Size})
}

func (h *Handler) DeleteChunk(c *gin.Context) {
	if h.chunk == nil {
		h.writeDependencyMissing(c)
		return
	}
	var req dto.ChunkDeleteRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		h.writeBindError(c, err)
		return
	}
	if err := h.chunk.DeleteByID(c.Request.Context(), req.ID); err != nil {
		h.writeServiceError(c, err)
		return
	}
	webmiddleware.WriteOK(c, gin.H{})
}

func (h *Handler) UpdateChunkStatus(c *gin.Context) {
	if h.chunk == nil {
		h.writeDependencyMissing(c)
		return
	}
	var req dto.ChunkStatusUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.writeBindError(c, err)
		return
	}
	if err := h.chunk.UpdateStatusByIDs(c.Request.Context(), req.IDs, req.Status); err != nil {
		h.writeServiceError(c, err)
		return
	}
	webmiddleware.WriteOK(c, gin.H{})
}

func (h *Handler) UpdateChunkContent(c *gin.Context) {
	if h.chunk == nil {
		h.writeDependencyMissing(c)
		return
	}
	var req dto.ChunkContentUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.writeBindError(c, err)
		return
	}
	if err := h.chunk.UpdateContentByID(c.Request.Context(), req.ID, req.Content); err != nil {
		h.writeServiceError(c, err)
		return
	}
	webmiddleware.WriteOK(c, gin.H{})
}
