package handler

import (
	"github.com/gaohao-creator/go-rag/api/dto"
	"github.com/gaohao-creator/go-rag/internal/service"
	webmiddleware "github.com/gaohao-creator/go-rag/internal/web/middleware"
	"github.com/gin-gonic/gin"
)

func (h *Handler) Retrieve(c *gin.Context) {
	if h.retriever == nil {
		h.writeDependencyMissing(c)
		return
	}
	var req dto.RetrieveRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.writeBindError(c, err)
		return
	}
	chunks, err := h.retriever.Retrieve(c.Request.Context(), service.RetrieveInput{
		Question:      req.Question,
		TopK:          req.TopK,
		Score:         req.Score,
		KnowledgeName: req.KnowledgeName,
	})
	if err != nil {
		h.writeServiceError(c, err)
		return
	}
	webmiddleware.WriteOK(c, dto.RetrieveResponse{Document: toSchemaDocuments(chunks)})
}

func (h *Handler) RetrieveDify(c *gin.Context) {
	if h.retriever == nil {
		h.writeDependencyMissing(c)
		return
	}
	var req dto.RetrieveDifyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.writeBindError(c, err)
		return
	}
	chunks, err := h.retriever.Retrieve(c.Request.Context(), service.RetrieveInput{
		Question:      req.Query,
		TopK:          req.RetrievalSetting.TopK,
		Score:         req.RetrievalSetting.ScoreThreshold,
		KnowledgeName: req.KnowledgeID,
	})
	if err != nil {
		h.writeServiceError(c, err)
		return
	}
	c.JSON(200, toRetrieveDifyResponse(chunks))
}
