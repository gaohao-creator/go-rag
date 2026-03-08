package model

type KnowledgeBasePatch struct {
	Name        *string `json:"name,omitempty"`
	Description *string `json:"description,omitempty"`
	Category    *string `json:"category,omitempty"`
	Status      *int    `json:"status,omitempty"`
}
