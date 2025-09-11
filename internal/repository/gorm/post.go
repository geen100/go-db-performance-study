// internal/repository/gorm/post.go
package gorm_repo

import (
	"fmt"
	"strings"
	"time"

	"go-db-performance-study/internal/models"
	"go-db-performance-study/internal/repository/interfaces"

	"gorm.io/gorm"
)

// postRepository 投稿リポジトリの実装
type postRepository struct {
	*BaseRepository
}

// NewPostRepository 投稿リポジトリを作成
func NewPostRepository(db *gorm.DB) interfaces.PostRepository {
	return &postRepository{
		BaseRepository: NewBaseRepository(db),
	}
}

// Create 投稿作成
func (r *postRepository) Create(post *models.Post) error {
	if err := post.Validate(); err != nil {
		return fmt.Errorf("バリデーションエラー: %w", err)
	}

	return r.db.Create(post).Error
}

// GetByID IDで投稿取得
func (r *postRepository) GetByID(id uint) (*models.Post, error) {
	var post models.Post
	err := r.db.Preload("User").Preload("Tags").Preload("Comments.User").First(&post, id).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("投稿が見つかりません: ID=%d", id)
		}
		return nil, err
	}
	return &post, nil
}

// GetBySlug スラッグで投稿取得
func (r *postRepository) GetBySlug(slug string) (*models.Post, error) {
	var post models.Post
	err := r.db.Where("slug = ?", slug).Preload("User").Preload("Tags").First(&post).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("投稿が見つかりません: Slug=%s", slug)
		}
		return nil, err
	}
	return &post, nil
}

// Update 投稿更新
func (r *postRepository) Update(id uint, updates *models.PostForUpdate) error {
	if err := models.ValidateStruct(updates); err != nil {
		return fmt.Errorf("バリデーションエラー: %w", err)
	}

	return r.WithTransaction(func(tx *gorm.DB) error {
		// 投稿本体を更新するための map に限定
		updateData := map[string]interface{}{}
		if updates.Title != nil {
			updateData["title"] = *updates.Title
		}
		if updates.Body != nil {
			updateData["body"] = *updates.Body
		}
		if updates.Status != nil {
			updateData["status"] = *updates.Status
		}

		if len(updateData) > 0 {
			result := tx.Model(&models.Post{}).Where("id = ?", id).Updates(updateData)
			if result.Error != nil {
				return result.Error
			}
			if result.RowsAffected == 0 {
				return fmt.Errorf("更新対象の投稿が見つかりません: ID=%d", id)
			}
		}

		// タグ関連付けを更新
		if updates.TagIDs != nil {
			var post models.Post
			if err := tx.First(&post, id).Error; err != nil {
				return err
			}

			var tags []models.Tag
			if len(updates.TagIDs) > 0 {
				if err := tx.Find(&tags, updates.TagIDs).Error; err != nil {
					return err
				}
			}

			if err := tx.Model(&post).Association("Tags").Replace(tags); err != nil {
				return err
			}
		}

		return nil
	})
}

// Delete 投稿削除
func (r *postRepository) Delete(id uint) error {
	result := r.db.Delete(&models.Post{}, id)
	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("削除対象の投稿が見つかりません: ID=%d", id)
	}

	return nil
}

// List 投稿一覧取得
func (r *postRepository) List(limit, offset int) ([]models.Post, error) {
	var posts []models.Post
	err := r.db.Preload("User").Preload("Tags").
		Order("created_at DESC").
		Limit(limit).Offset(offset).
		Find(&posts).Error
	return posts, err
}

// ListByUser ユーザー別投稿一覧取得
func (r *postRepository) ListByUser(userID uint, limit, offset int) ([]models.Post, error) {
	var posts []models.Post
	err := r.db.Where("user_id = ?", userID).
		Preload("User").Preload("Tags").
		Order("created_at DESC").
		Limit(limit).Offset(offset).
		Find(&posts).Error
	return posts, err
}

// ListByStatus ステータス別投稿一覧取得
func (r *postRepository) ListByStatus(status models.PostStatus, limit, offset int) ([]models.Post, error) {
	var posts []models.Post
	err := r.db.Where("status = ?", status).
		Preload("User").Preload("Tags").
		Order("created_at DESC").
		Limit(limit).Offset(offset).
		Find(&posts).Error
	return posts, err
}

// ListByTag タグ別投稿一覧取得
func (r *postRepository) ListByTag(tagID uint, limit, offset int) ([]models.Post, error) {
	var posts []models.Post
	err := r.db.Joins("JOIN post_tags ON posts.id = post_tags.post_id").
		Where("post_tags.tag_id = ?", tagID).
		Preload("User").Preload("Tags").
		Order("posts.created_at DESC").
		Limit(limit).Offset(offset).
		Find(&posts).Error
	return posts, err
}

// Search 投稿検索
func (r *postRepository) Search(query string, limit, offset int) ([]models.Post, error) {
	var posts []models.Post
	searchQuery := "%" + strings.ToLower(query) + "%"

	err := r.db.Where("LOWER(title) LIKE ? OR LOWER(body) LIKE ?", searchQuery, searchQuery).
		Preload("User").Preload("Tags").
		Order("created_at DESC").
		Limit(limit).Offset(offset).
		Find(&posts).Error

	return posts, err
}

// GetPopularPosts 人気投稿取得（閲覧数順）
func (r *postRepository) GetPopularPosts(limit int) ([]models.Post, error) {
	var posts []models.Post
	err := r.db.Where("status = ?", models.PostStatusPublished).
		Preload("User").Preload("Tags").
		Order("view_count DESC").
		Limit(limit).
		Find(&posts).Error
	return posts, err
}

// GetRecentPosts 最新投稿取得
func (r *postRepository) GetRecentPosts(limit int) ([]models.Post, error) {
	var posts []models.Post
	err := r.db.Where("status = ?", models.PostStatusPublished).
		Preload("User").Preload("Tags").
		Order("created_at DESC").
		Limit(limit).
		Find(&posts).Error
	return posts, err
}

// GetPostsByDateRange 日付範囲で投稿取得
func (r *postRepository) GetPostsByDateRange(from, to time.Time, limit, offset int) ([]models.Post, error) {
	var posts []models.Post
	err := r.db.Where("created_at BETWEEN ? AND ?", from, to).
		Preload("User").Preload("Tags").
		Order("created_at DESC").
		Limit(limit).Offset(offset).
		Find(&posts).Error
	return posts, err
}

// Count 投稿総数取得
func (r *postRepository) Count() (int64, error) {
	var count int64
	err := r.db.Model(&models.Post{}).Count(&count).Error
	return count, err
}

// CountByUser ユーザー別投稿数取得
func (r *postRepository) CountByUser(userID uint) (int64, error) {
	var count int64
	err := r.db.Model(&models.Post{}).Where("user_id = ?", userID).Count(&count).Error
	return count, err
}

// CountByStatus ステータス別投稿数取得
func (r *postRepository) CountByStatus(status models.PostStatus) (int64, error) {
	var count int64
	err := r.db.Model(&models.Post{}).Where("status = ?", status).Count(&count).Error
	return count, err
}

// AddTags 投稿にタグ追加
func (r *postRepository) AddTags(postID uint, tagIDs []uint) error {
	var post models.Post
	if err := r.db.First(&post, postID).Error; err != nil {
		return fmt.Errorf("投稿が見つかりません: ID=%d", postID)
	}

	var tags []models.Tag
	if err := r.db.Find(&tags, tagIDs).Error; err != nil {
		return err
	}

	return r.db.Model(&post).Association("Tags").Append(tags)
}

// RemoveTags 投稿からタグ削除
func (r *postRepository) RemoveTags(postID uint, tagIDs []uint) error {
	var post models.Post
	if err := r.db.First(&post, postID).Error; err != nil {
		return fmt.Errorf("投稿が見つかりません: ID=%d", postID)
	}

	var tags []models.Tag
	if err := r.db.Find(&tags, tagIDs).Error; err != nil {
		return err
	}

	return r.db.Model(&post).Association("Tags").Delete(tags)
}

// UpdateViewCount 閲覧数更新
func (r *postRepository) UpdateViewCount(id uint) error {
	return r.db.Model(&models.Post{}).Where("id = ?", id).
		UpdateColumn("view_count", gorm.Expr("view_count + 1")).Error
}
