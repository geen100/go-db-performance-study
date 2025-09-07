// internal/config/config.go (修正版 - 命名競合解決)
package config

import (
    "fmt"
    "os"
    "strconv"
    "time"

    "gopkg.in/yaml.v3"
)

// DatabaseConfig データベース設定構造体
type DatabaseConfig struct {
    Host                string `yaml:"host"`
    Port                int    `yaml:"port"`
    User                string `yaml:"user"`
    Password            string `yaml:"password"`
    Name                string `yaml:"name"`
    Charset             string `yaml:"charset"`
    ParseTime           bool   `yaml:"parse_time"`
    Loc                 string `yaml:"loc"`
    MaxIdleConns        int    `yaml:"max_idle_conns"`
    MaxOpenConns        int    `yaml:"max_open_conns"`
    ConnMaxLifetimeSeconds int    `yaml:"conn_max_lifetime"`  // ← フィールド名を変更
}

// Config アプリケーション全体の設定
type Config struct {
    Development DatabaseConfig `yaml:"development"`
    Testing     DatabaseConfig `yaml:"testing"`
    Production  DatabaseConfig `yaml:"production"`
}

// LoadDatabaseConfig データベース設定を読み込み
func LoadDatabaseConfig(env string) (*DatabaseConfig, error) {
    // 本番環境の場合は環境変数から直接読み込み
    if env == "production" {
        return loadFromEnv()
    }

    // 開発・テスト環境はYAMLファイルから読み込み
    configFile := "configs/database.yaml"
    
    data, err := os.ReadFile(configFile)
    if err != nil {
        return nil, fmt.Errorf("設定ファイルの読み込みエラー: %w", err)
    }

    var config Config
    if err := yaml.Unmarshal(data, &config); err != nil {
        return nil, fmt.Errorf("YAML解析エラー: %w", err)
    }

    var dbConfig *DatabaseConfig
    switch env {
    case "development":
        dbConfig = &config.Development
    case "testing":
        dbConfig = &config.Testing
    default:
        return nil, fmt.Errorf("サポートされていない環境: %s", env)
    }

    return dbConfig, nil
}

// loadFromEnv 環境変数からデータベース設定を読み込み
func loadFromEnv() (*DatabaseConfig, error) {
    config := &DatabaseConfig{
        Host:                   getEnvOrDefault("DB_HOST", "localhost"),
        User:                   getEnvOrDefault("DB_USER", "root"),
        Password:               getEnvOrDefault("DB_PASSWORD", "password"),
        Name:                   getEnvOrDefault("DB_NAME", "blog_benchmark"),
        Charset:                getEnvOrDefault("DB_CHARSET", "utf8mb4"),
        ParseTime:              true,
        Loc:                    getEnvOrDefault("DB_LOC", "Local"),
        MaxIdleConns:           getEnvIntOrDefault("DB_MAX_IDLE_CONNS", 20),
        MaxOpenConns:           getEnvIntOrDefault("DB_MAX_OPEN_CONNS", 200),
        ConnMaxLifetimeSeconds: getEnvIntOrDefault("DB_CONN_MAX_LIFETIME", 3600),
    }

    var err error
    config.Port, err = strconv.Atoi(getEnvOrDefault("DB_PORT", "3306"))
    if err != nil {
        return nil, fmt.Errorf("DB_PORT環境変数の変換エラー: %w", err)
    }

    return config, nil
}

// getEnvOrDefault 環境変数を取得、なければデフォルト値
func getEnvOrDefault(key, defaultValue string) string {
    if value := os.Getenv(key); value != "" {
        return value
    }
    return defaultValue
}

// getEnvIntOrDefault 環境変数をint型で取得、なければデフォルト値
func getEnvIntOrDefault(key string, defaultValue int) int {
    if value := os.Getenv(key); value != "" {
        if intValue, err := strconv.Atoi(value); err == nil {
            return intValue
        }
    }
    return defaultValue
}

// DSN データソース名を生成
func (c *DatabaseConfig) DSN() string {
    return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=%s&parseTime=%t&loc=%s",
        c.User, c.Password, c.Host, c.Port, c.Name,
        c.Charset, c.ParseTime, c.Loc)
}

// ConnMaxLifetime 接続最大生存時間をtime.Duration型で取得
func (c *DatabaseConfig) ConnMaxLifetime() time.Duration {
    return time.Duration(c.ConnMaxLifetimeSeconds) * time.Second  // ← フィールド名を変更
}