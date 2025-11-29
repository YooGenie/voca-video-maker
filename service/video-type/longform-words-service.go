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

	// 서비스 초기화
	imageService := service.NewImageService()
	longformConfig := service.VideoConfig{Width: 1920, Height: 1080}
	videoService := service.NewVideoService(imageService, longformConfig)
	audioService := service.NewAudioService()

	// 디렉토리 생성
	audioDir := "audio"
	if err := os.MkdirAll(audioDir, 0755); err != nil {
		log.Fatalf("audio 디렉토리 생성 실패: %v", err)
	}
	videosDir := "videos"
	if err := os.MkdirAll(videosDir, 0755); err != nil {
		log.Fatalf("videos 디렉토리 생성 실패: %v", err)
	}

	// 1. 타이틀 시퀀스 생성 (이미지, 음성, 비디오)
	titleVideoPath, err := s.createTitleSequence(title.Title, title.SubTitle, imageService, audioService, videoService, audioDir, videosDir)
	if err != nil {
		log.Fatalf("타이틀 시퀀스 생성 실패: %v", err)
	}
	videoPaths := []string{titleVideoPath}

	// 1-1. 스타트 코멘트 비디오 연결 (이미 생성된 template/start_comment.mp4 사용)
	startCommentVideoPath := "template/start_comment.mp4"
	if _, err := os.Stat(startCommentVideoPath); os.IsNotExist(err) {
		log.Fatalf("스타트 코멘트 비디오를 찾을 수 없습니다: %s", startCommentVideoPath)
	}
	videoPaths = append(videoPaths, startCommentVideoPath)
	log.Println("✅ 스타트 코멘트 비디오 연결 완료!")

	// 2. 본문 이미지 생성
	words := make([]string, len(longformWords))
	meanings := make([]string, len(longformWords))
	pronunciations := make([]string, len(longformWords))
	for i, lw := range longformWords {
		words[i] = lw.Word
		meanings[i] = lw.Meaning
		pronunciations[i] = lw.PronunciationKr
	}

	if err := imageService.GenerateBasicImages("template/long.png", words, meanings, pronunciations, "images/output", len(longformWords)*2); err != nil {
		log.Fatalf("이미지 생성 실패: %v", err)
	}
	log.Println("✅ 본문 이미지 생성 완료!")

	// 3. 본문 음성 생성
	log.Println("🎤 영어 단어 원어민 음성을 생성합니다...")
	for i, word := range words {
		audioPath := fmt.Sprintf("%s/eng_%d.mp3", audioDir, i)
		if err := audioService.CreateNativeEnglishAudio(word, audioPath); err != nil {
			log.Fatalf("영어 원어민 음성 생성 실패 (%s): %v", word, err)
		}
	}

	log.Println("🎤 한국어 단어 음성을 생성합니다...")
	for i, meaning := range meanings {
		audioPath := fmt.Sprintf("%s/kor_%d.mp3", audioDir, i)
		if err := audioService.CreateKoreanAudioWithRate(meaning, audioPath, 125); err != nil {
			log.Fatalf("한국어 음성 생성 실패 (%s): %v", meaning, err)
		}
	}
	log.Println("✅ 본문 음성 파일 생성 완료!")

	// 4. 본문 비디오 생성
	for i := 0; i < len(longformWords)*2; i++ {
		var videoFileName string
		if i%2 == 0 { // 짝수 - 한국어
			imagePath := fmt.Sprintf("images/output_%02d.png", i+1)
			koreanAudioPath := fmt.Sprintf("audio/kor_%d.mp3", i/2)
			videoFileName = fmt.Sprintf("video_%d.mp4", i)
			if err := videoService.CreateVideoWithKorean(imagePath, koreanAudioPath, filepath.Join(videosDir, videoFileName), 1); err != nil {
				log.Fatalf("한국어 영상 생성 실패 (%d): %v", i, err)
			}
		} else { // 홀수 - 영어
			imagePath := fmt.Sprintf("images/output_%02d.png", i+1)
			englishAudioPath := fmt.Sprintf("audio/eng_%d.mp3", i/2)
			videoFileName = fmt.Sprintf("video_%d.mp4", i)
			if err := videoService.CreateVideoWithEnglish(imagePath, englishAudioPath, filepath.Join(videosDir, videoFileName), 2); err != nil {
				log.Fatalf("영어 영상 생성 실패 (%d): %v", i, err)
			}
		}
		videoPaths = append(videoPaths, filepath.Join(videosDir, videoFileName))
		log.Printf("📹 영상 생성 완료: %d/%d", i+1, len(longformWords)*2)
	}
	log.Println("✅ 개별 영상 생성 완료!")

	// 5. 최종 영상 합치기
	finalFileName := fmt.Sprintf("%02d%02d%02d_longform.mp4", targetDate.Year()%100, targetDate.Month(), targetDate.Day())
	if err = videoService.ConcatenateVideos(videoPaths, finalFileName); err != nil {
		log.Fatalf("영상 합치기 실패: %v", err)
	}
	log.Println("✅ 최종 영상 생성 완료!")

	// 6. 중간 파일 정리
	log.Println("🗑️ 중간 파일들 정리 중...")
	if err := s.cleanupFiles(videoPaths, len(longformWords)); err != nil {
		log.Printf("중간 파일 정리에 실패했습니다: %v", err)
	}
	log.Println("✅ 중간 파일 정리 완료!")
}

func (s *LongformWordService) cleanupFiles(videoPaths []string, wordCount int) error {
	// 1. 중간 비디오 파일 삭제 (타이틀 영상 포함)
	for _, path := range videoPaths {
		if err := os.Remove(path); err != nil {
			log.Printf("중간 비디오 파일 삭제 실패 (%s): %v", path, err)
		}
	}

	// 2. 생성된 본문 이미지 삭제
	for i := 0; i < wordCount*2; i++ {
		imagePath := fmt.Sprintf("images/output_%02d.png", i+1)
		if err := os.Remove(imagePath); err != nil {
			log.Printf("이미지 파일 삭제 실패 (%s): %v", imagePath, err)
		}
	}

	// 3. 생성된 타이틀 이미지 삭제
	if err := os.Remove("template/titleImage.png"); err != nil {
		log.Printf("타이틀 이미지 파일 삭제 실패: %v", err)
	}

	// 4. 생성된 본문 오디오 파일 삭제
	for i := 0; i < wordCount; i++ {
		engAudioPath := fmt.Sprintf("audio/eng_%d.mp3", i)
		if err := os.Remove(engAudioPath); err != nil {
			log.Printf("영어 오디오 파일 삭제 실패 (%s): %v", engAudioPath, err)
		}
		korAudioPath := fmt.Sprintf("audio/kor_%d.mp3", i)
		if err := os.Remove(korAudioPath); err != nil {
			log.Printf("한국어 오디오 파일 삭제 실패 (%s): %v", korAudioPath, err)
		}
	}

	return nil
}

// createTitleSequence는 타이틀 이미지, 오디오, 비디오를 모두 생성합니다.
func (s *LongformWordService) createTitleSequence(
	title, subTitle string,
	imageService *service.ImageService,
	audioService *service.AudioService,
	videoService *service.VideoService,
	audioDir, videosDir string,
) (string, error) {
	log.Println("🎬 타이틀 시퀀스를 생성합니다...")

	// 1. 타이틀 이미지 생성
	titleImagePath := "template/titleImage.png"
	if err := imageService.SetTitleOnImage(title, subTitle, "template/title.png", titleImagePath); err != nil {
		return "", fmt.Errorf("타이틀 이미지 생성 실패: %w", err)
	}
	log.Println("✅ 타이틀 이미지 생성 완료")

	// 2. 타이틀 오디오 생성 (이미지에 표시된 title과 subTitle을 음성으로 변환)
	slowRate := 150
	audioPart1Path := filepath.Join(audioDir, "title_part1.mp3")
	defer os.Remove(audioPart1Path)
	if err := audioService.CreateKoreanAudioWithRate(title, audioPart1Path, slowRate); err != nil {
		return "", fmt.Errorf("타이틀 음성(part1) 생성 실패: %w", err)
	}

	silenceAudioPath := filepath.Join(audioDir, "silence.mp3")
	defer os.Remove(silenceAudioPath)
	cmd := exec.Command("ffmpeg", "-f", "lavfi", "-i", "anullsrc=r=22050:cl=mono", "-t", "1.5", "-ab", "128k", "-acodec", "libmp3lame", "-y", silenceAudioPath)
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("무음 오디오 생성 실패: %w", err)
	}

	concatAudioPath := filepath.Join(audioDir, "longform_title.mp3")
	defer os.Remove(concatAudioPath)

	// subTitle이 있는 경우와 없는 경우를 분기 처리
	if subTitle != "" {
		audioPart2Path := filepath.Join(audioDir, "title_part2.mp3")
		defer os.Remove(audioPart2Path)
		if err := audioService.CreateKoreanAudioWithRate(subTitle, audioPart2Path, slowRate); err != nil {
			return "", fmt.Errorf("타이틀 음성(part2) 생성 실패: %w", err)
		}

		// title + 무음 + subTitle + 무음 합치기
		concatCmd := exec.Command("ffmpeg",
			"-i", audioPart1Path,
			"-i", silenceAudioPath,
			"-i", audioPart2Path,
			"-i", silenceAudioPath,
			"-filter_complex", "[0:a]aformat=sample_fmts=s16:sample_rates=22050:channel_layouts=mono[a0];[1:a]aformat=sample_fmts=s16:sample_rates=22050:channel_layouts=mono[a1];[2:a]aformat=sample_fmts=s16:sample_rates=22050:channel_layouts=mono[a2];[3:a]aformat=sample_fmts=s16:sample_rates=22050:channel_layouts=mono[a3];[a0][a1][a2][a3]concat=n=4:v=0:a=1[out]",
			"-map", "[out]",
			"-acodec", "libmp3lame",
			"-ab", "128k",
			"-y", concatAudioPath,
		)
		if err := concatCmd.Run(); err != nil {
			return "", fmt.Errorf("타이틀 음성 파일 합치기 실패: %w", err)
		}
	} else {
		// subTitle이 없는 경우: title + 무음만 합치기
		concatCmd := exec.Command("ffmpeg",
			"-i", audioPart1Path,
			"-i", silenceAudioPath,
			"-filter_complex", "[0:a]aformat=sample_fmts=s16:sample_rates=22050:channel_layouts=mono[a0];[1:a]aformat=sample_fmts=s16:sample_rates=22050:channel_layouts=mono[a1];[a0][a1]concat=n=2:v=0:a=1[out]",
			"-map", "[out]",
			"-acodec", "libmp3lame",
			"-ab", "128k",
			"-y", concatAudioPath,
		)
		if err := concatCmd.Run(); err != nil {
			return "", fmt.Errorf("타이틀 음성 파일 합치기 실패: %w", err)
		}
	}
	log.Println("✅ 타이틀 오디오 생성 완료")

	// 3. 최종 타이틀 영상 생성
	titleVideoPath := "title_video.mp4"
	fullTitleVideoPath := filepath.Join(videosDir, titleVideoPath)
	if err := videoService.CreateVideoToAudioLength(titleImagePath, concatAudioPath, fullTitleVideoPath); err != nil {
		return "", fmt.Errorf("타이틀 영상 생성 실패: %w", err)
	}
	log.Println("✅ 타이틀 비디오 생성 완료")

	return fullTitleVideoPath, nil
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
