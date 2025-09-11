// internal/testdata/generator_safe.go
package testdata

import (
	"fmt"
	"math/rand"

	"go-db-performance-study/internal/models"
)

// -----------------------
// DataGenerator に追加するメソッド
// -----------------------

// GenerateAllSafe 全データを生成（投稿生成時に文字化け対策済み）
func (g *DataGenerator) GenerateAllSafe() error {
	if err := g.GenerateUsers(); err != nil {
		return fmt.Errorf("ユーザー生成エラー: %w", err)
	}

	if err := g.GenerateTags(); err != nil {
		return fmt.Errorf("タグ生成エラー: %w", err)
	}

	if err := g.GeneratePostsSafe(); err != nil {
		return fmt.Errorf("投稿生成エラー: %w", err)
	}

	if err := g.GenerateComments(); err != nil {
		return fmt.Errorf("コメント生成エラー: %w", err)
	}

	if err := g.AssignTagsToPosts(); err != nil {
		return fmt.Errorf("タグ関連付けエラー: %w", err)
	}

	return nil
}

// GeneratePostsSafe 投稿生成（文字列長・文字コードを安全に調整）
func (g *DataGenerator) GeneratePostsSafe() error {
	batchSize := g.config.BatchSize
	for i := 0; i < g.config.PostCount; i += batchSize {
		posts := make([]models.Post, 0, batchSize)
		for j := 0; j < batchSize && i+j < g.config.PostCount; j++ {
			post := g.generateSinglePost()
			posts = append(posts, post)
		}

		if err := g.db.CreateInBatches(posts, batchSize).Error; err != nil {
			return err
		}
	}

	return nil
}

// generateSinglePost 単一の投稿データを生成（文字コード安全）
func (g *DataGenerator) generateSinglePost() models.Post {
	templates := g.getPostTemplates()
	template := templates[rand.Intn(len(templates))]

	// タイトル・本文を utf8mb4 に収まるように調整
	title := g.generateTitle(template.Category)
	if len(title) > 255 {
		title = title[:255]
	}

	body := g.generateBody(template)
	if len(body) > 65535 { // TEXT 型の上限
		body = body[:65535]
	}

	excerpt := ""
	if len(body) > 200 {
		excerpt = body[:200]
	} else {
		excerpt = body
	}

	statuses := []models.PostStatus{
		models.PostStatusDraft,
		models.PostStatusPublished,
		models.PostStatusArchived,
	}
	weights := []int{20, 70, 10}

	return models.Post{
		UserID:    uint(rand.Intn(g.config.UserCount) + 1),
		Title:     title,
		Body:      body,
		Excerpt:   excerpt,
		Status:    g.weightedRandomStatus(statuses, weights),
		ViewCount: uint(rand.Intn(1000)),
		CreatedAt: g.randomPastTime(365),
	}
}
