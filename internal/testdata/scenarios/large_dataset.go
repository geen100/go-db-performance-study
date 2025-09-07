// internal/testdata/scenarios/large_dataset.go
package scenarios

import (
	"go-db-performance-study/internal/testdata"

	"gorm.io/gorm"
)

// GenerateLargeDataset 大規模データセット生成
func GenerateLargeDataset(db *gorm.DB) error {
	config := testdata.GeneratorConfig{
		UserCount:    100000,
		PostCount:    500000,
		TagCount:     2000,
		CommentCount: 5000000,
		BatchSize:    5000,
	}

	generator := testdata.NewDataGenerator(db, config)
	return generator.GenerateAll()
}
