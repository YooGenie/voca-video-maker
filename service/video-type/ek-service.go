package video_type

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"auto-video-service/dto"
	"auto-video-service/repository"
	"auto-video-service/service"
)

type EKService struct{}

func NewEKService() *EKService {
	return &EKService{}
}

// CreateEKReels - EK 타입 릴스 제작 전체 과정 (영어 -> 한국어 순서)
func (s *EKService) CreateEKReels(ctx context.Context, targetDate time.Time, serviceType *string) dto.VideoCreationResponse {
	response := dto.VideoCreationResponse{
		Success: false,
	}

	// 1. 영어 단어 DB에서 조회
	eng, kor, pronounce, err := s.GetWordsByDate(ctx, targetDate)
	if err != nil {
		log.Printf("영어단어 조회 실패: %v", err)
		response.Error = err
		return response
	}
	contentCount := len(eng)
	response.ContentCount = contentCount

	// 이미지 서비스 생성
	imageService := service.NewImageService()

	// 2. 조회된 컨텐츠 개수만큼 이미지 생성 (영어 -> 한국어 순서)
	// 먼저 컨텐츠 개수를 표시하는 이미지 생성
	templateConfig := dto.TemplateConfig{
		BaseTemplate:  "template/word.png", // W타입과 동일한 템플릿 사용
		CountTemplate: "template/wordCount",
	}

	err = imageService.SetWordCountOnImage(
		templateConfig.BaseTemplate,
		fmt.Sprintf("%d", contentCount),
		templateConfig.CountTemplate,
		*serviceType, // 서비스 타입
	)
	if err != nil {
		log.Printf("contentCount 이미지 생성 실패: %v", err)
		response.Error = err
		return response
	} else {
		log.Println("contentCount 이미지 생성 완료!")
	}

	// 그 다음 기본 이미지들 생성
	newTemplateImagePath := templateConfig.CountTemplate + ".png"
	err = imageService.GenerateEKImages(
		newTemplateImagePath, // 컨텐츠 개수가 표시된 이미지 템플릿
		eng,                  // 영어 단어들
		kor,                  // 한국어 번역들
		pronounce,            // 발음들
		"images/output",      // 출력 파일 접두사 (images 디렉토리에 저장)
		contentCount*2,       // 생성할 이미지 개수 (동적)
	)
	if err != nil {
		log.Printf("이미지 생성 실패: %v", err)
		response.Error = err
		return response
	}
	log.Println("이미지 생성 완료!")

	// 3. 서비스 생성
	reelsConfig := service.VideoConfig{Width: 1080, Height: 1920} // 세로형
	videoService := service.NewVideoService(imageService, reelsConfig)
	audioService := service.NewAudioService()

	// 4. 각 컨텐츠에 대한 음성 파일 생성
	audioDir := "audio"
	if err := os.MkdirAll(audioDir, 0755); err != nil {
		log.Printf("audio 디렉토리 생성 실패: %v", err)
		response.Error = err
		return response
	}

	// 영어 컨텐츠 원어민 음성 생성
	log.Println("🎤 영어 컨텐츠 원어민 음성을 생성합니다...")
	for i, content := range eng {
		audioPath := fmt.Sprintf("%s/eng_%d.mp3", audioDir, i)
		if err := audioService.CreateNativeEnglishAudio(content, audioPath); err != nil {
			log.Printf("영어 원어민 음성 생성 실패 (%s): %v", content, err)
		}
	}

	// 한국어 컨텐츠 음성 생성
	log.Println("🎤 한국어 컨텐츠 음성을 생성합니다...")
	for i, content := range kor {
		audioPath := fmt.Sprintf("%s/kor_%d.mp3", audioDir, i)
		if err := audioService.CreateKoreanAudioWithRate(content, audioPath, 175); err != nil {
			log.Printf("한국어 음성 생성 실패 (%s): %v", content, err)
		}
	}
	log.Println("음성 파일 생성 완료!")

	// videos 디렉토리 생성
	if err := os.MkdirAll("videos", 0755); err != nil {
		log.Printf("videos 디렉토리 생성 실패: %v", err)
		response.Error = err
		return response
	}

	// 5. 각 이미지에 음성을 추가한 영상 생성 (영어 -> 한국어 순서)
	videoPaths := make([]string, 0, contentCount*2)
	for i := 0; i < contentCount*2; i++ {
		var outputPath string
		isFirstClipOfPair := i%2 == 0

		if isFirstClipOfPair { // 첫 번째 클립: 영어
			imagePath := fmt.Sprintf("images/output_%02d.png", i+1)
			englishAudioPath := fmt.Sprintf("audio/eng_%d.mp3", i/2)
			outputPath = fmt.Sprintf("videos/video_%d.mp4", i)

			if err := videoService.CreateVideoWithEnglish(imagePath, englishAudioPath, outputPath, 0.5); err != nil {
				log.Printf("영어 영상 생성 실패 (%d): %v", i, err)
				response.Error = err
				return response
			}
		} else { // 두 번째 클립: 한국어
			imagePath := fmt.Sprintf("images/output_%02d.png", i+1)
			koreanAudioPath := fmt.Sprintf("audio/kor_%d.mp3", i/2)
			outputPath = fmt.Sprintf("videos/video_%d.mp4", i)

			if err := videoService.CreateVideoWithKorean(imagePath, koreanAudioPath, outputPath, 0.5); err != nil {
				log.Printf("한국어 영상 생성 실패 (%d): %v", i, err)
				response.Error = err
				return response
			}
		}
		videoPaths = append(videoPaths, outputPath)
		log.Printf("영상 생성 완료: %d/%d", i+1, contentCount*2)
	}
	log.Println("개별 영상 생성 완료!")

	// 6. 모든 영상을 하나로 합치기
	finalFileName := fmt.Sprintf("%02d%02d%02d_ek.mp4", targetDate.Year()%100, targetDate.Month(), targetDate.Day())
	response.FinalFileName = finalFileName

	err = videoService.ConcatenateVideos(videoPaths, finalFileName)
	if err != nil {
		log.Printf("영상 합치기 실패: %v", err)
		response.Error = err
		return response
	}
	log.Println("최종 영상 생성 완료!")

	// 7. 중간 파일들 정리
	s.cleanupTempFiles()

	log.Println("중간 파일들 정리 완료!")
	log.Printf("최종 영상: %s", finalFileName)

	response.Success = true
	return response
}

// GetWordsByDate - W타입 서비스의 로직과 동일
func (s *EKService) GetWordsByDate(ctx context.Context, targetDate time.Time) ([]string, []string, []string, error) {
	englishWordRepo := repository.EnglishWordRepository()
	dateStr := targetDate.Format("20060102")
	englishWords, err := englishWordRepo.FindByDate(ctx, dateStr)
	if err != nil {
		log.Printf("데이터베이스 조회 실패: %v", err)
		return nil, nil, nil, err
	}

	if len(englishWords) == 0 {
		return nil, nil, nil, fmt.Errorf("%s에 생성된 영어단어가 없습니다", dateStr)
	}

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

// cleanupTempFiles - 중간 파일들 정리
func (s *EKService) cleanupTempFiles() {
	log.Println("중간 파일들 정리 중...")

	// images 디렉토리 안의 파일들만 삭제
	if files, err := os.ReadDir("images"); err == nil {
		for _, file := range files {
			if !file.IsDir() {
				os.Remove(filepath.Join("images", file.Name()))
			}
		}
	}

	// audio 디렉토리 안의 파일들만 삭제
	if files, err := os.ReadDir("audio"); err == nil {
		for _, file := range files {
			if !file.IsDir() {
				os.Remove(filepath.Join("audio", file.Name()))
			}
		}
	}

	// videos 디렉토리 안의 파일들만 삭제
	if files, err := os.ReadDir("videos"); err == nil {
		for _, file := range files {
			if !file.IsDir() {
				os.Remove(filepath.Join("videos", file.Name()))
			}
		}
	}
}
