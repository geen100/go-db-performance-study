// cmd/generate-data/main.go
package main

import (
	"flag"
	"fmt"
	"log"
	"time"

	"go-db-performance-study/internal/database"
	"go-db-performance-study/internal/models"
	"go-db-performance-study/internal/testdata"
	"go-db-performance-study/internal/testdata/scenarios"

	"gorm.io/gorm"
)

func main() {
	var (
		scenario = flag.String("scenario", "small", "データセット規模 (small/medium/large/custom)")
		users    = flag.Int("users", 1000, "生成するユーザー数（customの場合）")
		posts    = flag.Int("posts", 5000, "生成する投稿数（customの場合）")
		tags     = flag.Int("tags", 100, "生成するタグ数（customの場合）")
		comments = flag.Int("comments", 10000, "生成するコメント数（customの場合）")
		env      = flag.String("env", "development", "環境 (development/testing)")
		clean    = flag.Bool("clean", false, "既存データを削除してから実行")
	)
	flag.Parse()

	log.Printf("=== テストデータ生成ツール ===")
	log.Printf("シナリオ: %s", *scenario)
	log.Printf("環境: %s", *env)

	// データベース接続
	db, err := database.Connect(*env)
	if err != nil {
		log.Fatalf("データベース接続エラー: %v", err)
	}
	defer database.Close()

	// 既存データクリーンアップ
	if *clean {
		log.Println("既存データを削除中...")
		if err := cleanupData(db); err != nil {
			log.Fatalf("データクリーンアップエラー: %v", err)
		}
	}

	startTime := time.Now()

	// シナリオ別実行
	switch *scenario {
	case "small":
		err = scenarios.GenerateSmallDataset(db)
	case "medium":
		err = scenarios.GenerateMediumDataset(db)
	case "large":
		err = scenarios.GenerateLargeDataset(db)
	case "custom":
		config := testdata.GeneratorConfig{
			UserCount:    *users,
			PostCount:    *posts,
			TagCount:     *tags,
			CommentCount: *comments,
			BatchSize:    1000,
		}
		generator := testdata.NewDataGenerator(db, config)
		err = generator.GenerateAllSafe() // ← UTF-8安全版メソッド
	default:
		log.Fatalf("未知のシナリオ: %s", *scenario)
	}

	if err != nil {
		log.Fatalf("データ生成エラー: %v", err)
	}

	duration := time.Since(startTime)
	log.Printf("=== データ生成完了 ===")
	log.Printf("実行時間: %v", duration)

	if err := showDataCounts(db); err != nil {
		log.Printf("データ件数取得エラー: %v", err)
	}
}

// cleanupData 既存データを削除
func cleanupData(db *gorm.DB) error {
	tables := []string{"comments", "post_tags", "posts", "tags", "users"}

	for _, table := range tables {
		if err := db.Exec(fmt.Sprintf("DELETE FROM %s", table)).Error; err != nil {
			return fmt.Errorf("テーブル %s のクリーンアップエラー: %w", table, err)
		}
		if err := db.Exec(fmt.Sprintf("ALTER TABLE %s AUTO_INCREMENT = 1", table)).Error; err != nil {
			log.Printf("AUTO_INCREMENT リセット警告 (%s): %v", table, err)
		}
	}

	log.Println("既存データの削除が完了しました")
	return nil
}

// showDataCounts 生成されたデータ件数を表示
func showDataCounts(db *gorm.DB) error {
	var counts struct {
		Users    int64
		Posts    int64
		Tags     int64
		Comments int64
		PostTags int64
	}

	db.Model(&models.User{}).Count(&counts.Users)
	db.Model(&models.Post{}).Count(&counts.Posts)
	db.Model(&models.Tag{}).Count(&counts.Tags)
	db.Model(&models.Comment{}).Count(&counts.Comments)
	db.Table("post_tags").Count(&counts.PostTags)

	log.Printf("生成されたデータ件数:")
	log.Printf("  ユーザー: %d件", counts.Users)
	log.Printf("  投稿: %d件", counts.Posts)
	log.Printf("  タグ: %d件", counts.Tags)
	log.Printf("  コメント: %d件", counts.Comments)
	log.Printf("  投稿-タグ関係: %d件", counts.PostTags)

	return nil
}
