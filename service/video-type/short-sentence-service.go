package video_type

import (
	"auto-video-service/dto"
	"auto-video-service/repository"
	"auto-video-service/service"
	"context"
	"fmt"
	"log"
	"time"
)

type ShortSentenceService struct{}

func NewShortSentenceService() *ShortSentenceService {
	return &ShortSentenceService{}
}

func (s *ShortSentenceService) CreateShortSentenceReels(ctx context.Context, targetDate time.Time, serviceType *string) {
	// 1. DB에서 데이터 조회 및 가공
	engSentences, korSentences, pronunciations, err := s.GetShortSentencesByDate(ctx, targetDate)
	if err != nil {
		log.Fatalf("단문 조회 실패: %v", err)
	}

	// 2. DTO 생성
	request := dto.VideoCreationRequest{
		TargetDate:  targetDate,
		ServiceType: *serviceType,
	}

	contentData := dto.ContentData{
		Primary:   engSentences,
		Secondary: korSentences,
		Tertiary:  pronunciations,
		Count:     len(engSentences),
	}

	// 'SS' 타입에 맞는 템플릿 설정
	templateConfig := dto.TemplateConfig{
		BaseTemplate:  "template/short_sentence.png",
		CountTemplate: "template/shortSentenceCount.png",
	}

	// 3. 릴스 제작 서비스 호출
	reelsService := service.NewReelsCreationService()
	response := reelsService.CreateCompleteReels(ctx, request, contentData, templateConfig)

	if !response.Success {
		log.Fatalf("비디오 생성 실패: %v", response.Error)
	}

	// 4. 생성된 문장 목록 출력
	fmt.Println("\n📚 생성된 영어 단문 목록:")
	fmt.Println("=" + fmt.Sprintf("%*s", 40, "") + "=")
	for i := 0; i < len(engSentences); i++ {
		fmt.Printf("%d) %s\n   - %s\n", i+1, engSentences[i], korSentences[i])
	}
	fmt.Println("=" + fmt.Sprintf("%*s", 40, "") + "=")
}

// GetShortSentencesByDate - DB에서 데이터를 가져와 릴스 생성 서비스가 이해할 수 있는 형식으로 가공
func (s *ShortSentenceService) GetShortSentencesByDate(ctx context.Context, targetDate time.Time) (engs []string, kors []string, pros []string, err error) {
	repo := repository.ShortSentenceRepository()
	dateStr := targetDate.Format("20060102")

	dbData, err := repo.FindByDate(ctx, dateStr)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("데이터베이스 조회 실패: %w", err)
	}

	if len(dbData) == 0 {
		return nil, nil, nil, fmt.Errorf("%s에 생성된 단문이 없습니다", dateStr)
	}

	// DB 데이터를 릴스 생성 서비스에 맞는 형식으로 변환 (Flatten)
	for _, data := range dbData {
		// 첫 번째 문장 쌍 추가
		engs = append(engs, data.EnglishSentence1)
		kors = append(kors, data.KoreanSentence1)
		pros = append(pros, data.Pronunciation)

		// 두 번째 문장 쌍이 존재하면 추가
		if data.EnglishSentence2.Valid && data.KoreanSentence2.Valid {
			engs = append(engs, data.EnglishSentence2.String)
			kors = append(kors, data.KoreanSentence2.String)
			pros = append(pros, data.Pronunciation) // 동일한 발음 정보 사용
		}
	}

	log.Printf("데이터베이스에서 %s 날짜의 %d개 행을 조회하여 %d개의 클립 데이터를 생성했습니다.", dateStr, len(dbData), len(engs))
	return engs, kors, pros, nil
}
