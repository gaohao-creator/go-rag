CREATE DATABASE IF NOT EXISTS `go_rag`
  DEFAULT CHARACTER SET utf8mb4
  COLLATE utf8mb4_general_ci;

USE `go_rag`;

CREATE TABLE IF NOT EXISTS `knowledge_base` (
  `id` BIGINT NOT NULL AUTO_INCREMENT,
  `name` VARCHAR(50) NOT NULL,
  `description` VARCHAR(200) NOT NULL,
  `category` VARCHAR(50) DEFAULT '',
  `status` INT NOT NULL DEFAULT 1,
  `create_time` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `update_time` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `idx_knowledge_base_name` (`name`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

CREATE TABLE IF NOT EXISTS `knowledge_documents` (
  `id` BIGINT NOT NULL AUTO_INCREMENT,
  `knowledge_base_name` VARCHAR(50) NOT NULL,
  `file_name` VARCHAR(255) NOT NULL,
  `status` INT NOT NULL DEFAULT 0,
  `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `idx_documents_knowledge_base_name` (`knowledge_base_name`),
  KEY `idx_documents_status` (`status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

CREATE TABLE IF NOT EXISTS `knowledge_chunks` (
  `id` BIGINT NOT NULL AUTO_INCREMENT,
  `knowledge_doc_id` BIGINT NOT NULL,
  `chunk_id` VARCHAR(128) NOT NULL,
  `content` LONGTEXT NOT NULL,
  `ext` LONGTEXT NULL,
  `status` INT NOT NULL DEFAULT 0,
  `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `idx_chunks_knowledge_doc_id` (`knowledge_doc_id`),
  KEY `idx_chunks_chunk_id` (`chunk_id`),
  KEY `idx_chunks_status` (`status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;
