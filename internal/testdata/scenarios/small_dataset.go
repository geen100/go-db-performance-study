// internal/testdata/scenarios/small_dataset.go
package scenarios

import (
    "go-db-performance-study/internal/testdata"
    "gorm.io/gorm"
)

// GenerateSmallDataset 小規模データセット生成
func GenerateSmallDataset(db *gorm.DB) error {
    config := testdata.GeneratorConfig{
        UserCount:    100,
        PostCount:    500,
        TagCount:     30,
        CommentCount: 1000,
        BatchSize:    100,
    }
    
    generator := testdata.NewDataGenerator(db, config)
    return generator.GenerateAll()
}