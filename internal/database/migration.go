// internal/database/migration.go
package database

import (
    "fmt"
    "log"

    "go-db-performance-study/internal/models"
    
    "gorm.io/gorm"
)

// Migrate データベースマイグレーションを実行
func Migrate(db *gorm.DB) error {
    log.Println("データベースマイグレーションを開始...")

    // モデルの順序が重要（外部キー依存関係を考慮）
    err := db.AutoMigrate(
        &models.User{},    // 1. 最初にユーザー
        &models.Tag{},     // 2. タグ（独立）
        &models.Post{},    // 3. 投稿（ユーザーに依存）
        &models.Comment{}, // 4. コメント（ユーザー・投稿に依存）
        // 多対多の中間テーブル（post_tags）は自動作成される
    )

    if err != nil {
        return fmt.Errorf("マイグレーションエラー: %w", err)
    }

    log.Println("データベースマイグレーションが完了しました")
    return nil
}

// DropAllTables 全テーブルを削除（テスト用）
func DropAllTables(db *gorm.DB) error {
    log.Println("全テーブルを削除中...")

    // 外部キー制約の関係で削除順序が重要
    tables := []interface{}{
        &models.Comment{},
        &models.Post{},
        &models.Tag{},
        &models.User{},
    }

    for _, table := range tables {
        if err := db.Migrator().DropTable(table); err != nil {
            return fmt.Errorf("テーブル削除エラー: %w", err)
        }
    }

    // 多対多中間テーブルも削除
    if err := db.Migrator().DropTable("post_tags"); err != nil {
        // エラーがあっても続行（テーブルが存在しない場合もある）
        log.Printf("中間テーブル削除時の警告: %v", err)
    }

    log.Println("全テーブルの削除が完了しました")
    return nil
}

// CheckTablesExist テーブルの存在確認
func CheckTablesExist(db *gorm.DB) error {
    tables := []string{"users", "posts", "tags", "comments", "post_tags"}
    
    for _, table := range tables {
        if !db.Migrator().HasTable(table) {
            return fmt.Errorf("テーブル '%s' が存在しません", table)
        }
    }
    
    log.Println("全テーブルの存在を確認しました")
    return nil
}