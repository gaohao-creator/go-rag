package handler

import (
	"github.com/gaohao-creator/go-rag/api/dto"
	"github.com/gaohao-creator/go-rag/internal/service"
	webmiddleware "github.com/gaohao-creator/go-rag/internal/web/middleware"
	"github.com/gin-gonic/gin"
)

func (h *Handler) Chat(c *gin.Context) {
	if h.chat == nil {
		h.writeDependencyMissing(c)
		return
	}
	var req dto.ChatRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.writeBindError(c, err)
		return
	}
	result, err := h.chat.Chat(c.Request.Context(), service.ChatInput{ConvID: req.ConvID, Question: req.Question, KnowledgeName: req.KnowledgeName, TopK: req.TopK, Score: req.Score})
	if err != nil {
		h.writeServiceError(c, err)
		return
	}
	webmiddleware.WriteOK(c, dto.ChatResponse{Answer: result.Answer, References: result.References})
}

func (h *Handler) ChatStream(c *gin.Context) {
	if h.chat == nil {
		h.writeDependencyMissing(c)
		return
	}
	var req dto.ChatRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.writeBindError(c, err)
		return
	}
	result, err := h.chat.ChatStream(c.Request.Context(), service.ChatInput{ConvID: req.ConvID, Question: req.Question, KnowledgeName: req.KnowledgeName, TopK: req.TopK, Score: req.Score})
	if err != nil {
		h.writeServiceError(c, err)
		return
	}
	c.Writer.Header().Set("Content-Type", "text/event-stream")
	c.Writer.Header().Set("Cache-Control", "no-cache")
	c.Writer.Header().Set("Connection", "keep-alive")
	writeSSEEvent(c, "references", encodeJSON(result.References))
	for _, chunk := range result.Chunks {
		writeSSEEvent(c, "message", chunk)
	}
	writeSSEEvent(c, "done", "[DONE]")
}
