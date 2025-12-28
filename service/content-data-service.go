package service

import (
	"auto-video-service/dto"
	"auto-video-service/entity"
	"auto-video-service/enum"
	"auto-video-service/repository"
	"context"
	"fmt"
	"log"
	"time"
)

type ContentDataService struct{}

func NewContentDataService() *ContentDataService {
	return &ContentDataService{}
}

func (s *ContentDataService) GetShortsContentByContentType(ctx context.Context, targetDate time.Time, contentType enum.ContentType) (*dto.ContentDataResult, error) {
	dateStr := targetDate.Format("20060102")

	switch contentType {
	case enum.ContentWord:
		return s.getWordByDate(ctx, dateStr)
	case enum.ContentIdiom:
		return s.getIdiomByDate(ctx, dateStr)
	case enum.ContentSentence:
		return s.getSentenceByDate(ctx, dateStr)
	default:
		return nil, fmt.Errorf("알 수 없는 콘텐츠 타입: %s", contentType)
	}
}

func (s *ContentDataService) getWordByDate(ctx context.Context, dateStr string) (*dto.ContentDataResult, error) {
	repo := repository.EnglishWordRepository()
	words, err := repo.FindByDate(ctx, dateStr)
	if err != nil {
		return nil, fmt.Errorf("단어 조회 실패: %w", err)
	}
	if len(words) == 0 {
		return nil, fmt.Errorf("%s에 생성된 영어단어가 없습니다", dateStr)
	}

	result := &dto.ContentDataResult{
		Primary:   make([]string, 0, len(words)),
		Secondary: make([]string, 0, len(words)),
		Tertiary:  make([]string, 0, len(words)),
	}

	for _, word := range words {
		result.Primary = append(result.Primary, word.EnglishWord)
		result.Secondary = append(result.Secondary, word.Meaning)
		result.Tertiary = append(result.Tertiary, word.PronunciationKr)
	}

	log.Printf("숏폼 DB에서 %s 날짜의 %d개 단어를 조회했습니다.", dateStr, len(words))
	return result, nil
}

func (s *ContentDataService) getIdiomByDate(ctx context.Context, dateStr string) (*dto.ContentDataResult, error) {
	repo := repository.EnglishIdiomRepository()
	idioms, err := repo.FindByDate(ctx, dateStr)
	if err != nil {
		return nil, fmt.Errorf("숙어 조회 실패: %w", err)
	}
	if len(idioms) == 0 {
		return nil, fmt.Errorf("%s에 생성된 영어숙어가 없습니다", dateStr)
	}

	result := &dto.ContentDataResult{
		Primary:   make([]string, 0, len(idioms)),
		Secondary: make([]string, 0, len(idioms)),
		Tertiary:  make([]string, 0, len(idioms)),
	}

	for _, idiom := range idioms {
		result.Primary = append(result.Primary, idiom.Idiom)
		result.Secondary = append(result.Secondary, idiom.Meaning)
		result.Tertiary = append(result.Tertiary, idiom.PronunciationKr)
	}

	log.Printf("숏폼 DB에서 %s 날짜의 %d개 숙어를 조회했습니다.", dateStr, len(idioms))
	return result, nil
}

func (s *ContentDataService) getSentenceByDate(ctx context.Context, dateStr string) (*dto.ContentDataResult, error) {
	repo := repository.ShortSentenceRepository()
	sentences, err := repo.FindByDate(ctx, dateStr)
	if err != nil {
		return nil, fmt.Errorf("문장 조회 실패: %w", err)
	}
	if len(sentences) == 0 {
		return nil, fmt.Errorf("%s에 생성된 단문이 없습니다", dateStr)
	}

	result := &dto.ContentDataResult{
		Primary:        make([]string, 0, len(sentences)),
		PrimaryLine2:   make([]string, 0, len(sentences)),
		Secondary:      make([]string, 0, len(sentences)),
		SecondaryLine2: make([]string, 0, len(sentences)),
		Tertiary:       make([]string, 0, len(sentences)),
	}

	for _, s := range sentences {
		result.Primary = append(result.Primary, s.EnglishSentence1)
		result.Secondary = append(result.Secondary, s.KoreanSentence1)
		result.Tertiary = append(result.Tertiary, s.Pronunciation)

		if s.EnglishSentence2.Valid {
			result.PrimaryLine2 = append(result.PrimaryLine2, s.EnglishSentence2.String)
		} else {
			result.PrimaryLine2 = append(result.PrimaryLine2, "")
		}

		if s.KoreanSentence2.Valid {
			result.SecondaryLine2 = append(result.SecondaryLine2, s.KoreanSentence2.String)
		} else {
			result.SecondaryLine2 = append(result.SecondaryLine2, "")
		}
	}

	log.Printf("숏폼 DB에서 %s 날짜의 %d개 문장을 조회했습니다.", dateStr, len(sentences))
	return result, nil
}

func (s *ContentDataService) GetYoutubeShortsContentByDate(ctx context.Context, targetDate time.Time, contentType enum.ContentType) (*dto.ContentDataResult, error) {
	dateStr := targetDate.Format("20060102")
	repo := repository.LongformWordRepository()

	longformWords, err := repo.FindByShortsDateAndContentType(ctx, dateStr, string(contentType))
	if err != nil {
		return nil, fmt.Errorf("유튜브 숏폼 DB 조회 실패: %w", err)
	}
	if len(longformWords) == 0 {
		return nil, fmt.Errorf("%s에 해당하는 유튜브 숏폼용 %s 데이터가 없습니다", dateStr, contentType)
	}

	result := &dto.ContentDataResult{
		Primary:   make([]string, 0, len(longformWords)),
		Secondary: make([]string, 0, len(longformWords)),
		Tertiary:  make([]string, 0, len(longformWords)),
	}

	for _, word := range longformWords {
		result.Primary = append(result.Primary, word.Word)
		result.Secondary = append(result.Secondary, word.Meaning)
		result.Tertiary = append(result.Tertiary, word.PronunciationKr)
	}

	log.Printf("유튜브 숏폼 DB에서 %s 날짜의 %d개 %s를 조회했습니다.", dateStr, len(longformWords), contentType)
	return result, nil
}

var _ = entity.LongformWord{}
