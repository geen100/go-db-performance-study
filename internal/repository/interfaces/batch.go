// internal/repository/interfaces/batch.go
package interfaces

import (
	"go-db-performance-study/internal/models"
)

// BatchRepository バッチ操作インターフェース
type BatchRepository interface {
	// ユーザーバッチ操作
	CreateUsersBatch(users []models.User, batchSize int) error
	UpdateUsersBatch(updates []UserBatchUpdate, batchSize int) error
	DeleteUsersBatch(ids []uint, batchSize int) error

	// 投稿バッチ操作
	CreatePostsBatch(posts []models.Post, batchSize int) error
	UpdatePostsBatch(updates []PostBatchUpdate, batchSize int) error
	DeletePostsBatch(ids []uint, batchSize int) error

	// コメントバッチ操作
	CreateCommentsBatch(comments []models.Comment, batchSize int) error
	UpdateCommentsBatch(updates []CommentBatchUpdate, batchSize int) error
	DeleteCommentsBatch(ids []uint, batchSize int) error

	// タグバッチ操作
	CreateTagsBatch(tags []models.Tag, batchSize int) error
	UpdateTagsBatch(updates []TagBatchUpdate, batchSize int) error
	DeleteTagsBatch(ids []uint, batchSize int) error
}

// バッチ更新用構造体
type UserBatchUpdate struct {
	ID   uint                  `json:"id" validate:"required"`
	Data *models.UserForUpdate `json:"data" validate:"required"`
}

type PostBatchUpdate struct {
	ID   uint                  `json:"id" validate:"required"`
	Data *models.PostForUpdate `json:"data" validate:"required"`
}

type CommentBatchUpdate struct {
	ID   uint                     `json:"id" validate:"required"`
	Data *models.CommentForUpdate `json:"data" validate:"required"`
}

type TagBatchUpdate struct {
	ID   uint                 `json:"id" validate:"required"`
	Data *models.TagForUpdate `json:"data" validate:"required"`
}
