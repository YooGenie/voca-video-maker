package repository

import (
	"auto-video-service/config"
	"auto-video-service/entity"
	"context"
	"sync"
)

var (
	shortSentenceRepositoryOnce     sync.Once
	shortSentenceRepositoryInstance *shortSentenceRepository
)

// ShortSentenceRepository returns the singleton instance of the short sentence repository.
func ShortSentenceRepository() *shortSentenceRepository {
	shortSentenceRepositoryOnce.Do(func() {
		shortSentenceRepositoryInstance = &shortSentenceRepository{}
	})

	return shortSentenceRepositoryInstance
}

type shortSentenceRepository struct{}

// FindByDate retrieves short sentences for a specific date from the database.
func (r *shortSentenceRepository) FindByDate(ctx context.Context, dateStr string) ([]entity.ShortSentence, error) {
	db := config.ConfigureDatabase()
	var sentences []entity.ShortSentence

	err := db.Table("short_sentences").Where("created_date = ?", dateStr).Find(&sentences)
	if err != nil {
		return nil, err
	}

	return sentences, nil
}
