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
	engSentences, engSentences2, korSentences, korSentences2, pronunciations, err := s.GetShortSentencesByDate(ctx, targetDate)
	if err != nil {
		log.Fatalf("단문 조회 실패: %v", err)
	}

	// 2. DTO 생성
	request := dto.VideoCreationRequest{
		TargetDate:  targetDate,
		ServiceType: *serviceType,
	}

	contentData := dto.ContentData{
		Primary:        engSentences,
		PrimaryLine2:   engSentences2, // 영어 두 번째 줄
		Secondary:      korSentences,
		SecondaryLine2: korSentences2, // 한국어 두 번째 줄
		Tertiary:       pronunciations,
		Count:          len(engSentences),
	}

	// 'SS' 타입에 맞는 템플릿 설정
	templateConfig := dto.TemplateConfig{
		BaseTemplate:  "template/short_sentence.png",
		CountTemplate: "template/shortSentenceCount.png",
		TextColor:     "black", // SS 타입은 검정색 글씨
	}

	// 3. 릴스 제작 서비스 호출
	reelsService := service.NewReelsCreationService()
	response := reelsService.CreateCompleteReels(ctx, request, contentData, templateConfig)

	if !response.Success {
		log.Fatalf("비디오 생성 실패: %v", response.Error)
	}

	// 4. 생성된 문장 목록 출력
	fmt.Println("\n📚 ")
	fmt.Println("=" + fmt.Sprintf("%*s", 40, "") + "=")
	for i := 0; i < len(engSentences); i++ {
		fmt.Printf("%d) %s (%s)\n", i+1, engSentences[i], korSentences[i])
	}
	fmt.Println("=" + fmt.Sprintf("%*s", 40, "") + "=")
}

// GetShortSentencesByDate - DB에서 데이터를 가져와 릴스 생성 서비스가 이해할 수 있는 형식으로 가공
func (s *ShortSentenceService) GetShortSentencesByDate(ctx context.Context, targetDate time.Time) (engs []string, engs2 []string, kors []string, kors2 []string, pros []string, err error) {
	repo := repository.ShortSentenceRepository()
	dateStr := targetDate.Format("20060102")

	dbData, err := repo.FindByDate(ctx, dateStr)
	if err != nil {
		return nil, nil, nil, nil, nil, fmt.Errorf("데이터베이스 조회 실패: %w", err)
	}

	if len(dbData) == 0 {
		return nil, nil, nil, nil, nil, fmt.Errorf("%s에 생성된 단문이 없습니다", dateStr)
	}

	// DB 데이터를 릴스 생성 서비스에 맞는 형식으로 변환
	// EnglishSentence1/2와 KoreanSentence1/2를 별도 배열로 관리
	for _, data := range dbData {
		engs = append(engs, data.EnglishSentence1)
		kors = append(kors, data.KoreanSentence1)
		pros = append(pros, data.Pronunciation)

		// EnglishSentence2가 있으면 추가, 없으면 빈 문자열
		if data.EnglishSentence2.Valid {
			engs2 = append(engs2, data.EnglishSentence2.String)
		} else {
			engs2 = append(engs2, "")
		}

		// KoreanSentence2가 있으면 추가, 없으면 빈 문자열
		if data.KoreanSentence2.Valid {
			kors2 = append(kors2, data.KoreanSentence2.String)
		} else {
			kors2 = append(kors2, "")
		}
	}

	log.Printf("데이터베이스에서 %s 날짜의 %d개 행을 조회하여 %d개의 클립 데이터를 생성했습니다.", dateStr, len(dbData), len(engs))
	return engs, engs2, kors, kors2, pros, nil
}
