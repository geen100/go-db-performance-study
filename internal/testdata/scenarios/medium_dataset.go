// internal/testdata/scenarios/medium_dataset.go
package scenarios

import (
    "go-db-performance-study/internal/testdata"
    "gorm.io/gorm"
)

// GenerateMediumDataset 中規模データセット生成
func GenerateMediumDataset(db *gorm.DB) error {
    config := testdata.GeneratorConfig{
        UserCount:    10000,
        PostCount:    50000,
        TagCount:     500,
        CommentCount: 200000,
        BatchSize:    1000,
    }
    
    generator := testdata.NewDataGenerator(db, config)
    return generator.GenerateAll()
}