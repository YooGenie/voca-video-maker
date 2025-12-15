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

type EnglishWordService struct{}

func NewEnglishWordService() *EnglishWordService {
	return &EnglishWordService{}
}

func (s *EnglishWordService) CreateWordsReels(ctx context.Context, targetDate time.Time, serviceType *string) {
	// 영어 단어 DB에서 조회
	eng, kor, pronounce, err := s.GetWordsByDate(ctx, targetDate)
	if err != nil {
		log.Fatalf("영어단어 조회 실패: %v", err)
	}

	// DTO 생성
	request := dto.VideoCreationRequest{
		TargetDate:  targetDate,
		ServiceType: *serviceType,
	}

	contentData := dto.ContentData{
		Primary:        eng,
		PrimaryLine2:   []string{}, // W 타입은 영어 2줄 표시 안함
		Secondary:      kor,
		SecondaryLine2: []string{}, // W 타입은 한국어 2줄 표시 안함
		Tertiary:       pronounce,
		Count:          len(eng),
	}

	templateConfig := dto.TemplateConfig{
		BaseTemplate:  "template/word.png",
		CountTemplate: "template/wordCount",
		TextColor:     "white", // W 타입은 흰색 글씨
	}

	// 릴스 제작 서비스 사용
	reelsService := service.NewReelsCreationService()
	response := reelsService.CreateCompleteReels(ctx, request, contentData, templateConfig)

	if !response.Success {
		log.Fatalf("비디오 생성 실패: %v", response.Error)
	}

	// 7. 생성된 영어 단어 목록 출력
	fmt.Println("\n📚 생성된 영어 단어 목록:")
	fmt.Println("=" + fmt.Sprintf("%*s", 40, "") + "=")
	for i := 0; i < len(eng); i++ {
		fmt.Printf("%d) %s (%s)\n", i+1, eng[i], kor[i])
	}
	fmt.Println("=" + fmt.Sprintf("%*s", 40, "") + "=")
}

// GetWordsByDate - 지정된 날짜의 영어단어를 조회하여 3개의 배열로 반환
func (s *EnglishWordService) GetWordsByDate(ctx context.Context, targetDate time.Time) ([]string, []string, []string, error) {
	// 영어단어 Repository 생성
	englishWordRepo := repository.EnglishWordRepository()

	// 날짜를 YYYYMMDD 형식으로 변환
	dateStr := targetDate.Format("20060102")

	// 데이터베이스에서 지정된 날짜의 영어단어 조회
	englishWords, err := englishWordRepo.FindByDate(ctx, dateStr)
	if err != nil {
		log.Printf("데이터베이스 조회 실패: %v", err)
		return nil, nil, nil, err
	}

	// 조회된 데이터가 없으면 에러 처리
	if len(englishWords) == 0 {
		return nil, nil, nil, fmt.Errorf("%s에 생성된 영어단어가 없습니다", dateStr)
	}

	// 3개의 배열로 데이터 분리
	eng := make([]string, 0, len(englishWords))
	kor := make([]string, 0, len(englishWords))
	pronounce := make([]string, 0, len(englishWords))

	for _, word := range englishWords {
		eng = append(eng, word.EnglishWord)
		kor = append(kor, word.Meaning)
		pronounce = append(pronounce, word.PronunciationKr)
	}

	log.Printf("데이터베이스에서 %s 날짜의 %d개 영어단어를 조회했습니다.", dateStr, len(englishWords))

	return eng, kor, pronounce, nil
}
