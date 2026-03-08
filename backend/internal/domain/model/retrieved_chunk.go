package model

type RetrievedChunk struct {
	KnowledgeDocID int64   `json:"knowledge_doc_id"`
	ChunkID        string  `json:"chunk_id"`
	Content        string  `json:"content"`
	Ext            string  `json:"ext"`
	Score          float64 `json:"score"`
}
