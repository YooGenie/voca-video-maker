package video_type

import (
	"auto-video-service/service"
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

type StartService struct{}

func NewStartService() *StartService {
	return &StartService{}
}

func (s *StartService) CreateStartCommentVideo(ctx context.Context, targetDate time.Time, serviceType *string) {
	log.Println("🎬 스타트 멘트와 good 비디오를 생성합니다...")

	// 서비스 초기화
	imageService := service.NewImageService()
	videoConfig := service.VideoConfig{Width: 1920, Height: 1080}
	videoService := service.NewVideoService(imageService, videoConfig)

	// 출력 경로 설정
	outputPath := "template/start_comment.mp4"
	
	// 디렉토리 확인 및 생성
	if err := os.MkdirAll(filepath.Dir(outputPath), 0755); err != nil {
		log.Fatalf("디렉토리 생성 실패: %v", err)
	}

	// 임시 디렉토리 생성
	tempDir := "temp"
	if err := os.MkdirAll(tempDir, 0755); err != nil {
		log.Fatalf("임시 디렉토리 생성 실패: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// 1. 스타트 멘트 비디오 생성
	startVideoPath := filepath.Join(tempDir, "start_temp.mp4")
	if err := videoService.CreateStartCommentVideo(startVideoPath); err != nil {
		log.Fatalf("스타트 멘트 비디오 생성 실패: %v", err)
	}
	defer os.Remove(startVideoPath)

	// 2. good 비디오 생성
	goodVideoPath := filepath.Join(tempDir, "good_temp.mp4")
	if err := videoService.CreateGoodVideo(goodVideoPath); err != nil {
		log.Fatalf("good 비디오 생성 실패: %v", err)
	}
	defer os.Remove(goodVideoPath)

	// 3. 두 비디오 합치기
	fileListPath := filepath.Join(tempDir, "concat_list.txt")
	file, err := os.Create(fileListPath)
	if err != nil {
		log.Fatalf("파일 목록 생성 실패: %v", err)
	}
	defer os.Remove(fileListPath)

	// 절대 경로로 변환하여 안전하게 처리
	absStartPath, _ := filepath.Abs(startVideoPath)
	absGoodPath, _ := filepath.Abs(goodVideoPath)

	file.WriteString(fmt.Sprintf("file '%s'\n", absStartPath))
	file.WriteString(fmt.Sprintf("file '%s'\n", absGoodPath))
	file.Close()

	// ffmpeg로 두 비디오 합치기
	cmd := exec.Command("ffmpeg",
		"-f", "concat",
		"-safe", "0",
		"-i", fileListPath,
		"-c", "copy",
		"-y", // 기존 파일 덮어쓰기
		outputPath,
	)

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		log.Fatalf("비디오 합치기 실패: %v", err)
	}

	log.Printf("✅ 스타트 비디오 생성 완료: %s", outputPath)
}

