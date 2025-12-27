package repository

import (
	"auto-video-service/config"
	"auto-video-service/entity"
	"context"
	"errors"
	"sync"
	"time"
)

var (
	englishIdiomRepositoryOnce     sync.Once
	englishIdiomRepositoryInstance *englishIdiomRepository
)

func EnglishIdiomRepository() *englishIdiomRepository {
	englishIdiomRepositoryOnce.Do(func() {
		englishIdiomRepositoryInstance = &englishIdiomRepository{}
	})

	return englishIdiomRepositoryInstance
}

type englishIdiomRepository struct {
}

func (englishIdiomRepository) FindById(ctx context.Context, id int64) (entity.EnglishIdiom, error) {
	db := config.GetDatabase()
	var englishIdiom entity.EnglishIdiom
	q := db.Table("english_idioms").Where("id=?", id)

	has, err := q.Get(&englishIdiom)
	if err != nil {
		return englishIdiom, err
	}

	if has == false {
		err = errors.New("영어숙어를 찾을 수 없습니다")
		return englishIdiom, err
	}

	return englishIdiom, nil
}

func (englishIdiomRepository) FindByToday(ctx context.Context) ([]entity.EnglishIdiom, error) {
	db := config.GetDatabase()
	var englishIdioms []entity.EnglishIdiom
	today := time.Now().Format("20060102")

	err := db.Table("english_idioms").Where("created_date = ?", today).Find(&englishIdioms)
	if err != nil {
		return nil, err
	}

	return englishIdioms, nil
}

func (englishIdiomRepository) FindByDate(ctx context.Context, dateStr string) ([]entity.EnglishIdiom, error) {
	db := config.GetDatabase()
	var englishIdioms []entity.EnglishIdiom

	err := db.Table("english_idioms").Where("created_date = ?", dateStr).Find(&englishIdioms)
	if err != nil {
		return nil, err
	}

	return englishIdioms, nil
}
