package rerank

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	appconfig "github.com/gaohao-creator/go-rag/config"
	domainmodel "github.com/gaohao-creator/go-rag/internal/domain/model"
	domainservice "github.com/gaohao-creator/go-rag/internal/domain/service"
)

// Rerank 负责调用外部服务对召回结果重排。
type Rerank struct {
	client *http.Client
	config appconfig.RerankConfig
}

type request struct {
	Model     string   `json:"model,omitempty"`
	Query     string   `json:"query"`
	Documents []string `json:"documents"`
	TopN      int      `json:"top_n"`
}

type response struct {
	Results []result `json:"results"`
}

type result struct {
	Index          int     `json:"index"`
	RelevanceScore float64 `json:"relevance_score"`
}

// NewRerank 创建外部 rerank 服务适配器。
func NewRerank(client *http.Client, config appconfig.RerankConfig) *Rerank {
	if client == nil {
		client = http.DefaultClient
	}
	return &Rerank{
		client: client,
		config: config,
	}
}

// Rerank 调用外部服务对召回结果重新排序。
func (r *Rerank) Rerank(ctx context.Context, in domainservice.RerankInput) ([]domainmodel.RetrievedChunk, error) {
	if len(in.Chunks) == 0 {
		return nil, nil
	}
	requestBody := request{
		Model:     strings.TrimSpace(r.config.Model),
		Query:     in.Question,
		Documents: make([]string, 0, len(in.Chunks)),
		TopN:      r.resolveTopN(in.TopK),
	}
	for _, chunk := range in.Chunks {
		requestBody.Documents = append(requestBody.Documents, chunk.Content)
	}

	body, err := json.Marshal(requestBody)
	if err != nil {
		return nil, err
	}
	httpRequest, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		strings.TrimRight(r.config.BaseURL, "/")+"/rerank",
		bytes.NewReader(body),
	)
	if err != nil {
		return nil, err
	}
	httpRequest.Header.Set("Content-Type", "application/json")
	if strings.TrimSpace(r.config.APIKey) != "" {
		httpRequest.Header.Set("Authorization", "Bearer "+strings.TrimSpace(r.config.APIKey))
	}

	httpResponse, err := r.client.Do(httpRequest)
	if err != nil {
		return nil, err
	}
	defer httpResponse.Body.Close()
	if httpResponse.StatusCode < http.StatusOK || httpResponse.StatusCode >= http.StatusMultipleChoices {
		return nil, fmt.Errorf("rerank 服务返回状态码 %d", httpResponse.StatusCode)
	}

	var payload response
	if err = json.NewDecoder(httpResponse.Body).Decode(&payload); err != nil {
		return nil, err
	}

	output := make([]domainmodel.RetrievedChunk, 0, len(payload.Results))
	for _, item := range payload.Results {
		if item.Index < 0 || item.Index >= len(in.Chunks) {
			continue
		}
		chunk := in.Chunks[item.Index]
		chunk.Score = item.RelevanceScore
		if chunk.Score < r.config.MinScore {
			continue
		}
		output = append(output, chunk)
	}
	return output, nil
}

func (r *Rerank) resolveTopN(topK int) int {
	if r.config.TopN > 0 {
		return r.config.TopN
	}
	if topK > 0 {
		return topK
	}
	return 5
}
