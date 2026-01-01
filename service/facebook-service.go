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

type FacebookService struct{}

func NewFacebookService() *FacebookService {
	return &FacebookService{}
}

func (s *FacebookService) CreateReels(ctx context.Context, targetDate time.Time, serviceType string) {
	contentType := s.getContentType(serviceType)

	contentDataService := NewContentDataService()
	contentResult, err := contentDataService.GetShortsContentByContentType(ctx, targetDate, contentType)
	if err != nil {
		log.Fatalf("콘텐츠 조회 실패: %v", err)
	}

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
		IsReverse:      false,
	}

	templateConfig := s.getTemplateConfig(contentType)

	options := dto.VideoCreationOptions{
		Platform:           enum.PlatformFacebook,
		VideoLength:        enum.VideoLengthShort,
		EnglishRepeatCount: 3,
		SpeakSpeed:         0.8,
		PauseDuration:      0.5,
		TemplateType:       enum.TemplateIndividual,
	}

	reelsService := NewReelsCreationService()
	response := reelsService.CreateCompleteReels(ctx, request, contentData, templateConfig, options)

	if !response.Success {
		log.Fatalf("비디오 생성 실패: %v", response.Error)
	}

	s.printResult(contentType, contentResult)
}

func (s *FacebookService) getContentType(serviceType string) enum.ContentType {
	switch enum.ServiceType(serviceType) {
	case enum.FacebookWord:
		return enum.ContentWord
	case enum.FacebookIdiom:
		return enum.ContentIdiom
	case enum.FacebookSentence:
		return enum.ContentSentence
	default:
		return enum.ContentWord
	}
}

// getTemplateConfig - 콘텐츠 타입별 템플릿 설정
// 2026년부터 모든 세로형 비디오는 Vertical 템플릿 하나로 통일
func (s *FacebookService) getTemplateConfig(contentType enum.ContentType) dto.TemplateConfig {
	paths := config.Config.Paths.Templates
	// 세로형 비디오는 모두 동일한 템플릿 사용
	return dto.TemplateConfig{
		BaseTemplate: paths.Vertical,
		TextColor:    enum.TextColorBeige,
	}
}

// printResult - 생성 결과 출력
func (s *FacebookService) printResult(contentType enum.ContentType, result *dto.ContentDataResult) {
	fmt.Printf("\n📘 페이스북 %s 영상 생성 완료!\n", contentType)
	fmt.Println("=" + fmt.Sprintf("%*s", 40, "") + "=")
	for i := 0; i < len(result.Primary); i++ {
		fmt.Printf("%d) %s (%s)\n", i+1, result.Primary[i], result.Secondary[i])
	}
	fmt.Println("=" + fmt.Sprintf("%*s", 40, "") + "=")
}
