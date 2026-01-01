package service

import (
	"auto-video-service/config"
	"auto-video-service/dto"
	"auto-video-service/enum"
	"context"
	"fmt"
	"log"
	"time"
)

type YoutubeShortsService struct{}

func NewYoutubeShortsService() *YoutubeShortsService {
	return &YoutubeShortsService{}
}

// CreateReels - 유튜브 숏폼 영상 생성
// serviceType: ysw (단어), ysi (숙어), yss (문장)
func (s *YoutubeShortsService) CreateReels(ctx context.Context, targetDate time.Time, serviceType string) {
	contentType := s.getContentType(serviceType)

	contentDataService := NewContentDataService()
	contentResult, err := contentDataService.GetYoutubeShortsContentByDate(ctx, targetDate, contentType)
	if err != nil {
		log.Fatalf("콘텐츠 조회 실패: %v", err)
	}

	// DTO 생성
	request := dto.VideoCreationRequest{
		TargetDate:  targetDate,
		ServiceType: serviceType,
		ContentType: contentType,
	}

	contentData := dto.ContentData{
		Primary:        contentResult.Primary,
		PrimaryLine2:   contentResult.PrimaryLine2,
		Secondary:      contentResult.Secondary,
		SecondaryLine2: contentResult.SecondaryLine2,
		Tertiary:       contentResult.Tertiary,
		Count:          len(contentResult.Primary),
		IsReverse:      false, // 기본값
	}

	templateConfig := s.getTemplateConfig(contentType)

	// 유튜브 숏폼 기본 옵션
	options := dto.VideoCreationOptions{
		Platform:           enum.PlatformYoutube,
		VideoLength:        enum.VideoLengthShort,
		EnglishRepeatCount: 2,
		SpeakSpeed:         1.0,
		PauseDuration:      0,
		TemplateType:       enum.TemplateIndividual,
	}

	// 릴스 생성
	reelsService := NewReelsCreationService()
	response := reelsService.CreateCompleteReels(ctx, request, contentData, templateConfig, options)

	if !response.Success {
		log.Fatalf("비디오 생성 실패: %v", response.Error)
	}

	// 생성 결과 출력
	s.printResult(contentType, contentResult)
}

// getContentType - serviceType에서 콘텐츠 타입 추출
func (s *YoutubeShortsService) getContentType(serviceType string) enum.ContentType {
	switch enum.ServiceType(serviceType) {
	case enum.YoutubeShortsWord:
		return enum.ContentWord
	case enum.YoutubeShortsIdiom:
		return enum.ContentIdiom
	case enum.YoutubeShotsSentence:
		return enum.ContentSentence
	default:
		return enum.ContentWord
	}
}

// getTemplateConfig - 콘텐츠 타입별 템플릿 설정
// 2026년부터 모든 세로형 비디오는 Vertical 템플릿 하나로 통일
func (s *YoutubeShortsService) getTemplateConfig(contentType enum.ContentType) dto.TemplateConfig {
	paths := config.Config.Paths.Templates
	// 세로형 비디오는 모두 동일한 템플릿 사용
	return dto.TemplateConfig{
		BaseTemplate: paths.Vertical,
		TextColor:    enum.TextColorBeige,
	}
}

// printResult - 생성 결과 출력
func (s *YoutubeShortsService) printResult(contentType enum.ContentType, result *dto.ContentDataResult) {
	fmt.Printf("\n▶️ 유튜브 숏폼 %s 영상 생성 완료!\n", contentType)
	fmt.Println("=" + fmt.Sprintf("%*s", 40, "") + "=")
	for i := 0; i < len(result.Primary); i++ {
		fmt.Printf("%d) %s (%s)\n", i+1, result.Primary[i], result.Secondary[i])
	}
	fmt.Println("=" + fmt.Sprintf("%*s", 40, "") + "=")
}
