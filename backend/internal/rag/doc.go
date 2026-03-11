// Package rag 提供 backend 内部统一的 RAG 门面。
//
// 这个包不关心具体实现细节，只负责把索引、存储、检索和质量判定
// 这些能力收口成一个统一对象，供 service 层直接使用。
package rag
