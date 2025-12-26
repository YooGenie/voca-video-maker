package factory

import (
	"auto-video-service/enum"
	"auto-video-service/service"
	"context"
	"log"
	"time"
)

type VideoServiceFactory struct{}

func NewVideoServiceFactory() *VideoServiceFactory {
	return &VideoServiceFactory{}
}

func (f *VideoServiceFactory) CreateVideo(ctx context.Context, dateFlag string, serviceType string) {
	targetDate := f.getTargetDate(dateFlag)

	switch enum.ServiceType(serviceType) {

	case enum.InstagramWord, enum.InstagramIdiom, enum.InstagramSentence:
		instagramService := service.NewInstagramService()
		instagramService.CreateReels(ctx, targetDate, serviceType)

	case enum.FacebookWord, enum.FacebookIdiom, enum.FacebookSentence:
		facebookService := service.NewFacebookService()
		facebookService.CreateReels(ctx, targetDate, serviceType)

	case enum.YoutubeShortsWord, enum.YoutubeShortsIdiom, enum.YoutubeShotsSentence:
		youtubeShortsService := service.NewYoutubeShortsService()
		youtubeShortsService.CreateReels(ctx, targetDate, serviceType)

	case enum.YoutubeLongform:
		longformService := service.NewLongformWordService()
		longformService.CreateLongformWords(ctx, targetDate, serviceType)

	case enum.Start:
		startService := service.NewStartService()
		startService.CreateStartCommentVideo(ctx, targetDate, serviceType)

	default:
		log.Fatalf("잘못된 서비스 타입입니다: %s", serviceType)
	}
}

func (f *VideoServiceFactory) getTargetDate(dateFlag string) time.Time {
	var targetDate time.Time
	if dateFlag != "" {
		parsedDate, err := time.Parse("20060102", dateFlag)
		if err != nil {
			log.Fatalf("날짜 형식이 잘못되었습니다. YYYYMMDD 형식으로 입력하세요: %v", err)
		}
		targetDate = parsedDate
		log.Printf("지정된 날짜: %s", targetDate.Format("2006-01-02"))
	} else {
		targetDate = time.Now()
		log.Printf("오늘 날짜: %s", targetDate.Format("2006-01-02"))
	}
	return targetDate
}
