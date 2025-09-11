// internal/repository/gorm/user.go
package gorm_repo

import (
    "fmt"
    "strings"

    "go-db-performance-study/internal/models"
    "go-db-performance-study/internal/repository/interfaces"
    "gorm.io/gorm"
)

// userRepository ユーザーリポジトリの実装
type userRepository struct {
    *BaseRepository
}

// NewUserRepository ユーザーリポジトリを作成
func NewUserRepository(db *gorm.DB) interfaces.UserRepository {
    return &userRepository{
        BaseRepository: NewBaseRepository(db),
    }
}

// Create ユーザー作成
func (r *userRepository) Create(user *models.User) error {
    if err := user.Validate(); err != nil {
        return fmt.Errorf("バリデーションエラー: %w", err)
    }
    
    return r.db.Create(user).Error
}

// GetByID IDでユーザー取得
func (r *userRepository) GetByID(id uint) (*models.User, error) {
    var user models.User
    err := r.db.First(&user, id).Error
    if err != nil {
        if err == gorm.ErrRecordNotFound {
            return nil, fmt.Errorf("ユーザーが見つかりません: ID=%d", id)
        }
        return nil, err
    }
    return &user, nil
}

// GetByEmail メールアドレスでユーザー取得
func (r *userRepository) GetByEmail(email string) (*models.User, error) {
    var user models.User
    err := r.db.Where("email = ?", email).First(&user).Error
    if err != nil {
        if err == gorm.ErrRecordNotFound {
            return nil, fmt.Errorf("ユーザーが見つかりません: Email=%s", email)
        }
        return nil, err
    }
    return &user, nil
}

// Update ユーザー更新
func (r *userRepository) Update(id uint, updates *models.UserForUpdate) error {
    if err := models.ValidateStruct(updates); err != nil {
        return fmt.Errorf("バリデーションエラー: %w", err)
    }
    
    result := r.db.Model(&models.User{}).Where("id = ?", id).Updates(updates)
    if result.Error != nil {
        return result.Error
    }
    
    if result.RowsAffected == 0 {
        return fmt.Errorf("更新対象のユーザーが見つかりません: ID=%d", id)
    }
    
    return nil
}

// Delete ユーザー削除
func (r *userRepository) Delete(id uint) error {
    result := r.db.Delete(&models.User{}, id)
    if result.Error != nil {
        return result.Error
    }
    
    if result.RowsAffected == 0 {
        return fmt.Errorf("削除対象のユーザーが見つかりません: ID=%d", id)
    }
    
    return nil
}

// List ユーザー一覧取得
func (r *userRepository) List(limit, offset int) ([]models.User, error) {
    var users []models.User
    err := r.db.Order("created_at DESC").Limit(limit).Offset(offset).Find(&users).Error
    return users, err
}

// ListWithStats 統計情報付きユーザー一覧取得
func (r *userRepository) ListWithStats(limit, offset int) ([]models.UserStats, error) {
    var users []models.UserStats
    err := r.db.Table("users").
        Select("users.id, users.name, users.email, " +
               "COUNT(DISTINCT posts.id) as post_count, " +
               "COUNT(comments.id) as comment_count").
        Joins("LEFT JOIN posts ON users.id = posts.user_id").
        Joins("LEFT JOIN comments ON users.id = comments.user_id").
        Group("users.id").
        Order("post_count DESC").
        Limit(limit).Offset(offset).
        Scan(&users).Error
    return users, err
}

// Search ユーザー検索
func (r *userRepository) Search(query string, limit, offset int) ([]models.User, error) {
    var users []models.User
    searchQuery := "%" + strings.ToLower(query) + "%"
    
    err := r.db.Where("LOWER(name) LIKE ? OR LOWER(email) LIKE ?", searchQuery, searchQuery).
        Order("created_at DESC").
        Limit(limit).Offset(offset).
        Find(&users).Error
    
    return users, err
}

// GetActiveUsers アクティブユーザー取得
func (r *userRepository) GetActiveUsers(limit int) ([]models.User, error) {
    var users []models.User
    err := r.db.Where("email_verified_at IS NOT NULL").
        Order("created_at DESC").
        Limit(limit).
        Find(&users).Error
    return users, err
}

// Count ユーザー総数取得
func (r *userRepository) Count() (int64, error) {
    var count int64
    err := r.db.Model(&models.User{}).Count(&count).Error
    return count, err
}

// CountByStatus ステータス別ユーザー数取得
func (r *userRepository) CountByStatus(verified bool) (int64, error) {
    var count int64
    query := r.db.Model(&models.User{})
    
    if verified {
        query = query.Where("email_verified_at IS NOT NULL")
    } else {
        query = query.Where("email_verified_at IS NULL")
    }
    
    err := query.Count(&count).Error
    return count, err
}