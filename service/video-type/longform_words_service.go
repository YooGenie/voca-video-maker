package video_type

import (
	"auto-video-service/entity"
	"auto-video-service/repository"
	"auto-video-service/service"
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

type LongformWordService struct{}

func NewLongformWordService() *LongformWordService {
	return &LongformWordService{}
}

func (s *LongformWordService) CreateLongformWords(ctx context.Context, targetDate time.Time, serviceType *string) {
	title, longformWords, err := s.getTitleByDate(ctx, targetDate)
	if err != nil {
		log.Fatalf("데이터 조회 실패: %v", err)
	}

	imageService := service.NewImageService()

	// 1. 타이틀 이미지 생성
	err = imageService.SetTitleOnImage(
		title.Title,
		"template/long.png",
		"template/titleImage.png",
	)
	if err != nil {
		log.Printf("타이틀 이미지 생성 실패: %v", err)
	} else {
		log.Println("타이틀 이미지 생성 완료!")
	}

	// 2. 본문 이미지 생성
	newTemplateImagePath := "template/long.png"
	words := make([]string, len(longformWords))
	meanings := make([]string, len(longformWords))
	pronunciations := make([]string, len(longformWords))
	for i, lw := range longformWords {
		words[i] = lw.Word
		meanings[i] = lw.Meaning
		pronunciations[i] = lw.PronunciationKr
	}

	err = imageService.GenerateBasicImages(
		newTemplateImagePath,
		words,
		meanings,
		pronunciations,
		"images/output",
		len(longformWords)*2,
	)
	if err != nil {
		log.Fatalf("이미지 생성 실패: %v", err)
	}
	log.Println("이미지 생성 완료!")

	longformConfig := service.VideoConfig{Width: 1920, Height: 1080}
	videoService := service.NewVideoService(imageService, longformConfig)
	audioService := service.NewAudioService()

	audioDir := "audio"
	if err := os.MkdirAll(audioDir, 0755); err != nil {
		log.Fatalf("audio 디렉토리 생성 실패: %v", err)
	}

	videosDir := "videos"
	if err := os.MkdirAll(videosDir, 0755); err != nil {
		log.Fatalf("videos 디렉토리 생성 실패: %v", err)
	}

	titleVideoPath, err := s.createTitleVideo(videoService, audioService, audioDir, videosDir)
	if err != nil {
		log.Fatalf("타이틀 영상 제작 실패: %v", err)
	}

	videoPaths := []string{titleVideoPath}

	log.Println("🎤 영어 단어 원어민 음성을 생성합니다...")
	log.Printf("영어 단어 목록: %v", words)
	for i, word := range words {
		audioPath := fmt.Sprintf("%s/eng_%d.mp3", audioDir, i)
		if err := audioService.CreateNativeEnglishAudio(word, audioPath); err != nil {
			log.Printf("영어 원어민 음성 생성 실패 (%s): %v", word, err)
		}
	}

	log.Println("🎤 한국어 단어 음성을 생성합니다...")
	log.Printf("한국어 뜻 목록: %v", meanings)
	for i, meaning := range meanings {
		audioPath := fmt.Sprintf("%s/kor_%d.mp3", audioDir, i)
		if err := audioService.CreateKoreanAudioWithRate(meaning, audioPath, 125); err != nil {
			log.Printf("한국어 음성 생성 실패 (%s): %v", meaning, err)
		}
	}

	log.Println("음성 파일 생성 완료!")

	for i := 0; i < len(longformWords)*2; i++ {
		var videoFileName string

		if i%2 == 0 { // 짝수 - 한국어
			imagePath := fmt.Sprintf("images/output_%02d.png", i+1)
			koreanAudioPath := fmt.Sprintf("audio/kor_%d.mp3", i/2)
			videoFileName = fmt.Sprintf("video_%d.mp4", i)

			if err := videoService.CreateVideoWithKorean(imagePath, koreanAudioPath, filepath.Join(videosDir, videoFileName)); err != nil {
				log.Fatalf("한국어 영상 생성 실패 (%d): %v", i, err)
			}
		} else { // 홀수 - 영어
			imagePath := fmt.Sprintf("images/output_%02d.png", i+1)
			englishAudioPath := fmt.Sprintf("audio/eng_%d.mp3", i/2)
			videoFileName = fmt.Sprintf("video_%d.mp4", i)

			if err := videoService.CreateVideoWithEnglish(imagePath, englishAudioPath, filepath.Join(videosDir, videoFileName)); err != nil {
				log.Fatalf("영어 영상 생성 실패 (%d): %v", i, err)
			}
		}
		videoPaths = append(videoPaths, videoFileName)
		log.Printf("영상 생성 완료: %d/%d", i+1, len(longformWords)*2)
	}

	log.Println("개별 영상 생성 완료!")

	finalFileName := fmt.Sprintf("%02d%02d%02d_longform.mp4", targetDate.Year()%100, targetDate.Month(), targetDate.Day())

	err = videoService.ConcatenateVideos(
		videoPaths,
		finalFileName,
	)

	if err != nil {
		log.Fatalf("영상 합치기 실패: %v", err)
	}

	log.Println("최종 영상 생성 완료!")

	log.Println("중간 파일들 정리 중...")

	if files, err := os.ReadDir("images"); err == nil {
		for _, file := range files {
			if !file.IsDir() {
				os.Remove(filepath.Join("images", file.Name()))
			}
		}
	}

	if files, err := os.ReadDir("audio"); err == nil {
		for _, file := range files {
			if !file.IsDir() {
				os.Remove(filepath.Join("audio", file.Name()))
			}
		}
	}

	if files, err := os.ReadDir("videos"); err == nil {
		for _, file := range files {
			if !file.IsDir() && file.Name() != "title_video.mp4" {
				os.Remove(filepath.Join("videos", file.Name()))
			}
		}
	}

	log.Println("중간 파일들 정리 완료!")
	log.Printf("최종 영상: %s", finalFileName)

	fmt.Println("\n📚 생성된 Longform 단어 목록:")
	fmt.Println("=" + fmt.Sprintf("%*s", 40, "") + "=")
	for i := 0; i < len(words); i++ {
		fmt.Printf("%d) %s (%s)\n", i+1, words[i], meanings[i])
	}
	fmt.Println("=" + fmt.Sprintf("%*s", 40, "") + "=")
}

func (s *LongformWordService) createTitleVideo(videoService *service.VideoService, audioService *service.AudioService, audioDir, videosDir string) (string, error) {
	// 음성 속도 설정 (기본값: 175)
	slowRate := 123

	// 1. 두 부분으로 나누어 음성 파일 생성
	audioPart1Path := filepath.Join(audioDir, "title_part1.mp3")
	if err := audioService.CreateKoreanAudioWithRate("누워서 영어공부", audioPart1Path, slowRate); err != nil {
		return "", fmt.Errorf("타이틀 음성(part1) 생성 실패: %w", err)
	}

	audioPart2Path := filepath.Join(audioDir, "title_part2.mp3")
	if err := audioService.CreateKoreanAudioWithRate("시작합니다", audioPart2Path, slowRate); err != nil {
		return "", fmt.Errorf("타이틀 음성(part2) 생성 실패: %w", err)
	}

	// 2. 1.5초짜리 무음 오디오 생성
	silenceAudioPath := filepath.Join(audioDir, "silence.mp3")
	cmd := exec.Command("ffmpeg", "-f", "lavfi", "-i", "anullsrc=r=22050:cl=mono", "-t", "1.5", "-ab", "128k", "-acodec", "libmp3lame", "-y", silenceAudioPath)
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("무음 오디오 생성 실패: %w", err)
	}

	// 3. 음성 파일들 합치기 (concat 필터 사용)
	concatAudioPath := filepath.Join(audioDir, "longform_title.mp3")
	concatCmd := exec.Command("ffmpeg",
		"-i", audioPart1Path,
		"-i", silenceAudioPath,
		"-i", audioPart2Path,
		"-i", silenceAudioPath,
		"-filter_complex", "[0:a]aformat=sample_fmts=s16:sample_rates=22050:channel_layouts=mono[a0];[1:a]aformat=sample_fmts=s16:sample_rates=22050:channel_layouts=mono[a1];[2:a]aformat=sample_fmts=s16:sample_rates=22050:channel_layouts=mono[a2];[a0][a1][a2]concat=n=4:v=0:a=1[out]",
		"-map", "[out]",
		"-acodec", "libmp3lame", // 인코더를 libmp3lame으로 변경
		"-ab", "128k", // 비트레이트 설정
		"-y", concatAudioPath,
	)
	if err := concatCmd.Run(); err != nil {
		return "", fmt.Errorf("타이틀 음성 파일 합치기 실패: %w", err)
	}

	// 5. 최종 타이틀 영상 생성
	titleVideoPath := "title_video.mp4"
	if err := videoService.CreateVideoToAudioLength("template/titleImage.png", concatAudioPath, filepath.Join(videosDir, titleVideoPath)); err != nil {
		return "", fmt.Errorf("타이틀 영상 생성 실패: %w", err)
	}

	return titleVideoPath, nil
}

func (s *LongformWordService) getTitleByDate(ctx context.Context, targetDate time.Time) (*entity.Title, []entity.LongformWord, error) {
	titleRepo := repository.TitleRepository()
	longformWordRepo := repository.LongformWordRepository()

	dateStr := targetDate.Format("20060102")

	title, err := titleRepo.FindByDate(ctx, dateStr)
	if err != nil {
		return nil, nil, fmt.Errorf("타이틀 조회 실패: %w", err)
	}

	longformWords, err := longformWordRepo.FindByDate(ctx, dateStr)
	if err != nil {
		return nil, nil, fmt.Errorf("Longform 단어 조회 실패: %w", err)
	}

	if len(longformWords) == 0 {
		return nil, nil, fmt.Errorf("%s에 해당하는 Longform 단어가 없습니다", dateStr)
	}

	log.Printf("데이터베이스에서 %s 날짜의 타이틀과 %d개 Longform 단어를 조회했습니다.", dateStr, len(longformWords))

	return title, longformWords, nil
}