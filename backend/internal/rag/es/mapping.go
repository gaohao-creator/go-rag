package es

import appconfig "github.com/gaohao-creator/go-rag/config"

func buildMappings(conf appconfig.VectorConfig) map[string]any {
	return map[string]any{
		"properties": map[string]any{
			conf.ContentField: map[string]any{
				"type": "text",
			},
			conf.ContentVectorField: map[string]any{
				"type":       "dense_vector",
				"dims":       conf.Dimensions,
				"index":      true,
				"similarity": "cosine",
			},
			conf.QAContentField: map[string]any{
				"type": "text",
			},
			conf.QAContentVectorField: map[string]any{
				"type":       "dense_vector",
				"dims":       conf.Dimensions,
				"index":      true,
				"similarity": "cosine",
			},
			conf.KnowledgeField: map[string]any{
				"type": "keyword",
			},
			conf.ExtField: map[string]any{
				"type": "text",
			},
		},
	}
}
