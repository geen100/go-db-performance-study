package main

import (
	"fmt"
	"log"
	"time"

	"go-db-performance-study/internal/database"
	"go-db-performance-study/internal/models"
	gorm_repo "go-db-performance-study/internal/repository/gorm"
	"go-db-performance-study/internal/repository/interfaces"
)

func main() {
	log.Println("=== CRUD操作テスト開始 ===")

	// データベース接続
	db, err := database.Connect("development")
	if err != nil {
		log.Fatalf("データベース接続エラー: %v", err)
	}
	defer database.Close()

	// リポジトリ作成
	userRepo := gorm_repo.NewUserRepository(db)
	postRepo := gorm_repo.NewPostRepository(db)

	// ユーザーCRUDテスト
	if err := testUserCRUD(userRepo); err != nil {
		log.Fatalf("ユーザーCRUDテストエラー: %v", err)
	}

	// 投稿CRUDテスト
	if err := testPostCRUD(postRepo, userRepo); err != nil {
		log.Fatalf("投稿CRUDテストエラー: %v", err)
	}

	log.Println("=== 全テスト完了 ===")
}

func testUserCRUD(repo interfaces.UserRepository) error {
	log.Println("--- ユーザーCRUDテスト ---")

	// Create（メールをユニークに）
	email := fmt.Sprintf("crud-test-%d@example.com", time.Now().UnixNano())
	user := &models.User{
		Name:     "テストユーザー",
		Email:    email,
		Password: "password123",
	}

	if err := repo.Create(user); err != nil {
		return fmt.Errorf("ユーザー作成エラー: %w", err)
	}
	log.Printf("✅ ユーザー作成: ID=%d, Email=%s", user.ID, user.Email)

	// Read
	retrievedUser, err := repo.GetByID(user.ID)
	if err != nil {
		return fmt.Errorf("ユーザー取得エラー: %w", err)
	}
	log.Printf("✅ ユーザー取得: Name=%s, Email=%s", retrievedUser.Name, retrievedUser.Email)

	// Update
	updates := &models.UserForUpdate{
		Name: ptr("更新されたユーザー"),
	}

	if err := repo.Update(user.ID, updates); err != nil {
		return fmt.Errorf("ユーザー更新エラー: %w", err)
	}
	log.Printf("✅ ユーザー更新完了")

	// 更新確認
	updatedUser, err := repo.GetByID(user.ID)
	if err != nil {
		return fmt.Errorf("更新後ユーザー取得エラー: %w", err)
	}
	log.Printf("✅ 更新確認: Name=%s", updatedUser.Name)

	// List
	users, err := repo.List(10, 0)
	if err != nil {
		return fmt.Errorf("ユーザー一覧取得エラー: %w", err)
	}
	log.Printf("✅ ユーザー一覧取得: %d件", len(users))

	// Count
	count, err := repo.Count()
	if err != nil {
		return fmt.Errorf("ユーザー数取得エラー: %w", err)
	}
	log.Printf("✅ ユーザー総数: %d件", count)

	return nil
}

func testPostCRUD(postRepo interfaces.PostRepository, userRepo interfaces.UserRepository) error {
	log.Println("--- 投稿CRUDテスト ---")

	// テスト用ユーザー取得
	users, err := userRepo.List(1, 0)
	if err != nil || len(users) == 0 {
		return fmt.Errorf("テスト用ユーザーが見つかりません")
	}
	userID := users[0].ID

	// Create（タイトルをユニークに）
	post := &models.Post{
		UserID: userID,
		Title:  fmt.Sprintf("CRUDテスト投稿 %d", time.Now().UnixNano()),
		Body:   "これはCRUD操作のテスト投稿です。",
		Status: models.PostStatusDraft,
	}

	if err := postRepo.Create(post); err != nil {
		return fmt.Errorf("投稿作成エラー: %w", err)
	}
	log.Printf("✅ 投稿作成: ID=%d, Title=%s", post.ID, post.Title)

	// Read
	retrievedPost, err := postRepo.GetByID(post.ID)
	if err != nil {
		return fmt.Errorf("投稿取得エラー: %w", err)
	}
	log.Printf("✅ 投稿取得: Title=%s, User=%s", retrievedPost.Title, retrievedPost.User.Name)

	// Update
	newStatus := models.PostStatusPublished
	updates := &models.PostForUpdate{
		Title:  ptr(fmt.Sprintf("更新されたタイトル %d", time.Now().UnixNano())),
		Status: &newStatus,
	}

	if err := postRepo.Update(post.ID, updates); err != nil {
		return fmt.Errorf("投稿更新エラー: %w", err)
	}
	log.Printf("✅ 投稿更新完了")

	// List
	posts, err := postRepo.List(10, 0)
	if err != nil {
		return fmt.Errorf("投稿一覧取得エラー: %w", err)
	}
	log.Printf("✅ 投稿一覧取得: %d件", len(posts))

	// Search
	searchResults, err := postRepo.Search("更新", 10, 0)
	if err != nil {
		return fmt.Errorf("投稿検索エラー: %w", err)
	}
	log.Printf("✅ 投稿検索: %d件ヒット", len(searchResults))

	return nil
}

// ptr 文字列のポインタを返すヘルパー関数
func ptr(s string) *string {
	return &s
}
