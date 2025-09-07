package models

import (
	"html"
	"strings"
	"time"

	"gorm.io/gorm"
)

// CommentStatus コメントステータス
type CommentStatus string

const (
    CommentStatusPending  CommentStatus = "pending"   // 承認待ち
    CommentStatusApproved CommentStatus = "approved"  // 承認済み
    CommentStatusSpam     CommentStatus = "spam"      // スパム
    CommentStatusDeleted  CommentStatus = "deleted"   // 削除済み
)

// Comment コメントモデル
type Comment struct {
    ID        uint          `gorm:"primaryKey;autoIncrement" json:"id"`
    PostID    uint          `gorm:"not null;index:idx_comment_post_id" json:"post_id" validate:"required"`
    UserID    uint          `gorm:"not null;index:idx_comment_user_id" json:"user_id" validate:"required"`
    ParentID  *uint         `gorm:"null;index:idx_comment_parent_id" json:"parent_id,omitempty"` // 返信コメント用
    Body      string        `gorm:"type:text;not null" json:"body" validate:"required,min=1,max=2000"`
    Status    CommentStatus `gorm:"size:20;not null;default:pending;index:idx_comment_status" json:"status" validate:"required,oneof=pending approved spam deleted"`
    IPAddress string        `gorm:"size:45;index:idx_comment_ip" json:"ip_address,omitempty"`
    UserAgent string        `gorm:"size:500" json:"user_agent,omitempty"`
    IsEdited  bool          `gorm:"default:false" json:"is_edited"`
    EditedAt  *time.Time    `gorm:"null" json:"edited_at,omitempty"`
    CreatedAt time.Time     `gorm:"autoCreateTime;index:idx_comment_created_at" json:"created_at"`
    UpdatedAt time.Time     `gorm:"autoUpdateTime" json:"updated_at"`

    // リレーション
    Post    Post      `gorm:"foreignKey:PostID;constraint:OnDelete:CASCADE" json:"post,omitempty"`
    User    User      `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE" json:"user,omitempty"`
    Parent  *Comment  `gorm:"foreignKey:ParentID" json:"parent,omitempty"`
    Replies []Comment `gorm:"foreignKey:ParentID" json:"replies,omitempty"`
}

// TableName テーブル名を明示的に指定
func (Comment) TableName() string {
    return "comments"
}

// CommentForCreate コメント作成用構造体
type CommentForCreate struct {
    PostID   uint   `json:"post_id" validate:"required"`
    UserID   uint   `json:"user_id" validate:"required"`
    ParentID *uint  `json:"parent_id,omitempty"`
    Body     string `json:"body" validate:"required,min=1,max=2000"`
}

// CommentForUpdate コメント更新用構造体
type CommentForUpdate struct {
    Body   *string        `json:"body,omitempty" validate:"omitempty,min=1,max=2000"`
    Status *CommentStatus `json:"status,omitempty" validate:"omitempty,oneof=pending approved spam deleted"`
}

// CommentResponse API レスポンス用構造体
type CommentResponse struct {
    ID        uint         `json:"id"`
    Body      string       `json:"body"`
    Status    CommentStatus `json:"status"`
    IsEdited  bool         `json:"is_edited"`
    EditedAt  *time.Time   `json:"edited_at,omitempty"`
    CreatedAt time.Time    `json:"created_at"`
    User      UserResponse `json:"user"`
    ParentID  *uint        `json:"parent_id,omitempty"`
    ReplyCount int         `json:"reply_count,omitempty"`
}

// CommentTree 階層構造のコメント
type CommentTree struct {
    CommentResponse
    Replies []CommentTree `json:"replies"`
}

// CommentStats コメント統計情報
type CommentStats struct {
    TotalCount    int64 `json:"total_count"`
    ApprovedCount int64 `json:"approved_count"`
    PendingCount  int64 `json:"pending_count"`
    SpamCount     int64 `json:"spam_count"`
    TodayCount    int64 `json:"today_count"`
    WeekCount     int64 `json:"week_count"`
    MonthCount    int64 `json:"month_count"`
}

// UserCommentActivity ユーザーコメント活動統計
type UserCommentActivity struct {
    UserID       uint      `json:"user_id"`
    UserName     string    `json:"user_name"`
    CommentCount int       `json:"comment_count"`
    LastComment  time.Time `json:"last_comment"`
    AvgLength    float64   `json:"avg_length"`
    SpamRatio    float64   `json:"spam_ratio"`
}

// BeforeCreate 作成前処理
func (c *Comment) BeforeCreate(tx *gorm.DB) error {
    // HTMLエスケープ処理
    c.Body = html.EscapeString(c.Body)
    
    // バリデーション実行
    return c.Validate()
}

// BeforeUpdate 更新前処理
func (c *Comment) BeforeUpdate(tx *gorm.DB) error {
    // 本文が更新された場合
    if tx.Statement.Changed("body") {
        c.Body = html.EscapeString(c.Body)
        c.IsEdited = true
        now := time.Now()
        c.EditedAt = &now
    }
    
    return nil
}

// Validate バリデーション実行
func (c *Comment) Validate() error {
    return ValidateStruct(c)
}

// IsReply 返信コメントかどうか判定
func (c *Comment) IsReply() bool {
    return c.ParentID != nil
}

// IsApproved 承認済みかどうか判定
func (c *Comment) IsApproved() bool {
    return c.Status == CommentStatusApproved
}

// IsPending 承認待ちかどうか判定
func (c *Comment) IsPending() bool {
    return c.Status == CommentStatusPending
}

// IsSpam スパムかどうか判定
func (c *Comment) IsSpam() bool {
    return c.Status == CommentStatusSpam
}

// IsDeleted 削除済みかどうか判定
func (c *Comment) IsDeleted() bool {
    return c.Status == CommentStatusDeleted
}

// GetPlainBody HTMLタグを除去した本文を取得
func (c *Comment) GetPlainBody() string {
    // 簡易的なHTMLタグ除去
    body := c.Body
    body = strings.ReplaceAll(body, "&lt;", "<")
    body = strings.ReplaceAll(body, "&gt;", ">")
    body = strings.ReplaceAll(body, "&amp;", "&")
    return body
}

// GetExcerpt 抜粋を取得
func (c *Comment) GetExcerpt(length int) string {
    plainBody := c.GetPlainBody()
    if len(plainBody) <= length {
        return plainBody
    }
    return plainBody[:length] + "..."
}

// Approve コメントを承認
func (c *Comment) Approve(tx *gorm.DB) error {
    c.Status = CommentStatusApproved
    return tx.Model(c).UpdateColumn("status", c.Status).Error
}

// MarkAsSpam スパムとしてマーク
func (c *Comment) MarkAsSpam(tx *gorm.DB) error {
    c.Status = CommentStatusSpam
    return tx.Model(c).UpdateColumn("status", c.Status).Error
}

// SoftDelete ソフトデリート
func (c *Comment) SoftDelete(tx *gorm.DB) error {
    c.Status = CommentStatusDeleted
    return tx.Model(c).UpdateColumn("status", c.Status).Error
}

// CountReplies 返信数をカウント
func (c *Comment) CountReplies(tx *gorm.DB) (int64, error) {
    var count int64
    err := tx.Model(&Comment{}).Where("parent_id = ? AND status = ?", c.ID, CommentStatusApproved).Count(&count).Error
    return count, err
}

// GetDepth コメントの階層の深さを取得
func (c *Comment) GetDepth(tx *gorm.DB) (int, error) {
    depth := 0
    currentComment := c
    
    for currentComment.ParentID != nil {
        depth++
        var parent Comment
        if err := tx.First(&parent, *currentComment.ParentID).Error; err != nil {
            return depth, err
        }
        currentComment = &parent
        
        // 無限ループ防止
        if depth > 10 {
            break
        }
    }
    
    return depth, nil
}

// DetectSpamKeywords スパムキーワード検出
func (c *Comment) DetectSpamKeywords() bool {
    spamKeywords := []string{
        "viagra", "casino", "lottery", "winner", "congratulations",
        "click here", "make money", "work from home", "free money",
        // 日本語のスパムキーワード
        "出会い", "副業", "稼げる", "無料", "限定",
    }
    
    lowerBody := strings.ToLower(c.Body)
    
    for _, keyword := range spamKeywords {
        if strings.Contains(lowerBody, keyword) {
            return true
        }
    }
    
    return false
}

// IsFromSuspiciousIP 疑わしいIPからのコメントか判定
func (c *Comment) IsFromSuspiciousIP() bool {
    // 簡易的な実装例
    // 実際にはIPレピュテーションDBやブラックリストと照合
    suspiciousPatterns := []string{
        "192.168.", // プライベートIP（実際の運用では除外）
        "10.",      // プライベートIP
        "172.",     // プライベートIP
    }
    
    for _, pattern := range suspiciousPatterns {
        if strings.HasPrefix(c.IPAddress, pattern) {
            return true
        }
    }
    
    return false
}

// CalculateSpamScore スパムスコア計算
func (c *Comment) CalculateSpamScore() float64 {
    score := 0.0
    
    // キーワードベースのスコア
    if c.DetectSpamKeywords() {
        score += 0.5
    }
    
    // IP ベースのスコア
    if c.IsFromSuspiciousIP() {
        score += 0.3
    }
    
    // 長さベースのスコア（極端に短い、または長い）
    bodyLength := len(c.Body)
    if bodyLength < 10 || bodyLength > 1500 {
        score += 0.2
    }
    
    // URL の数（多すぎる場合）
    urlCount := strings.Count(strings.ToLower(c.Body), "http")
    if urlCount > 2 {
        score += float64(urlCount) * 0.1
    }
    
    // スコアは0.0-1.0の範囲に正規化
    if score > 1.0 {
        score = 1.0
    }
    
    return score
}

// AutoModerate 自動モデレーション
func (c *Comment) AutoModerate() CommentStatus {
    spamScore := c.CalculateSpamScore()
    
    if spamScore >= 0.7 {
        return CommentStatusSpam
    } else if spamScore >= 0.4 {
        return CommentStatusPending // 手動確認が必要
    }
    
    return CommentStatusApproved // 自動承認
}