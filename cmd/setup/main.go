package main

import (
	"log"
	"os"

	"go-db-performance-study/internal/database"
	"go-db-performance-study/internal/models"

	"gorm.io/gorm" // ← この行を追加
)

func main() {
	log.Println("=== データベースセットアップ開始 ===")

	// 環境取得（デフォルトは development）
	env := os.Getenv("APP_ENV")
	if env == "" {
		env = "development"
	}
	log.Printf("環境: %s", env)

	// データベース接続
	db, err := database.Connect(env)
	if err != nil {
		log.Fatalf("データベース接続エラー: %v", err)
	}
	defer database.Close()

	// マイグレーション実行
	if err := database.Migrate(db); err != nil {
		log.Fatalf("マイグレーションエラー: %v", err)
	}

	// テーブル存在確認
	if err := database.CheckTablesExist(db); err != nil {
		log.Fatalf("テーブル確認エラー: %v", err)
	}

	// テストユーザー作成
	if err := createTestUser(db); err != nil {
		log.Fatalf("テストユーザー作成エラー: %v", err)
	}

	// テスト投稿作成
	if err := createTestPost(db); err != nil {
		log.Fatalf("テスト投稿作成エラー: %v", err)
	}

	log.Println("=== セットアップ完了 ===")
}

// createTestUser テストユーザーを作成
func createTestUser(db *gorm.DB) error {
	user := &models.User{
		Name:     "テストユーザー",
		Email:    "test@example.com",
		Password: "password123", // 本来はハッシュ化が必要
	}

	// 既存ユーザーがいるかチェック
	var existingUser models.User
	if err := db.Where("email = ?", user.Email).First(&existingUser).Error; err == nil {
		log.Printf("ユーザー '%s' は既に存在します", user.Email)
		return nil
	}

	if err := db.Create(user).Error; err != nil {
		return err
	}

	log.Printf("テストユーザーを作成しました: ID=%d, Name=%s, Email=%s",
		user.ID, user.Name, user.Email)
	return nil
}

// createTestPost テスト投稿を作成
func createTestPost(db *gorm.DB) error {
	// ユーザーを取得
	var user models.User
	if err := db.Where("email = ?", "test@example.com").First(&user).Error; err != nil {
		return err
	}

	post := &models.Post{
		UserID: user.ID,
		Title:  "最初のブログ投稿",
		Body:   "これはGORM性能テスト用の最初の投稿です。\nデータベース操作が正常に動作しているかを確認します。",
	}

	if err := db.Create(post).Error; err != nil {
		return err
	}

	log.Printf("テスト投稿を作成しました: ID=%d, Title=%s, UserID=%d",
		post.ID, post.Title, post.UserID)
	return nil
}
