// internal/testdata/generator.go
package testdata

import (
    "fmt"
    "log"
    "math/rand"
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

// GenerateUsers ユーザー生成
func (g *DataGenerator) GenerateUsers() error {
    log.Printf("ユーザー生成中: %d件", g.config.UserCount)
    
    bar := progressbar.Default(int64(g.config.UserCount))
    batchSize := g.config.BatchSize
    
    for i := 0; i < g.config.UserCount; i += batchSize {
        users := make([]models.User, 0, batchSize)
        
        for j := 0; j < batchSize && i+j < g.config.UserCount; j++ {
            users = append(users, models.User{
                Name:     g.faker.Name(),
                Email:    g.faker.Email(),
                Password: "password123", // BeforeCreateでハッシュ化される
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

// GenerateTags タグ生成
func (g *DataGenerator) GenerateTags() error {
    log.Printf("タグ生成中: %d件", g.config.TagCount)
    
    // 技術系タグのテンプレート
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
    
    for i := 0; i < g.config.TagCount; i++ {
        var tagName string
        if i < len(techTags) {
            tagName = techTags[i]
        } else {
            tagName = g.faker.Word() + " " + g.faker.Word()
        }
        
        tags = append(tags, models.Tag{
            Name:  tagName,
            Color: colors[rand.Intn(len(colors))],
        })
    }
    
    return g.db.CreateInBatches(tags, g.config.BatchSize).Error
}

// GeneratePosts 投稿生成
func (g *DataGenerator) GeneratePosts() error {
    log.Printf("投稿生成中: %d件", g.config.PostCount)
    
    bar := progressbar.Default(int64(g.config.PostCount))
    batchSize := g.config.BatchSize
    
    // 投稿テンプレート
    templates := g.getPostTemplates()
    statuses := []models.PostStatus{
        models.PostStatusDraft,
        models.PostStatusPublished,
        models.PostStatusArchived,
    }
    weights := []int{20, 70, 10} // 公開:70%, 下書き:20%, アーカイブ:10%
    
    for i := 0; i < g.config.PostCount; i += batchSize {
        posts := make([]models.Post, 0, batchSize)
        
        for j := 0; j < batchSize && i+j < g.config.PostCount; j++ {
            template := templates[rand.Intn(len(templates))]
            status := g.weightedRandomStatus(statuses, weights)
            
            posts = append(posts, models.Post{
                UserID:    uint(rand.Intn(g.config.UserCount) + 1),
                Title:     g.generateTitle(template.Category),
                Body:      g.generateBody(template),
                Status:    status,
                ViewCount: uint(rand.Intn(1000)),
                CreatedAt: g.randomPastTime(365), // 過去1年間
            })
            bar.Add(1)
        }
        
        if err := g.db.CreateInBatches(posts, batchSize).Error; err != nil {
            return err
        }
    }
    
    return nil
}

// GenerateComments コメント生成
func (g *DataGenerator) GenerateComments() error {
    log.Printf("コメント生成中: %d件", g.config.CommentCount)
    
    bar := progressbar.Default(int64(g.config.CommentCount))
    batchSize := g.config.BatchSize
    
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
    
    for i := 0; i < g.config.CommentCount; i += batchSize {
        comments := make([]models.Comment, 0, batchSize)
        
        for j := 0; j < batchSize && i+j < g.config.CommentCount; j++ {
            template := commentTemplates[rand.Intn(len(commentTemplates))]
            
            comments = append(comments, models.Comment{
                PostID:    uint(rand.Intn(g.config.PostCount) + 1),
                UserID:    uint(rand.Intn(g.config.UserCount) + 1),
                Body:      template + "\n\n" + g.faker.Sentence(10),
                CreatedAt: g.randomPastTime(180), // 過去6ヶ月間
            })
            bar.Add(1)
        }
        
        if err := g.db.CreateInBatches(comments, batchSize).Error; err != nil {
            return err
        }
    }
    
    return nil
}

// AssignTagsToPosts 投稿にタグを関連付け
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
        // ランダムに1-5個のタグを関連付け
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

// ヘルパーメソッド
func (g *DataGenerator) randomTimePointer() *time.Time {
    if rand.Float32() < 0.8 { // 80%の確率で認証済み
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

// PostTemplate 投稿テンプレート
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
        // 他のテンプレートも追加...
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
    
    // ランダムに追加コンテンツを生成
    content += g.faker.Paragraph(3, 5, 10, "\n\n")
    
    return content
}