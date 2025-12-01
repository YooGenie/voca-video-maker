package video_type

import (
	"context"
	"fmt"
	"log"
	"time"

	"auto-video-service/dto"
	"auto-video-service/repository"
	"auto-video-service/service"
)

type EnglishIdiomService struct{}

func NewEnglishIdiomService() *EnglishIdiomService {
	return &EnglishIdiomService{}
}

func (s *EnglishIdiomService) CreateIdiomsReels(ctx context.Context, targetDate time.Time, serviceType *string) {
	// 영어 숙어 DB에서 조회
	idiom, meaning, example, err := s.GetIdiomsByDate(ctx, targetDate)
	if err != nil {
		log.Fatalf("영어 숙어 조회 실패: %v", err)
	}

	// DTO 생성
	request := dto.VideoCreationRequest{
		TargetDate:  targetDate,
		ServiceType: *serviceType,
	}

	contentData := dto.ContentData{
		Primary:   idiom,
		Secondary: meaning,
		Tertiary:  example,
		Count:     len(idiom),
	}

	templateConfig := dto.TemplateConfig{
		BaseTemplate:  "template/idiom.png",
		CountTemplate: "template/idiomCount",
	}

	// 릴스 제작 서비스 사용
	reelsService := service.NewReelsCreationService()
	response := reelsService.CreateCompleteReels(ctx, request, contentData, templateConfig)

	if !response.Success {
		log.Fatalf("비디오 생성 실패: %v", response.Error)
	}

	fmt.Println("\n📚 생성된 영어 숙어 목록:")
	fmt.Println("=" + fmt.Sprintf("%*s", 40, "") + "=")
	for i := 0; i < len(idiom); i++ {
		fmt.Printf("%d) %s (%s)\n", i+1, idiom[i], meaning[i])
	}
	fmt.Println("=" + fmt.Sprintf("%*s", 40, "") + "=")
}

// GetIdiomsByDate - 지정된 날짜의 영어숙어를 조회하여 3개의 배열로 반환
func (s *EnglishIdiomService) GetIdiomsByDate(ctx context.Context, targetDate time.Time) ([]string, []string, []string, error) {
	// 영어숙어 Repository 생성
	idiomRepo := repository.EnglishIdiomRepository()
	
	// 날짜를 YYYYMMDD 형식으로 변환
	dateStr := targetDate.Format("20060102")
	
	// 데이터베이스에서 지정된 날짜의 영어숙어 조회
	idioms, err := idiomRepo.FindByDate(ctx, dateStr)
	if err != nil {
		log.Printf("데이터베이스 조회 실패: %v", err)
		return nil, nil, nil, err
	}
	
	// 조회된 데이터가 없으면 에러 처리
	if len(idioms) == 0 {
		return nil, nil, nil, fmt.Errorf("%s에 생성된 영어숙어가 없습니다", dateStr)
	}
	
	// 3개의 배열로 데이터 분리
	idiom := make([]string, 0, len(idioms))
	meaning := make([]string, 0, len(idioms))
	example := make([]string, 0, len(idioms))
	
	for _, i := range idioms {
		idiom = append(idiom, i.Idiom)
		meaning = append(meaning, i.Meaning)
		example = append(example, i.PronunciationKr)
	}
	
	log.Printf("데이터베이스에서 %s 날짜의 %d개 영어숙어를 조회했습니다.", dateStr, len(idioms))
	
	return idiom, meaning, example, nil
}