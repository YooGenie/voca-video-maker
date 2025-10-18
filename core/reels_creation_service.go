package core

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"auto-video-service/dto"
)

type ReelsCreationService struct{}

func NewReelsCreationService() *ReelsCreationService {
	return &ReelsCreationService{}
}

// CreateCompleteReels - 릴스 제작 전체 과정 (이미지 생성 → 음성 생성 → 비디오 생성 → 합치기 → 정리)
func (s *ReelsCreationService) CreateCompleteReels(ctx context.Context, request dto.VideoCreationRequest, contentData dto.ContentData, templateConfig dto.TemplateConfig) dto.VideoCreationResponse {
	response := dto.VideoCreationResponse{
		ContentCount: contentData.Count,
		Success:      false,
	}

	// 이미지 서비스 생성
	imageService := NewImageService()

	// 1. 조회된 컨텐츠 개수만큼 이미지 생성
	contentCount := contentData.Count

	// 먼저 컨텐츠 개수를 표시하는 이미지 생성
	err := imageService.GenerateOptionalImage(
		templateConfig.BaseTemplate,                    // 기본 이미지 템플릿
		fmt.Sprintf("%d", contentCount),               // contentCount를 문자열로 변환
		templateConfig.CountTemplate,                  // 출력 파일명
		request.ServiceType,                          // 서비스 타입 (W 또는 I)
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
	err = imageService.GenerateBasicImages(
		newTemplateImagePath,  // 컨텐츠 개수가 표시된 이미지 템플릿
		contentData.Primary,   // 영어 단어들 또는 숙어들
		contentData.Secondary, // 한국어 번역들 또는 의미들
		contentData.Tertiary,  // 발음들 또는 예문들
		"images/output",       // 출력 파일 접두사 (images 디렉토리에 저장)
		contentCount * 2,      // 생성할 이미지 개수 (동적)
	)
	if err != nil {
		log.Printf("이미지 생성 실패: %v", err)
		response.Error = err
		return response
	}
	log.Println("이미지 생성 완료!")

	// 2. 비디오 서비스 생성
	reelsConfig := VideoConfig{Width: 1080, Height: 1920}
	videoService := NewVideoService(imageService, reelsConfig)

	// 3. 각 컨텐츠에 대한 음성 파일 생성
	audioDir := "audio"
	if err := os.MkdirAll(audioDir, 0755); err != nil {
		log.Printf("audio 디렉토리 생성 실패: %v", err)
		response.Error = err
		return response
	}

	// 영어 컨텐츠 원어민 음성 생성
	log.Println("🎤 영어 컨텐츠 원어민 음성을 생성합니다...")
	for i, content := range contentData.Primary {
		audioPath := fmt.Sprintf("%s/eng_%d.mp3", audioDir, i)
		if err := videoService.GenerateNativeEnglishAudio(content, audioPath); err != nil {
			log.Printf("영어 원어민 음성 생성 실패 (%s): %v", content, err)
		}
	}

	// 한국어 컨텐츠 음성 생성
	log.Println("🎤 한국어 컨텐츠 음성을 생성합니다...")
	for i, content := range contentData.Secondary {
		audioPath := fmt.Sprintf("%s/kor_%d.mp3", audioDir, i)
		if err := videoService.GenerateKoreanAudioWithRate(content, audioPath, 175); err != nil {
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

	// 4. 각 이미지에 음성을 추가한 영상 생성
	for i := 0; i < contentCount*2; i++ {
		var outputPath string
		
		if i%2 == 0 { // 짝수 - 한국어
			imagePath := fmt.Sprintf("images/output_%02d.png", i+1)
			koreanAudioPath := fmt.Sprintf("audio/kor_%d.mp3", i/2)
			outputPath = fmt.Sprintf("videos/video_%d.mp4", i)
			
			if err := videoService.CreateVideoWithKorean(imagePath, koreanAudioPath, outputPath); err != nil {
				log.Printf("한국어 영상 생성 실패 (%d): %v", i, err)
				response.Error = err
				return response
			}
		} else { // 홀수 - 영어
			imagePath := fmt.Sprintf("images/output_%02d.png", i+1)
			englishAudioPath := fmt.Sprintf("audio/eng_%d.mp3", i/2)
			outputPath = fmt.Sprintf("videos/video_%d.mp4", i)
			
			if err := videoService.CreateVideoWithEnglish(imagePath, englishAudioPath, outputPath); err != nil {
				log.Printf("영어 영상 생성 실패 (%d): %v", i, err)
				response.Error = err
				return response
			}
		}
		
		log.Printf("영상 생성 완료: %d/%d", i+1, contentCount*2)
	}

	log.Println("개별 영상 생성 완료!")

	// 5. 모든 영상을 하나로 합치기
	// 지정된 날짜를 YYMMDD 형식으로 생성
	finalFileName := fmt.Sprintf("%02d%02d%02d.mp4", request.TargetDate.Year()%100, request.TargetDate.Month(), request.TargetDate.Day())
	response.FinalFileName = finalFileName

	videoPaths := make([]string, 0, contentCount*2)
	for i := 0; i < contentCount*2; i++ {
		videoPaths = append(videoPaths, fmt.Sprintf("video_%d.mp4", i))
	}

	err = videoService.ConcatenateVideos(
		videoPaths,
		finalFileName,
	)
	if err != nil {
		log.Printf("영상 합치기 실패: %v", err)
		response.Error = err
		return response
	}

	log.Println("최종 영상 생성 완료!")

	// 6. 중간 파일들 정리
	s.cleanupTempFiles()

	log.Println("중간 파일들 정리 완료!")
	log.Printf("최종 영상: %s", finalFileName)

	response.Success = true
	return response
}

// cleanupTempFiles - 중간 파일들 정리
func (s *ReelsCreationService) cleanupTempFiles() {
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
