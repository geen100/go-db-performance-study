// internal/testdata/generator.go
package testdata

import (
	"fmt"
	"log"
	"math/rand"
	"strings"
	"time"

	"go-db-performance-study/internal/models"

	"github.com/brianvoe/gofakeit/v6"
	"github.com/schollz/progressbar/v3"
	"gorm.io/gorm"
)

// DataGenerator テストデータ生成器
type DataGenerator struct {
	db     *gorm.DB
	faker  *gofakeit.Faker
	config GeneratorConfig
}

// GeneratorConfig 生成設定
type GeneratorConfig struct {
	UserCount    int
	PostCount    int
	TagCount     int
	CommentCount int
	BatchSize    int
}

// NewDataGenerator 新しいデータ生成器を作成
func NewDataGenerator(db *gorm.DB, config GeneratorConfig) *DataGenerator {
	return &DataGenerator{
		db:     db,
		faker:  gofakeit.New(0),
		config: config,
	}
}

// GenerateAll 全データを生成
func (g *DataGenerator) GenerateAll() error {
	log.Println("=== テストデータ生成開始 ===")

	if err := g.GenerateUsers(); err != nil {
		return fmt.Errorf("ユーザー生成エラー: %w", err)
	}

	if err := g.GenerateTags(); err != nil {
		return fmt.Errorf("タグ生成エラー: %w", err)
	}

	if err := g.GeneratePosts(); err != nil {
		return fmt.Errorf("投稿生成エラー: %w", err)
	}

	if err := g.GenerateComments(); err != nil {
		return fmt.Errorf("コメント生成エラー: %w", err)
	}

	if err := g.AssignTagsToPosts(); err != nil {
		return fmt.Errorf("タグ関連付けエラー: %w", err)
	}

	log.Println("=== テストデータ生成完了 ===")
	return nil
}

// ----------------- ユーザー生成 -----------------
func (g *DataGenerator) GenerateUsers() error {
	log.Printf("ユーザー生成中: %d件", g.config.UserCount)

	bar := progressbar.Default(int64(g.config.UserCount))
	batchSize := g.config.BatchSize

	for i := 0; i < g.config.UserCount; i += batchSize {
		users := make([]models.User, 0, batchSize)

		for j := 0; j < batchSize && i+j < g.config.UserCount; j++ {
			users = append(users, models.User{
				Name:            g.faker.Name(),
				Email:           g.faker.Email(),
				Password:        "password123",
				EmailVerifiedAt: g.randomTimePointer(),
			})
			bar.Add(1)
		}

		if err := g.db.CreateInBatches(users, batchSize).Error; err != nil {
			return err
		}
	}

	return nil
}

// ----------------- タグ生成 -----------------
func (g *DataGenerator) GenerateTags() error {
	log.Printf("タグ生成中: %d件", g.config.TagCount)

	var existing []models.Tag
	if err := g.db.Select("name").Find(&existing).Error; err != nil {
		return fmt.Errorf("既存タグ取得エラー: %w", err)
	}

	used := make(map[string]bool)
	for _, t := range existing {
		used[t.Name] = true
	}

	techTags := []string{
		"Go", "Python", "JavaScript", "React", "Vue.js", "Node.js",
		"MySQL", "PostgreSQL", "Redis", "Docker", "Kubernetes",
		"AWS", "GCP", "Azure", "Linux", "Git", "CI/CD",
		"Machine Learning", "AI", "Blockchain", "Web Development",
		"Mobile Development", "DevOps", "Security", "Testing",
	}

	colors := []string{
		"#007bff", "#28a745", "#dc3545", "#ffc107", "#17a2b8",
		"#6f42c1", "#e83e8c", "#fd7e14", "#20c997", "#6c757d",
	}

	tags := make([]models.Tag, 0, g.config.TagCount)

	for _, name := range techTags {
		if len(tags) >= g.config.TagCount {
			break
		}
		if !used[name] {
			used[name] = true
			tags = append(tags, models.Tag{
				Name:  name,
				Color: colors[rand.Intn(len(colors))],
			})
		}
	}

	for len(tags) < g.config.TagCount {
		name := g.faker.Word() + " " + g.faker.Word()
		if !used[name] {
			used[name] = true
			tags = append(tags, models.Tag{
				Name:  name,
				Color: colors[rand.Intn(len(colors))],
			})
		}
	}

	bar := progressbar.Default(int64(len(tags)))
	for i := 0; i < len(tags); i += g.config.BatchSize {
		end := i + g.config.BatchSize
		if end > len(tags) {
			end = len(tags)
		}
		if err := g.db.CreateInBatches(tags[i:end], g.config.BatchSize).Error; err != nil {
			return fmt.Errorf("タグ挿入エラー: %w", err)
		}
		bar.Add(end - i)
	}

	return nil
}

// ----------------- 投稿生成（slug対応） -----------------
func (g *DataGenerator) GeneratePosts() error {
	log.Printf("投稿生成中: %d件", g.config.PostCount)

	bar := progressbar.Default(int64(g.config.PostCount))
	batchSize := g.config.BatchSize

	templates := g.getPostTemplates()
	statuses := []models.PostStatus{
		models.PostStatusDraft,
		models.PostStatusPublished,
		models.PostStatusArchived,
	}
	weights := []int{20, 70, 10}

	for i := 0; i < g.config.PostCount; i += batchSize {
		posts := make([]models.Post, 0, batchSize)

		for j := 0; j < batchSize && i+j < g.config.PostCount; j++ {
			template := templates[rand.Intn(len(templates))]
			status := g.weightedRandomStatus(statuses, weights)
			title := g.generateTitle(template.Category)
			body := g.generateBody(template)

			// Excerpt: Body の先頭 100 文字（UTF-8 安全）
			excerpt := string([]rune(body)[:min(100, len([]rune(body)))])

			posts = append(posts, models.Post{
				UserID:    uint(rand.Intn(g.config.UserCount) + 1),
				Title:     title,
				Body:      body,
				Excerpt:   excerpt,
				Status:    status,
				ViewCount: uint(rand.Intn(1000)),
				CreatedAt: g.randomPastTime(365),
				Slug:      g.generateUniqueSlug(title), // ユニーク slug
			})
			bar.Add(1)
		}

		if err := g.db.CreateInBatches(posts, batchSize).Error; err != nil {
			return fmt.Errorf("投稿挿入エラー: %w", err)
		}
	}

	return nil
}

func (g *DataGenerator) GenerateComments() error {
	log.Printf("コメント生成中: %d件", g.config.CommentCount)

	bar := progressbar.Default(int64(g.config.CommentCount))
	batchSize := g.config.BatchSize

	// DB に存在する Post ID を取得
	var postIDs []uint
	if err := g.db.Model(&models.Post{}).Pluck("id", &postIDs).Error; err != nil {
		return fmt.Errorf("投稿ID取得エラー: %w", err)
	}
	if len(postIDs) == 0 {
		return fmt.Errorf("コメント生成用の投稿が存在しません")
	}

	commentTemplates := []string{
		"とても参考になりました！ありがとうございます。",
		"詳しい説明ありがとうございます。実際に試してみます。",
		"この方法は知りませんでした。勉強になります。",
		"素晴らしい記事ですね。続編も期待しています。",
		"実装例があるとより理解しやすいかもしれません。",
		"同じような問題に遭遇していたので、とても助かりました。",
		"別のアプローチも紹介していただけると嬉しいです。",
		"初心者にもわかりやすい説明で良かったです。",
	}

	statuses := []models.CommentStatus{
		models.CommentStatusPending,
		models.CommentStatusApproved,
		models.CommentStatusSpam,
	}
	weights := []int{20, 70, 10}

	for i := 0; i < g.config.CommentCount; i += batchSize {
		comments := make([]models.Comment, 0, batchSize)

		for j := 0; j < batchSize && i+j < g.config.CommentCount; j++ {
			template := commentTemplates[rand.Intn(len(commentTemplates))]
			body := template + "\n\n" + g.faker.Sentence(10)

			comments = append(comments, models.Comment{
				PostID:    postIDs[rand.Intn(len(postIDs))], // DB から取得した ID を使用
				UserID:    uint(rand.Intn(g.config.UserCount) + 1),
				Body:      body,
				Status:    g.weightedRandomCommentStatus(statuses, weights),
				CreatedAt: g.randomPastTime(180),
			})
			bar.Add(1)
		}

		if err := g.db.CreateInBatches(comments, batchSize).Error; err != nil {
			return err
		}
	}

	return nil
}

// ----------------- ヘルパー: 重み付きランダムでコメントステータス生成 -----------------
func (g *DataGenerator) weightedRandomCommentStatus(statuses []models.CommentStatus, weights []int) models.CommentStatus {
	total := 0
	for _, w := range weights {
		total += w
	}

	r := rand.Intn(total)
	for i, w := range weights {
		if r < w {
			return statuses[i]
		}
		r -= w
	}

	return statuses[0]
}

// ----------------- 投稿とタグ関連付け -----------------
func (g *DataGenerator) AssignTagsToPosts() error {
	log.Println("投稿とタグの関連付け中...")

	var posts []models.Post
	if err := g.db.Find(&posts).Error; err != nil {
		return err
	}

	var tags []models.Tag
	if err := g.db.Find(&tags).Error; err != nil {
		return err
	}

	bar := progressbar.Default(int64(len(posts)))

	for _, post := range posts {
		tagCount := rand.Intn(5) + 1
		selectedTags := make([]models.Tag, 0, tagCount)

		for i := 0; i < tagCount; i++ {
			tag := tags[rand.Intn(len(tags))]
			selectedTags = append(selectedTags, tag)
		}

		if err := g.db.Model(&post).Association("Tags").Replace(selectedTags); err != nil {
			return err
		}

		bar.Add(1)
	}

	return nil
}

// ----------------- ヘルパーメソッド -----------------
func (g *DataGenerator) randomTimePointer() *time.Time {
	if rand.Float32() < 0.8 {
		t := g.randomPastTime(30)
		return &t
	}
	return nil
}

func (g *DataGenerator) randomPastTime(days int) time.Time {
	return time.Now().AddDate(0, 0, -rand.Intn(days))
}

func (g *DataGenerator) weightedRandomStatus(statuses []models.PostStatus, weights []int) models.PostStatus {
	total := 0
	for _, w := range weights {
		total += w
	}

	r := rand.Intn(total)
	for i, w := range weights {
		if r < w {
			return statuses[i]
		}
		r -= w
	}

	return statuses[0]
}

// ----------------- 投稿テンプレート -----------------
type PostTemplate struct {
	Category string
	Content  []string
}

func (g *DataGenerator) getPostTemplates() []PostTemplate {
	return []PostTemplate{
		{
			Category: "Go言語",
			Content: []string{
				"Go言語でWebアプリケーションを開発する方法について詳しく解説します。",
				"基本的なHTTPサーバーの作成から、ルーティング、ミドルウェアの実装まで。",
				"実際のコード例も交えて、実践的な内容をお届けします。",
			},
		},
		{
			Category: "データベース",
			Content: []string{
				"データベース設計のベストプラクティスをまとめました。",
				"正規化、インデックス、クエリ最適化について。",
				"実際のパフォーマンス測定結果も含めて紹介します。",
			},
		},
	}
}

func (g *DataGenerator) generateTitle(category string) string {
	prefixes := []string{
		"初心者向け", "実践的な", "効率的な", "最新の", "詳解",
		"完全攻略", "基礎から学ぶ", "プロが教える", "実例で学ぶ",
	}

	suffixes := []string{
		"入門", "基礎講座", "実践ガイド", "チュートリアル", "まとめ",
		"解説", "手順", "方法", "テクニック", "ノウハウ",
	}

	prefix := prefixes[rand.Intn(len(prefixes))]
	suffix := suffixes[rand.Intn(len(suffixes))]

	return fmt.Sprintf("%s %s %s", prefix, category, suffix)
}

func (g *DataGenerator) generateBody(template PostTemplate) string {
	content := ""
	for _, paragraph := range template.Content {
		content += paragraph + "\n\n"
	}

	content += g.faker.Paragraph(3, 5, 10, "\n\n")
	return content
}

// ----------------- ユニーク slug 生成 -----------------
func (g *DataGenerator) generateUniqueSlug(title string) string {
	timestamp := time.Now().UnixNano()
	randomNum := rand.Intn(10000)
	slug := fmt.Sprintf("%s-%d-%d", title, timestamp, randomNum)
	slug = strings.ToLower(slug)
	slug = strings.ReplaceAll(slug, " ", "-")
	slug = strings.ReplaceAll(slug, "/", "-")
	return slug
}

// ----------------- ユーティリティ -----------------
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
