package ioc

import (
	appconfig "github.com/gaohao-creator/go-rag/config"
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

	knowledgeBaseRepository := datarepo.NewKnowledgeBaseRepository(knowledgeBaseDAO)
	documentRepository := datarepo.NewDocumentRepository(documentDAO)
	chunkRepository := datarepo.NewChunkRepository(chunkDAO)

	indexerEngine, err := appservice.NewDefaultIndexerEngine()
	if err != nil {
		return nil, err
	}
	retrieverEngine := appservice.NewDatabaseRetrieverEngine(documentRepository, chunkRepository)

	services := &Services{
		KnowledgeBase: appservice.NewKnowledgeBaseService(knowledgeBaseRepository),
		Document:      appservice.NewDocumentService(documentRepository),
		Chunk:         appservice.NewChunkService(chunkRepository),
		Indexer:       appservice.NewIndexerService(documentRepository, chunkRepository, indexerEngine),
		Retriever:     appservice.NewRetrieverService(retrieverEngine),
	}

	handler := webhandler.NewHandler(
		services.KnowledgeBase,
		services.Document,
		services.Chunk,
		services.Indexer,
		services.Retriever,
	)
	router := webrouter.NewRouter(handler)

	return &App{
		Config:   conf,
		DB:       db,
		Handler:  handler,
		Router:   router,
		Services: services,
	}, nil
}
