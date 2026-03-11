package ioc

import (
	"context"
	"fmt"
	"strings"

	appconfig "github.com/gaohao-creator/go-rag/config"
	domainservice "github.com/gaohao-creator/go-rag/internal/domain/service"
	apprag "github.com/gaohao-creator/go-rag/internal/rag"
	rages "github.com/gaohao-creator/go-rag/internal/rag/es"
	raggrade "github.com/gaohao-creator/go-rag/internal/rag/grade"
	ragindex "github.com/gaohao-creator/go-rag/internal/rag/index"
	ragrerank "github.com/gaohao-creator/go-rag/internal/rag/rerank"
	ragretrieve "github.com/gaohao-creator/go-rag/internal/rag/retrieve"
	ragstore "github.com/gaohao-creator/go-rag/internal/rag/store"
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

	indexerEngine, err := ragindex.NewIndex()
	if err != nil {
		return nil, err
	}
	chatModel, err := buildChatModel(conf)
	if err != nil {
		return nil, err
	}
	promptModel, reranker, grader, err := buildQualityComponents(conf, chatModel)
	if err != nil {
		return nil, err
	}
	chunkStore, vectorRetriever, qaVectorRetriever, err := buildVectorComponents(conf)
	if err != nil {
		return nil, err
	}
	var qaPromptModel domainservice.PromptModel
	var qaQuestionCount int
	if conf != nil && conf.Quality.QA.Enabled {
		qaPromptModel = promptModel
		qaQuestionCount = conf.Quality.QA.QuestionCount
	}
	retrieverEngine := ragretrieve.NewRetrieve(documentRepository, chunkRepository, vectorRetriever, qaVectorRetriever, reranker)
	chunkStoreFacade := ragstore.NewStore(chunkStore, qaPromptModel, qaQuestionCount)
	ragService := apprag.NewRAG(indexerEngine, chunkStoreFacade, retrieverEngine, grader)
	retrieverService := appservice.NewRetrieverService(ragService)

	services := &Services{
		KnowledgeBase: appservice.NewKnowledgeBaseService(knowledgeBaseRepository),
		Document:      appservice.NewDocumentService(documentRepository),
		Chunk:         appservice.NewChunkService(chunkRepository),
		Indexer:       appservice.NewIndexerService(documentRepository, chunkRepository, ragService),
		Retriever:     retrieverService,
		Chat:          appservice.NewChatService(messageRepository, retrieverService, chatModel, ragService),
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

func buildVectorComponents(conf *appconfig.Config) (domainservice.ChunkStore, domainservice.Retriever, domainservice.Retriever, error) {
	if conf == nil || !conf.Vector.Enabled {
		return nil, nil, nil, nil
	}

	esClient, err := rages.NewES(context.Background(), conf.Vector)
	if err != nil {
		return nil, nil, nil, err
	}

	store, err := esClient.NewStore(context.Background())
	if err != nil {
		return nil, nil, nil, err
	}
	retriever, err := esClient.NewContentRetriever(context.Background())
	if err != nil {
		return nil, nil, nil, err
	}
	qaRetriever, err := esClient.NewQARetriever(context.Background())
	if err != nil {
		return nil, nil, nil, err
	}
	return store, retriever, qaRetriever, nil
}

func buildQualityComponents(
	conf *appconfig.Config,
	chatModel domainservice.ChatModel,
) (domainservice.PromptModel, domainservice.Reranker, domainservice.Grader, error) {
	if conf == nil {
		return nil, nil, nil, nil
	}
	var promptModel domainservice.PromptModel
	if conf.Quality.QA.Enabled || conf.Quality.Grader.Enabled {
		value, ok := chatModel.(domainservice.PromptModel)
		if !ok {
			return nil, nil, nil, fmt.Errorf("当前 chat model 不支持 prompt 能力")
		}
		promptModel = value
	}

	var reranker domainservice.Reranker
	if conf.Rerank.Enabled {
		reranker = ragrerank.NewRerank(nil, conf.Rerank)
	}

	var grader domainservice.Grader
	if conf.Quality.Grader.Enabled {
		grader = raggrade.NewGrade(promptModel)
	}
	return promptModel, reranker, grader, nil
}
