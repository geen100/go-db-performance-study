// internal/repository/interfaces/post.go
package interfaces

import (
    "go-db-performance-study/internal/models"
    "time"
)

// PostRepository 投稿リポジトリインターフェース
type PostRepository interface {
    // 基本CRUD
    Create(post *models.Post) error
    GetByID(id uint) (*models.Post, error)
    GetBySlug(slug string) (*models.Post, error)
    Update(id uint, updates *models.PostForUpdate) error
    Delete(id uint) error
    
    // 一覧取得
    List(limit, offset int) ([]models.Post, error)
    ListByUser(userID uint, limit, offset int) ([]models.Post, error)
    ListByStatus(status models.PostStatus, limit, offset int) ([]models.Post, error)
    ListByTag(tagID uint, limit, offset int) ([]models.Post, error)
    
    // 検索・フィルタ
    Search(query string, limit, offset int) ([]models.Post, error)
    GetPopularPosts(limit int) ([]models.Post, error)
    GetRecentPosts(limit int) ([]models.Post, error)
    GetPostsByDateRange(from, to time.Time, limit, offset int) ([]models.Post, error)
    
    // 統計
    Count() (int64, error)
    CountByUser(userID uint) (int64, error)
    CountByStatus(status models.PostStatus) (int64, error)
    
    // 関連操作
    AddTags(postID uint, tagIDs []uint) error
    RemoveTags(postID uint, tagIDs []uint) error
    UpdateViewCount(id uint) error
}