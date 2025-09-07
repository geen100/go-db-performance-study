package database

import (
    "fmt"
    "log"

    "go-db-performance-study/internal/config"
    
    "gorm.io/driver/mysql"
    "gorm.io/gorm"
    "gorm.io/gorm/logger"
)

// DB データベース接続のグローバル変数
var DB *gorm.DB

// Connect データベースに接続
func Connect(env string) (*gorm.DB, error) {
    // 設定読み込み
    cfg, err := config.LoadDatabaseConfig(env)
    if err != nil {
        return nil, fmt.Errorf("設定読み込みエラー: %w", err)
    }

    // GORM設定
    gormConfig := &gorm.Config{
        Logger: logger.Default.LogMode(logger.Info), // SQL ログを有効化
    }

    // 開発環境以外では詳細ログを無効化
    if env == "production" {
        gormConfig.Logger = logger.Default.LogMode(logger.Error)
    }

    // データベース接続
    db, err := gorm.Open(mysql.Open(cfg.DSN()), gormConfig)
    if err != nil {
        return nil, fmt.Errorf("データベース接続エラー: %w", err)
    }

    // 基盤となるSQL DBを取得
    sqlDB, err := db.DB()
    if err != nil {
        return nil, fmt.Errorf("SQL DB取得エラー: %w", err)
    }

    // コネクションプール設定
    sqlDB.SetMaxIdleConns(cfg.MaxIdleConns)
    sqlDB.SetMaxOpenConns(cfg.MaxOpenConns)
    sqlDB.SetConnMaxLifetime(cfg.ConnMaxLifetime()) // ← メソッド呼び出しで修正

    // 接続確認
    if err := sqlDB.Ping(); err != nil {
        return nil, fmt.Errorf("データベース接続確認エラー: %w", err)
    }

    log.Printf("データベースに接続しました: %s:%d/%s", cfg.Host, cfg.Port, cfg.Name)
    
    // グローバル変数に保存
    DB = db
    
    return db, nil
}

// Close データベース接続を閉じる
func Close() error {
    if DB == nil {
        return nil
    }
    
    sqlDB, err := DB.DB()
    if err != nil {
        return err
    }
    
    return sqlDB.Close()
}

// GetDB データベースインスタンスを取得
func GetDB() *gorm.DB {
    return DB
}