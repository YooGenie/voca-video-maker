package service

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

// VideoConfig 비디오 설정을 담는 구조체
type VideoConfig struct {
	Width  int
	Height int
}

// VideoService 비디오 생성 서비스
type VideoService struct {
	imageService *ImageService
	config       VideoConfig // 비디오 설정 추가
}

// NewVideoService 새로운 비디오 서비스 생성
func NewVideoService(imageService *ImageService, config VideoConfig) *VideoService {
	return &VideoService{
		imageService: imageService,
		config:       config,
	}
}

// CreateVideoWithAudioAndImage 이미지와 음성을 합쳐서 영상을 생성합니다
func (s *VideoService) CreateVideoWithAudioAndImage(
	imagePath string,
	audioPath string,
	outputPath string,
	duration float64,
) error {
	cmd := exec.Command("ffmpeg",
		"-loop", "1",
		"-i", imagePath,
		"-i", audioPath,
		"-c:v", "libx264",
		"-preset", "fast",
		"-profile:v", "baseline",
		"-level", "3.0",
		"-crf", "25",
		"-vf", fmt.Sprintf("scale=%d:%d,fps=30", s.config.Width, s.config.Height),
		"-c:a", "aac",
		"-b:a", "128k",
		"-ar", "44100",
		"-shortest",
		"-avoid_negative_ts", "make_zero",
		"-fflags", "+genpts",
		"-movflags", "+faststart",
		"-t", fmt.Sprintf("%.2f", duration),
		"-y",
		outputPath,
	)

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

// CreateVideoToAudioLength 이미지와 음성을 합쳐 오디오 길이에 맞는 영상을 생성합니다
func (s *VideoService) CreateVideoToAudioLength(
	imagePath string,
	audioPath string,
	outputPath string,
) error {
	cmd := exec.Command("ffmpeg",
		"-loop", "1",
		"-i", imagePath,
		"-i", audioPath,
		"-c:v", "libx264",
		"-preset", "fast",
		"-profile:v", "baseline",
		"-level", "3.0",
		"-crf", "25",
		"-vf", fmt.Sprintf("scale=%d:%d,format=yuv420p,fps=30", s.config.Width, s.config.Height),
		"-c:a", "aac",
		"-b:a", "128k",
		"-ar", "44100",
		"-shortest", // 오디오 길이에 맞춰 비디오 종료
		"-y",        // 기존 파일 덮어쓰기
		outputPath,
	)

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

// CreateVideoWithKorean 한국어 영상을 생성합니다 (0.5초 무음 + 한국어 음성)
func (s *VideoService) CreateVideoWithKorean(
	imagePath string,
	koreanAudioPath string,
	outputPath string,
) error {
	// 한국어 오디오에 0.5초 무음 추가 (싱크 맞춤)
	tempKoreanPath := koreanAudioPath[:len(koreanAudioPath)-4] + "_temp.mp3"
	koreanCmd := exec.Command("ffmpeg",
		"-i", koreanAudioPath,
		"-af", "apad=pad_dur=0.5",
		"-avoid_negative_ts", "make_zero",
		"-fflags", "+genpts",
		"-y",
		tempKoreanPath,
	)

	koreanCmd.Stdout = os.Stdout
	koreanCmd.Stderr = os.Stderr

	if err := koreanCmd.Run(); err != nil {
		return fmt.Errorf("한국어 오디오 처리 실패: %v", err)
	}

	// 비디오 생성 (모바일 호환성 최적화)
	cmd := exec.Command("ffmpeg",
		"-loop", "1",
		"-i", imagePath,
		"-i", tempKoreanPath,
		"-c:v", "libx264",
		"-preset", "fast",
		"-profile:v", "baseline",
		"-level", "3.0",
		"-crf", "25",
		"-vf", fmt.Sprintf("scale=%d:%d,format=yuv420p,fps=30", s.config.Width, s.config.Height),
		"-c:a", "aac",
		"-b:a", "128k",
		"-ar", "44100",
		"-shortest",
		"-avoid_negative_ts", "make_zero",
		"-fflags", "+genpts",
		"-movflags", "+faststart",
		"-y",
		outputPath,
	)

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("비디오 생성 실패: %v", err)
	}

	// 임시 파일 삭제
	os.Remove(tempKoreanPath)

	return nil
}

// CreateVideoWithEnglish 영어 영상을 생성합니다 (0.5초 무음 + 영어 음성 + 0.3초 + 영어 음성)
func (s *VideoService) CreateVideoWithEnglish(
	imagePath string,
	englishAudioPath string,
	outputPath string,
) error {
	// 영어 오디오를 2번 반복하고 사이에 0.4초 무음 추가
	tempEnglishPath := englishAudioPath[:len(englishAudioPath)-4] + "_temp.mp3"
	englishCmd := exec.Command("ffmpeg",
		"-i", englishAudioPath,
		"-i", englishAudioPath,
		"-filter_complex", "[0:a]apad=pad_dur=0.5[a1];[a1][1:a]concat=n=2:v=0:a=1[a]",
		"-map", "[a]",
		"-avoid_negative_ts", "make_zero",
		"-fflags", "+genpts",
		"-y",
		tempEnglishPath,
	)

	englishCmd.Stdout = os.Stdout
	englishCmd.Stderr = os.Stderr

	if err := englishCmd.Run(); err != nil {
		return fmt.Errorf("영어 오디오 처리 실패: %v", err)
	}

	// 0.5초 무음을 앞에 추가
	finalAudioPath := outputPath[:len(outputPath)-4] + "_final.mp3"
	finalCmd := exec.Command("ffmpeg",
		"-i", tempEnglishPath,
		"-af", "apad=pad_dur=0.5",
		"-y",
		finalAudioPath,
	)

	finalCmd.Stdout = os.Stdout
	finalCmd.Stderr = os.Stderr

	if err := finalCmd.Run(); err != nil {
		return fmt.Errorf("최종 오디오 처리 실패: %v", err)
	}

	//
	// 비디오 생성 (모바일 호환성 최적화)
	cmd := exec.Command("ffmpeg",
		"-loop", "1",
		"-i", imagePath,
		"-i", finalAudioPath,
		"-c:v", "libx264",
		"-preset", "fast",
		"-profile:v", "baseline",
		"-level", "3.0",
		"-crf", "25",
		"-vf", fmt.Sprintf("scale=%d:%d,format=yuv420p,fps=30", s.config.Width, s.config.Height),
		"-c:a", "aac",
		"-b:a", "128k",
		"-ar", "44100",
		"-shortest",
		"-avoid_negative_ts", "make_zero",
		"-fflags", "+genpts",
		"-movflags", "+faststart",
		"-y",
		outputPath,
	)

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("비디오 생성 실패: %v", err)
	}

	// 임시 파일들 삭제
	os.Remove(tempEnglishPath)
	os.Remove(finalAudioPath)

	return nil
}

// ConcatenateVideos 여러 영상을 하나로 합칩니다
func (s *VideoService) ConcatenateVideos(
	videoPaths []string,
	outputPath string,
) error {
	// videos 디렉토리에서 영상 파일들 찾기
	videosDir := "videos"

	// 파일 목록 생성
	fileListPath := filepath.Join(videosDir, "filelist.txt")
	file, err := os.Create(fileListPath)
	if err != nil {
		return fmt.Errorf("파일 목록 생성 실패: %v", err)
	}
	defer file.Close()

	// 각 영상 파일을 목록에 추가
	for _, videoPath := range videoPaths {
		line := fmt.Sprintf("file '%s'\n", videoPath)
		if _, err := file.WriteString(line); err != nil {
			return fmt.Errorf("파일 목록 작성 실패: %v", err)
		}
		fmt.Printf("영상 파일 추가: %s\n", videoPath)
	}
	file.Close()

	// ffmpeg로 영상들 합치기
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

	return cmd.Run()
}
