// internal/repository/interfaces/user.go
package interfaces

import (
    "go-db-performance-study/internal/models"
)

// UserRepository ユーザーリポジトリインターフェース
type UserRepository interface {
    // 基本CRUD
    Create(user *models.User) error
    GetByID(id uint) (*models.User, error)
    GetByEmail(email string) (*models.User, error)
    Update(id uint, updates *models.UserForUpdate) error
    Delete(id uint) error
    
    // 一覧取得
    List(limit, offset int) ([]models.User, error)
    ListWithStats(limit, offset int) ([]models.UserStats, error)
    
    // 検索
    Search(query string, limit, offset int) ([]models.User, error)
    GetActiveUsers(limit int) ([]models.User, error)
    
    // 統計
    Count() (int64, error)
    CountByStatus(verified bool) (int64, error)
}