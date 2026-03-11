package router

import (
	"github.com/gin-gonic/gin"

	webhandler "github.com/gaohao-creator/go-rag/internal/web/handler"
	webmiddleware "github.com/gaohao-creator/go-rag/internal/web/middleware"
)

func NewRouter(handlers ...*webhandler.Handler) *gin.Engine {
	engine := gin.New()
	engine.Use(webmiddleware.Logger(), webmiddleware.Recovery(), webmiddleware.CORS())
	engine.GET("/healthz", func(c *gin.Context) { webmiddleware.WriteOK(c, "pong") })
	if len(handlers) > 0 && handlers[0] != nil {
		registerAPIRoutes(engine, handlers[0])
	}
	return engine
}

func registerAPIRoutes(engine *gin.Engine, h *webhandler.Handler) {
	api := engine.Group("/api/v1")
	api.POST("/kb", h.CreateKnowledgeBase)
	api.PUT("/kb/:id", h.UpdateKnowledgeBase)
	api.DELETE("/kb/:id", h.DeleteKnowledgeBase)
	api.GET("/kb", h.ListKnowledgeBases)
	api.GET("/kb/:id", h.GetKnowledgeBase)

	api.GET("/documents", h.ListDocuments)
	api.DELETE("/documents", h.DeleteDocument)

	api.GET("/chunks", h.ListChunks)
	api.DELETE("/chunks", h.DeleteChunk)
	api.PUT("/chunks", h.UpdateChunkStatus)
	api.PUT("/chunks-content", h.UpdateChunkContent)
	api.PUT("/chunks_content", h.UpdateChunkContent)

	api.POST("/indexer", h.Index)
	api.POST("/retriever", h.Retrieve)
	api.POST("/dify/retrieval", h.RetrieveDify)
	api.POST("/chat", h.Chat)
	api.POST("/chat/stream", h.ChatStream)
}
