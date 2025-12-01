package repository

import (
	"auto-video-service/config"
	"auto-video-service/entity"
	"context"
	"sync"
)

var (
	longformWordRepositoryOnce     sync.Once
	longformWordRepositoryInstance *longformWordRepository
)

func LongformWordRepository() *longformWordRepository {
	longformWordRepositoryOnce.Do(func() {
		longformWordRepositoryInstance = &longformWordRepository{}
	})

	return longformWordRepositoryInstance
}

type longformWordRepository struct{}

func (r *longformWordRepository) FindByDate(ctx context.Context, dateStr string) ([]entity.LongformWord, error) {
	db := config.ConfigureDatabase()
	var longformWords []entity.LongformWord

	err := db.Table("longform_words").Where("created_date = ?", dateStr).Find(&longformWords)
	if err != nil {
		return nil, err
	}

	return longformWords, nil
}