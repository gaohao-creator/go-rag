package service

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"path/filepath"
	"strings"

	fileloader "github.com/cloudwego/eino-ext/components/document/loader/file"
	urlloader "github.com/cloudwego/eino-ext/components/document/loader/url"
	htmlparser "github.com/cloudwego/eino-ext/components/document/parser/html"
	pdfparser "github.com/cloudwego/eino-ext/components/document/parser/pdf"
	xlsxparser "github.com/cloudwego/eino-ext/components/document/parser/xlsx"
	markdownsplitter "github.com/cloudwego/eino-ext/components/document/transformer/splitter/markdown"
	recursivesplitter "github.com/cloudwego/eino-ext/components/document/transformer/splitter/recursive"
	doccomponent "github.com/cloudwego/eino/components/document"
	docparser "github.com/cloudwego/eino/components/document/parser"
	domainservice "github.com/gaohao-creator/go-rag/internal/domain/service"
	"github.com/cloudwego/eino/schema"
	"github.com/google/uuid"
)

type defaultIndexerEngine struct {
	fileLoader           doccomponent.Loader
	urlLoader            doccomponent.Loader
	markdownTransformer  doccomponent.Transformer
	recursiveTransformer doccomponent.Transformer
}

func NewDefaultIndexerEngine() (domainservice.Indexer, error) {
	ctx := context.Background()
	parser, err := newExtParser(ctx)
	if err != nil {
		return nil, err
	}
	fileLoader, err := fileloader.NewFileLoader(ctx, &fileloader.FileLoaderConfig{
		UseNameAsID: false,
		Parser:      parser,
	})
	if err != nil {
		return nil, fmt.Errorf("初始化文件加载器失败: %w", err)
	}
	urlLoader, err := urlloader.NewLoader(ctx, &urlloader.LoaderConfig{})
	if err != nil {
		return nil, fmt.Errorf("初始化 URL 加载器失败: %w", err)
	}
	markdownTransformer, err := markdownsplitter.NewHeaderSplitter(ctx, &markdownsplitter.HeaderConfig{
		Headers: map[string]string{
			"#":   "title1",
			"##":  "title2",
			"###": "title3",
		},
		TrimHeaders: false,
	})
	if err != nil {
		return nil, fmt.Errorf("初始化 Markdown 切分器失败: %w", err)
	}
	recursiveTransformer, err := recursivesplitter.NewSplitter(ctx, &recursivesplitter.Config{
		ChunkSize:   1000,
		OverlapSize: 100,
		Separators:  []string{"\n\n", "\n", "。", "？", "！", ". ", "? ", "! ", "，", ","},
	})
	if err != nil {
		return nil, fmt.Errorf("初始化递归切分器失败: %w", err)
	}
	return &defaultIndexerEngine{
		fileLoader:           fileLoader,
		urlLoader:            urlLoader,
		markdownTransformer:  markdownTransformer,
		recursiveTransformer: recursiveTransformer,
	}, nil
}

func (e *defaultIndexerEngine) Index(ctx context.Context, req domainservice.IndexRequest) ([]domainservice.IndexedChunk, error) {
	if strings.TrimSpace(req.URI) == "" {
		return nil, fmt.Errorf("文档地址不能为空")
	}
	loader := e.fileLoader
	if isURL(req.URI) {
		loader = e.urlLoader
	}
	docs, err := loader.Load(ctx, doccomponent.Source{URI: req.URI})
	if err != nil {
		return nil, fmt.Errorf("加载文档失败: %w", err)
	}
	transformer := e.recursiveTransformer
	if strings.EqualFold(filepath.Ext(req.URI), ".md") {
		transformer = e.markdownTransformer
	}
	transformed, err := transformer.Transform(ctx, docs)
	if err != nil {
		return nil, fmt.Errorf("切分文档失败: %w", err)
	}
	return toIndexedChunks(transformed)
}

func newExtParser(ctx context.Context) (docparser.Parser, error) {
	textParser := docparser.TextParser{}
	html, err := htmlparser.NewParser(ctx, &htmlparser.Config{Selector: &[]string{"body"}[0]})
	if err != nil {
		return nil, fmt.Errorf("初始化 HTML 解析器失败: %w", err)
	}
	xlsx, err := xlsxparser.NewXlsxParser(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("初始化 XLSX 解析器失败: %w", err)
	}
	pdf, err := pdfparser.NewPDFParser(ctx, &pdfparser.Config{})
	if err != nil {
		return nil, fmt.Errorf("初始化 PDF 解析器失败: %w", err)
	}
	parser, err := docparser.NewExtParser(ctx, &docparser.ExtParserConfig{
		Parsers: map[string]docparser.Parser{
			".html": html,
			".pdf":  pdf,
			".xlsx": xlsx,
		},
		FallbackParser: textParser,
	})
	if err != nil {
		return nil, fmt.Errorf("初始化扩展解析器失败: %w", err)
	}
	return parser, nil
}

func toIndexedChunks(docs []*schema.Document) ([]domainservice.IndexedChunk, error) {
	chunks := make([]domainservice.IndexedChunk, 0, len(docs))
	for _, doc := range docs {
		if doc == nil || strings.TrimSpace(doc.Content) == "" {
			continue
		}
		if strings.TrimSpace(doc.ID) == "" {
			doc.ID = uuid.NewString()
		}
		ext := "{}"
		if len(doc.MetaData) > 0 {
			payload, err := json.Marshal(doc.MetaData)
			if err != nil {
				return nil, fmt.Errorf("序列化 chunk 元数据失败: %w", err)
			}
			ext = string(payload)
		}
		chunks = append(chunks, domainservice.IndexedChunk{
			ChunkID: doc.ID,
			Content: doc.Content,
			Ext:     ext,
		})
	}
	return chunks, nil
}

func isURL(raw string) bool {
	parsed, err := url.Parse(raw)
	if err != nil {
		return false
	}
	return parsed.Scheme != "" && parsed.Host != ""
}
