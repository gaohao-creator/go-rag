package handler

import (
	"strconv"

	"github.com/gaohao-creator/go-rag/api/dto"
	"github.com/gaohao-creator/go-rag/internal/service"
	webmiddleware "github.com/gaohao-creator/go-rag/internal/web/middleware"
	"github.com/gin-gonic/gin"
)

func (h *Handler) CreateKnowledgeBase(c *gin.Context) {
	if h.knowledgeBase == nil {
		h.writeDependencyMissing(c)
		return
	}
	var req dto.KnowledgeBaseCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.writeBindError(c, err)
		return
	}
	id, err := h.knowledgeBase.Create(c.Request.Context(), service.CreateKnowledgeBaseInput{Name: req.Name, Description: req.Description, Category: req.Category})
	if err != nil {
		h.writeServiceError(c, err)
		return
	}
	webmiddleware.WriteOK(c, dto.KnowledgeBaseCreateResponse{ID: id})
}

func (h *Handler) ListKnowledgeBases(c *gin.Context) {
	if h.knowledgeBase == nil {
		h.writeDependencyMissing(c)
		return
	}
	var req dto.KnowledgeBaseListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		h.writeBindError(c, err)
		return
	}
	list, err := h.knowledgeBase.List(c.Request.Context(), service.ListKnowledgeBasesInput{Name: req.Name, Status: req.Status, Category: req.Category})
	if err != nil {
		h.writeServiceError(c, err)
		return
	}
	webmiddleware.WriteOK(c, dto.KnowledgeBaseListResponse{List: toKnowledgeBaseViews(list)})
}

func (h *Handler) GetKnowledgeBase(c *gin.Context) {
	if h.knowledgeBase == nil {
		h.writeDependencyMissing(c)
		return
	}
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		h.writeBindError(c, err)
		return
	}
	kb, err := h.knowledgeBase.GetByID(c.Request.Context(), id)
	if err != nil {
		h.writeServiceError(c, err)
		return
	}
	webmiddleware.WriteOK(c, toKnowledgeBaseView(kb))
}

func (h *Handler) UpdateKnowledgeBase(c *gin.Context) {
	if h.knowledgeBase == nil {
		h.writeDependencyMissing(c)
		return
	}
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		h.writeBindError(c, err)
		return
	}
	var req dto.KnowledgeBaseUpdateRequest
	if err = c.ShouldBindJSON(&req); err != nil {
		h.writeBindError(c, err)
		return
	}
	if err = h.knowledgeBase.Update(c.Request.Context(), service.UpdateKnowledgeBaseInput{ID: id, Name: req.Name, Description: req.Description, Category: req.Category, Status: req.Status}); err != nil {
		h.writeServiceError(c, err)
		return
	}
	webmiddleware.WriteOK(c, gin.H{})
}

func (h *Handler) DeleteKnowledgeBase(c *gin.Context) {
	if h.knowledgeBase == nil {
		h.writeDependencyMissing(c)
		return
	}
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		h.writeBindError(c, err)
		return
	}
	if err = h.knowledgeBase.Delete(c.Request.Context(), id); err != nil {
		h.writeServiceError(c, err)
		return
	}
	webmiddleware.WriteOK(c, gin.H{})
}
