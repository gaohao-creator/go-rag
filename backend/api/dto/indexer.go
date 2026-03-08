package dto

type IndexRequest struct {
	URI           string `json:"uri" form:"uri"`
	URL           string `json:"url" form:"url"`
	KnowledgeName string `json:"knowledge_name" form:"knowledge_name" binding:"required"`
	FileName      string `json:"file_name" form:"file_name"`
}

type IndexResponse struct {
	DocIDs []string `json:"doc_ids"`
}
