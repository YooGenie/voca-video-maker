package service

import (
	"context"
	"log"
	"time"
)

type VideoServiceFactory struct{}

func NewVideoServiceFactory() *VideoServiceFactory {
	return &VideoServiceFactory{}
}

func (f *VideoServiceFactory) CreateVideo(ctx context.Context, dateFlag *string, serviceType *string) {
	targetDate := f.getTargetDate(dateFlag)
	var err error

	switch *serviceType {
	case "W":
		englishWordService := NewEnglishWordService()
		englishWordService.CreateWordsReels(ctx, targetDate, serviceType)
	case "I":
		englishIdiomService := NewEnglishIdiomService()
		englishIdiomService.CreateIdiomsReels(ctx, targetDate, serviceType)
	case "L":
		longformWordService := NewLongformWordService()
		longformWordService.CreateLongformWords(ctx, targetDate, serviceType)
	default:
		log.Fatalf("잘못된 서비스 타입입니다. W, I, L 중 하나를 선택하세요.")
	}

	if err != nil {
		log.Fatalf("데이터 조회 실패: %v", err)
	}
}

func (f *VideoServiceFactory) getTargetDate(dateFlag *string) time.Time {
	var targetDate time.Time
	if *dateFlag != "" {
		parsedDate, err := time.Parse("20060102", *dateFlag)
		if err != nil {
			log.Fatalf("날짜 형식이 잘못되었습니다. YYYYMMDD 형식으로 입력하세요 (예: 20250907): %v", err)
		}
		targetDate = parsedDate
		log.Printf("지정된 날짜: %s", targetDate.Format("2006-01-02"))
	} else {
		targetDate = time.Now()
		log.Printf("오늘 날짜: %s", targetDate.Format("2006-01-02"))
	}
	return targetDate
}