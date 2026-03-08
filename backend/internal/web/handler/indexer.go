package handler

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/gaohao-creator/go-rag/api/dto"
	"github.com/gaohao-creator/go-rag/internal/service"
	webmiddleware "github.com/gaohao-creator/go-rag/internal/web/middleware"
	"github.com/gin-gonic/gin"
)

func (h *Handler) Index(c *gin.Context) {
	if h.indexer == nil {
		h.writeDependencyMissing(c)
		return
	}
	input, err := h.buildIndexInput(c)
	if err != nil {
		h.writeBindError(c, err)
		return
	}
	ids, err := h.indexer.Index(c.Request.Context(), input)
	if err != nil {
		h.writeServiceError(c, err)
		return
	}
	webmiddleware.WriteOK(c, dto.IndexResponse{DocIDs: ids})
}

func (h *Handler) buildIndexInput(c *gin.Context) (service.IndexInput, error) {
	fileHeader, err := c.FormFile("file")
	if err == nil && fileHeader != nil {
		tempDir, createErr := os.MkdirTemp("", "go-rag-upload-*")
		if createErr != nil {
			return service.IndexInput{}, createErr
		}
		destination := filepath.Join(tempDir, fileHeader.Filename)
		if saveErr := c.SaveUploadedFile(fileHeader, destination); saveErr != nil {
			return service.IndexInput{}, saveErr
		}
		knowledgeName := strings.TrimSpace(c.PostForm("knowledge_name"))
		return service.IndexInput{URI: destination, KnowledgeName: knowledgeName, FileName: fileHeader.Filename}, nil
	}
	var req dto.IndexRequest
	if bindErr := c.ShouldBind(&req); bindErr != nil {
		return service.IndexInput{}, bindErr
	}
	uri := strings.TrimSpace(req.URI)
	if uri == "" {
		uri = strings.TrimSpace(req.URL)
	}
	fileName := strings.TrimSpace(req.FileName)
	if fileName == "" && uri != "" {
		fileName = filepath.Base(uri)
	}
	return service.IndexInput{URI: uri, KnowledgeName: req.KnowledgeName, FileName: fileName}, nil
}
