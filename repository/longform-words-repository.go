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

// FindByShortsDate - 유튜브 숏폼 날짜(shorts_date)로 데이터 조회
func (r *longformWordRepository) FindByShortsDate(ctx context.Context, dateStr string) ([]entity.LongformWord, error) {
	db := config.ConfigureDatabase()
	var longformWords []entity.LongformWord

	// shorts_date 컬럼을 기준으로 조회
	err := db.Table("longform_words").Where("shorts_date = ?", dateStr).Find(&longformWords)
	if err != nil {
		return nil, err
	}

	return longformWords, nil
}

// FindByShortsDateAndContentType - 유튜브 숏폼 날짜 + 콘텐츠 타입으로 데이터 조회
func (r *longformWordRepository) FindByShortsDateAndContentType(ctx context.Context, dateStr string, contentType string) ([]entity.LongformWord, error) {
	db := config.ConfigureDatabase()
	var longformWords []entity.LongformWord

	err := db.Table("longform_words").
		Where("shorts_date = ?", dateStr).
		And("content_type = ?", contentType).
		Find(&longformWords)
	if err != nil {
		return nil, err
	}

	return longformWords, nil
}
