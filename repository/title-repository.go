package repository

import (
	"auto-video-service/config"
	"auto-video-service/entity"
	"context"
	"errors"
	"sync"
)

var (
	titleRepositoryOnce     sync.Once
	titleRepositoryInstance *titleRepository
)

func TitleRepository() *titleRepository {
	titleRepositoryOnce.Do(func() {
		titleRepositoryInstance = &titleRepository{}
	})

	return titleRepositoryInstance
}

type titleRepository struct{}

func (r *titleRepository) FindByDate(ctx context.Context, dateStr string) (*entity.Title, error) {
	db := config.GetDatabase()
	var title entity.Title
	has, err := db.Table("title").Where("created_date = ?", dateStr).Get(&title)
	if err != nil {
		return nil, err
	}
	if !has {
		return nil, errors.New("해당 날짜의 타이틀을 찾을 수 없습니다")
	}
	return &title, nil
}
