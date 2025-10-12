package service

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"auto-video-service/repository"
)

type EnglishIdiomService struct{}

func NewEnglishIdiomService() *EnglishIdiomService {
	return &EnglishIdiomService{}
}

func (s *EnglishIdiomService) CreateIdiomsReels(ctx context.Context, targetDate time.Time, serviceType *string){
	// 영어 단어 DB에서 조회
	eng, kor, pronounce, err := s.GetIdiomsByDate(ctx, targetDate)
	if err != nil {
		log.Fatalf("영어 숙어 조회 실패: %v", err)
	}

// 이미지 서비스 생성
imageService := NewImageService()

// 1. 조회된 단어 개수만큼 이미지 생성
wordCount := len(eng)

// 먼저 단어 개수를 표시하는 이미지 생성
templateImagePath := "template/idiom.png"
err = imageService.GenerateOptionalImage(
	templateImagePath,                // img2 이미지 템플릿
	fmt.Sprintf("%d", wordCount),     // wordCount를 문자열로 변환
	"template/idiomCount",        	
	*serviceType,          // 서비스 타입 (W 또는 I)          // 출력 파일명
)
if err != nil {
	log.Printf("wordCount 이미지 생성 실패: %v", err)
} else {
	log.Println("wordCount 이미지 생성 완료!")
}

// 그 다음 기본 이미지들 생성 (img3.png 사용)
newTemplateImagePath := "template/idiomCount.png"
err = imageService.GenerateBasicImages(
	newTemplateImagePath,  // 단어 개수가 표시된 이미지 템플릿
	eng,                   // 영어 단어들
	kor,                   // 한국어 번역들
	pronounce,             // 발음들
	"images/output",       // 출력 파일 접두사 (images 디렉토리에 저장)
	wordCount * 2,         // 생성할 이미지 개수 (동적)
)
if err != nil {
	log.Fatalf("이미지 생성 실패: %v", err)
}
log.Println("이미지 생성 완료!")

	// 2. 비디오 서비스 생성
	reelsConfig := VideoConfig{Width: 1080, Height: 1920}
	videoService := NewVideoService(imageService, reelsConfig)
// 3. 각 단어에 대한 음성 파일 생성
audioDir := "audio"
if err := os.MkdirAll(audioDir, 0755); err != nil {
	log.Fatalf("audio 디렉토리 생성 실패: %v", err)
}

// 영어 단어 원어민 음성 생성
log.Println("🎤 영어 단어 원어민 음성을 생성합니다...")
for i, word := range eng {
	audioPath := fmt.Sprintf("%s/eng_%d.mp3", audioDir, i)
	if err := videoService.GenerateNativeEnglishAudio(word, audioPath); err != nil {
		log.Printf("영어 원어민 음성 생성 실패 (%s): %v", word, err)
	}
}

// 한국어 단어 음성 생성
log.Println("🎤 한국어 단어 음성을 생성합니다...")
for i, word := range kor {
	audioPath := fmt.Sprintf("%s/kor_%d.mp3", audioDir, i)
	if err := videoService.GenerateKoreanAudioFromText(word, audioPath); err != nil {
		log.Printf("한국어 음성 생성 실패 (%s): %v", word, err)
	}
}

log.Println("음성 파일 생성 완료!")

// videos 디렉토리 생성
if err := os.MkdirAll("videos", 0755); err != nil {
	log.Fatalf("videos 디렉토리 생성 실패: %v", err)
}

// 4. 각 이미지에 음성을 추가한 영상 생성
for i := 0; i < wordCount*2; i++ {
	var outputPath string
	
	if i%2 == 0 { // 짝수 - 한국어
		imagePath := fmt.Sprintf("images/output_%02d.png", i+1)
		koreanAudioPath := fmt.Sprintf("audio/kor_%d.mp3", i/2)
		outputPath = fmt.Sprintf("videos/video_%d.mp4", i)
		
		if err := videoService.CreateVideoWithKorean(imagePath, koreanAudioPath, outputPath); err != nil {
			log.Fatalf("한국어 영상 생성 실패 (%d): %v", i, err)
		}
	} else { // 홀수 - 영어
		imagePath := fmt.Sprintf("images/output_%02d.png", i+1)
		englishAudioPath := fmt.Sprintf("audio/eng_%d.mp3", i/2)
		outputPath = fmt.Sprintf("videos/video_%d.mp4", i)
		
		if err := videoService.CreateVideoWithEnglish(imagePath, englishAudioPath, outputPath); err != nil {
			log.Fatalf("영어 영상 생성 실패 (%d): %v", i, err)
		}
	}
	
	log.Printf("영상 생성 완료: %d/%d", i+1, wordCount*2)
}

log.Println("개별 영상 생성 완료!")

// 5. 모든 영상을 하나로 합치기
// 지정된 날짜를 YYMMDD 형식으로 생성
finalFileName := fmt.Sprintf("%02d%02d%02d.mp4", targetDate.Year()%100, targetDate.Month(), targetDate.Day())

	videoPaths := make([]string, 0, wordCount*2)
	for i := 0; i < wordCount*2; i++ {
		videoPaths = append(videoPaths, fmt.Sprintf("video_%d.mp4", i))
	}

	err = videoService.ConcatenateVideos(
		videoPaths,
		finalFileName,
	)
if err != nil {
	log.Fatalf("영상 합치기 실패: %v", err)
}

log.Println("최종 영상 생성 완료!")

// 6. 중간 파일들 정리
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

log.Println("중간 파일들 정리 완료!")
log.Printf("최종 영상: %s", finalFileName)
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