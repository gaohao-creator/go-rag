package ioc

import (
	"context"
	"fmt"
	"strings"

	appconfig "github.com/gaohao-creator/go-rag/config"
	domainservice "github.com/gaohao-creator/go-rag/internal/domain/service"
	datarepo "github.com/gaohao-creator/go-rag/internal/repository"
	internaldao "github.com/gaohao-creator/go-rag/internal/repository/dao"
	appservice "github.com/gaohao-creator/go-rag/internal/service"
	webhandler "github.com/gaohao-creator/go-rag/internal/web/handler"
	webrouter "github.com/gaohao-creator/go-rag/internal/web/router"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type Services struct {
	KnowledgeBase *appservice.KnowledgeBaseService
	Document      *appservice.DocumentService
	Chunk         *appservice.ChunkService
	Indexer       *appservice.IndexerService
	Retriever     *appservice.RetrieverService
	Chat          *appservice.ChatService
}

type App struct {
	Config   *appconfig.Config
	DB       *gorm.DB
	Handler  *webhandler.Handler
	Router   *gin.Engine
	Services *Services
}

func NewApp(configPath string) (*App, error) {
	conf, err := NewConfig(configPath)
	if err != nil {
		return nil, err
	}
	db, err := NewDB(conf)
	if err != nil {
		return nil, err
	}
	if err = internaldao.AutoMigrate(db); err != nil {
		return nil, err
	}

	knowledgeBaseDAO := internaldao.NewKnowledgeBaseDAO(db)
	documentDAO := internaldao.NewDocumentDAO(db)
	chunkDAO := internaldao.NewChunkDAO(db)
	messageDAO := internaldao.NewMessageDAO(db)

	knowledgeBaseRepository := datarepo.NewKnowledgeBaseRepository(knowledgeBaseDAO)
	documentRepository := datarepo.NewDocumentRepository(documentDAO, chunkDAO)
	chunkRepository := datarepo.NewChunkRepository(chunkDAO)
	messageRepository := datarepo.NewMessageRepository(messageDAO)

	indexerEngine, err := appservice.NewDefaultIndexerEngine()
	if err != nil {
		return nil, err
	}
	retrieverEngine := appservice.NewDatabaseRetrieverEngine(documentRepository, chunkRepository)
	retrieverService := appservice.NewRetrieverService(retrieverEngine)
	chatModel, err := buildChatModel(conf)
	if err != nil {
		return nil, err
	}

	services := &Services{
		KnowledgeBase: appservice.NewKnowledgeBaseService(knowledgeBaseRepository),
		Document:      appservice.NewDocumentService(documentRepository),
		Chunk:         appservice.NewChunkService(chunkRepository),
		Indexer:       appservice.NewIndexerService(documentRepository, chunkRepository, indexerEngine),
		Retriever:     retrieverService,
		Chat:          appservice.NewChatService(messageRepository, retrieverService, chatModel),
	}

	handler := webhandler.NewHandler(
		services.KnowledgeBase,
		services.Document,
		services.Chunk,
		services.Indexer,
		services.Retriever,
		services.Chat,
	)
	router := webrouter.NewRouter(handler)

	return &App{Config: conf, DB: db, Handler: handler, Router: router, Services: services}, nil
}

func buildChatModel(conf *appconfig.Config) (domainservice.ChatModel, error) {
	if conf == nil {
		return nil, fmt.Errorf("配置不能为空")
	}
	provider := strings.TrimSpace(conf.Chat.Provider)
	switch provider {
	case "", "fake":
		return appservice.NewFakeChatModel(), nil
	case "openai":
		return appservice.NewOpenAICompatibleChatModelFromConfig(context.Background(), conf.Chat)
	default:
		return nil, fmt.Errorf("不支持的 chat.provider: %s", provider)
	}
}
