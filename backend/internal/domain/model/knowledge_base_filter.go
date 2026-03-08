package model

type KnowledgeBaseFilter struct {
	Name     *string `json:"name,omitempty"`
	Status   *int    `json:"status,omitempty"`
	Category *string `json:"category,omitempty"`
}
