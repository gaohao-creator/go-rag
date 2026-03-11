package handler

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/cloudwego/eino/schema"
	"github.com/gaohao-creator/go-rag/api/dto"
	domainmodel "github.com/gaohao-creator/go-rag/internal/domain/model"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type legacyStreamData struct {
	ID       string             `json:"id"`
	Created  int64              `json:"created"`
	Content  string             `json:"content"`
	Document []*schema.Document `json:"document"`
}

func toSchemaDocuments(chunks []domainmodel.RetrievedChunk) []*schema.Document {
	documents := make([]*schema.Document, 0, len(chunks))
	for _, chunk := range chunks {
		meta := make(map[string]any)
		if chunk.KnowledgeDocID != 0 {
			meta["knowledge_doc_id"] = chunk.KnowledgeDocID
		}
		if ext := parseExt(chunk.Ext); ext != nil {
			meta["ext"] = ext
		}
		doc := &schema.Document{
			ID:      chunk.ChunkID,
			Content: chunk.Content,
		}
		if len(meta) > 0 {
			doc.MetaData = meta
		}
		doc.WithScore(chunk.Score)
		documents = append(documents, doc)
	}
	return documents
}

func toKnowledgeBaseViews(list []domainmodel.KnowledgeBase) []dto.KnowledgeBaseView {
	views := make([]dto.KnowledgeBaseView, 0, len(list))
	for _, item := range list {
		views = append(views, dto.KnowledgeBaseView{
			ID:          item.ID,
			Name:        item.Name,
			Description: item.Description,
			Category:    item.Category,
			Status:      item.Status,
			CreateTime:  item.CreateTime,
			UpdateTime:  item.UpdateTime,
		})
	}
	return views
}

func toKnowledgeBaseView(item *domainmodel.KnowledgeBase) dto.KnowledgeBaseView {
	if item == nil {
		return dto.KnowledgeBaseView{}
	}
	return dto.KnowledgeBaseView{
		ID:          item.ID,
		Name:        item.Name,
		Description: item.Description,
		Category:    item.Category,
		Status:      item.Status,
		CreateTime:  item.CreateTime,
		UpdateTime:  item.UpdateTime,
	}
}

func toDocumentViews(list []domainmodel.Document) []dto.DocumentView {
	views := make([]dto.DocumentView, 0, len(list))
	for _, item := range list {
		views = append(views, dto.DocumentView{
			ID:                item.ID,
			KnowledgeBaseName: item.KnowledgeBaseName,
			FileName:          item.FileName,
			Status:            item.Status,
			CreatedAt:         item.CreatedAt,
			UpdatedAt:         item.UpdatedAt,
		})
	}
	return views
}

func toChunkViews(list []domainmodel.Chunk) []dto.ChunkView {
	views := make([]dto.ChunkView, 0, len(list))
	for _, item := range list {
		views = append(views, dto.ChunkView{
			ID:             item.ID,
			KnowledgeDocID: item.KnowledgeDocID,
			ChunkID:        item.ChunkID,
			Content:        item.Content,
			Ext:            item.Ext,
			Status:         item.Status,
			CreatedAt:      item.CreatedAt,
			UpdatedAt:      item.UpdatedAt,
		})
	}
	return views
}

func toRetrieveDifyResponse(chunks []domainmodel.RetrievedChunk) dto.RetrieveDifyResponse {
	records := make([]dto.RetrieveDifyRecord, 0, len(chunks))
	for _, chunk := range chunks {
		record := dto.RetrieveDifyRecord{
			Score:   chunk.Score,
			Title:   "",
			Content: chunk.Content,
		}
		if ext, ok := parseExt(chunk.Ext).(map[string]any); ok {
			metadata := &dto.RetrieveDifyMetadata{}
			if path, ok := ext["path"].(string); ok {
				metadata.Path = path
			}
			if description, ok := ext["description"].(string); ok {
				metadata.Description = description
			}
			if metadata.Path != "" || metadata.Description != "" {
				record.Metadata = metadata
			}
		}
		records = append(records, record)
	}
	return dto.RetrieveDifyResponse{Records: records}
}

func newLegacyStreamData() *legacyStreamData {
	return &legacyStreamData{
		ID:      uuid.NewString(),
		Created: time.Now().Unix(),
	}
}

func writeLegacyStreamDocuments(c *gin.Context, data *legacyStreamData) {
	writeLegacyStreamLine(c, "documents", encodeJSON(data))
}

func writeLegacyStreamData(c *gin.Context, data *legacyStreamData) {
	writeLegacyStreamLine(c, "data", encodeJSON(data))
}

func writeLegacyStreamDone(c *gin.Context) {
	writeLegacyStreamLine(c, "data", "[DONE]")
}

func writeLegacyStreamLine(c *gin.Context, prefix string, data string) {
	_, _ = fmt.Fprintf(c.Writer, "%s:%s\n\n", prefix, data)
	c.Writer.Flush()
}

func encodeJSON(value any) string {
	payload, _ := json.Marshal(value)
	return string(payload)
}

func parseExt(value string) any {
	text := strings.TrimSpace(value)
	if text == "" {
		return nil
	}
	var decoded any
	if err := json.Unmarshal([]byte(text), &decoded); err == nil {
		return decoded
	}
	return value
}
