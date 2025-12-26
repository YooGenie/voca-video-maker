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
	englishWordRepositoryOnce     sync.Once
	englishWordRepositoryInstance *englishWordRepository
)

func EnglishWordRepository() *englishWordRepository {
	englishWordRepositoryOnce.Do(func() {
		englishWordRepositoryInstance = &englishWordRepository{}
	})

	return englishWordRepositoryInstance
}

type englishWordRepository struct {
}

func (englishWordRepository) FindById(ctx context.Context, id int64) (entity.EnglishWord, error) {
	db := config.ConfigureDatabase()
	var englishWord entity.EnglishWord
	q := db.Table("english_words").Where("id=?", id)

	has, err := q.Get(&englishWord)
	if err != nil {
		return englishWord, err
	}

	if has == false {
		err = errors.New("영어단어를 찾을 수 없습니다")
		return englishWord, err
	}

	return englishWord, nil
}

func (englishWordRepository) FindByToday(ctx context.Context) ([]entity.EnglishWord, error) {
	db := config.ConfigureDatabase()
	var englishWords []entity.EnglishWord
	today := time.Now().Format("20060102")

	err := db.Table("english_words").Where("created_date = ?", today).Find(&englishWords)
	if err != nil {
		return nil, err
	}

	return englishWords, nil
}

func (englishWordRepository) FindByDate(ctx context.Context, dateStr string) ([]entity.EnglishWord, error) {
	db := config.ConfigureDatabase()
	var englishWords []entity.EnglishWord

	err := db.Table("english_words").Where("created_date = ?", dateStr).Find(&englishWords)
	if err != nil {
		return nil, err
	}

	return englishWords, nil
}
