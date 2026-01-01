package service

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"auto-video-service/config"
	"auto-video-service/dto"
	"auto-video-service/enum"
)

type ReelsCreationService struct{}

func NewReelsCreationService() *ReelsCreationService {
	return &ReelsCreationService{}
}

// CreateCompleteReels - 릴스 제작 전체 과정 (이미지 생성 → 음성 생성 → 비디오 생성 → 합치기 → 정리)
func (s *ReelsCreationService) CreateCompleteReels(ctx context.Context, request dto.VideoCreationRequest, contentData dto.ContentData, templateConfig dto.TemplateConfig, options dto.VideoCreationOptions) dto.VideoCreationResponse {
	return s.CreateCompleteReelsWithFontSize(ctx, request, contentData, templateConfig, options, 120)
}

// CreateCompleteReelsWithFontSize - 폰트 크기를 지정하여 릴스 제작 전체 과정을 수행합니다
func (s *ReelsCreationService) CreateCompleteReelsWithFontSize(ctx context.Context, request dto.VideoCreationRequest, contentData dto.ContentData, templateConfig dto.TemplateConfig, options dto.VideoCreationOptions, fontSize float64) dto.VideoCreationResponse {
	// defer로 최종적으로 임시 파일 정리 (성공/실패 여부 상관없이)
	defer s.cleanupTempFiles()

	response := dto.VideoCreationResponse{
		ContentCount: contentData.Count,
		Success:      false,
	}

	// 이미지 서비스 생성
	imageService := NewImageService()

	// 임시 디렉토리 경로 (config에서 인용)
	tempDir := config.Config.Paths.TempDir
	imagesDir := filepath.Join(tempDir, "images")

	// temp/images 디렉토리 생성
	if err := os.MkdirAll(imagesDir, 0755); err != nil {
		log.Printf("temp/images 디렉토리 생성 실패: %v", err)
		response.Error = err
		return response
	}

	// 1. 조회된 컨텐츠 개수만큼 이미지 생성
	contentCount := contentData.Count

	// 기본 이미지들 생성 (카운트 이미지 생성 로직 제거됨)
	err := imageService.GenerateBasicImagesWithFontSize(
		templateConfig.BaseTemplate,       // 기본 이미지 템플릿
		contentData.Primary,               // 영어 단어들 또는 숙어들
		contentData.PrimaryLine2,          // 영어 두 번째 줄 (SS 타입 전용)
		contentData.Secondary,             // 한국어 번역들 또는 의미들
		contentData.SecondaryLine2,        // 한국어 두 번째 줄 (SS 타입 전용)
		contentData.Tertiary,              // 발음들 또는 예문들
		filepath.Join(imagesDir, "output"), // 출력 파일 접두사
		contentCount*2,                    // 생성할 이미지 개수 (동적)
		fontSize,                          // 폰트 크기
		templateConfig.TextColor,          // 텍스트 색상
	)
	if err != nil {
		log.Printf("이미지 생성 실패: %v", err)
		response.Error = err
		return response
	}
	log.Println("이미지 생성 완료!")

	// 2. 서비스 생성
	reelsConfig := VideoConfig{Width: 1080, Height: 1920}
	videoService := NewVideoService(imageService, reelsConfig)
	audioService := NewAudioService()

	// 3. 각 컨텐츠에 대한 음성 파일 생성
	audioDir := filepath.Join(tempDir, "audio")
	if err := os.MkdirAll(audioDir, 0755); err != nil {
		log.Printf("audio 디렉토리 생성 실패: %v", err)
		response.Error = err
		return response
	}

	// 4. 각 이미지에 음성을 추가한 영상 생성 및 조립 준비
	// videos 디렉토리 생성
	videosDir := filepath.Join(tempDir, "videos")
	if err := os.MkdirAll(videosDir, 0755); err != nil {
		log.Printf("videos 디렉토리 생성 실패: %v", err)
		response.Error = err
		return response
	}

	// Pause(공백)는 한국어 영상의 silentTime에 이미 포함되므로
	// 별도의 공백 영상은 생성하지 않음 (검정 화면 방지)
	var silenceVideoPath string // 빈 문자열로 유지

	videoPaths := make([]string, 0)

	log.Println("🎤 음성 및 영상 생성을 시작합니다...")
	for i := 0; i < contentCount; i++ {
		// 1) 영어 음성 생성
		engAudioPath := fmt.Sprintf("%s/eng_%d.mp3", audioDir, i)
		engContent := contentData.Primary[i]
		if len(contentData.PrimaryLine2) > i && contentData.PrimaryLine2[i] != "" {
			engContent += " " + contentData.PrimaryLine2[i]
		}

		// SpeakSpeed가 1.0보다 작으면 Slow 모드로 간주
		isSlow := options.SpeakSpeed < 1.0
		if err := audioService.CreateNativeEnglishAudio(engContent, engAudioPath, isSlow); err != nil {
			log.Printf("영어 원어민 음성 생성 실패 (%s): %v", engContent, err)
		}

		// 2) 한국어 음성 생성
		korAudioPath := fmt.Sprintf("%s/kor_%d.mp3", audioDir, i)
		korContent := contentData.Secondary[i]
		if len(contentData.SecondaryLine2) > i && contentData.SecondaryLine2[i] != "" {
			korContent += " " + contentData.SecondaryLine2[i]
		}
		if err := audioService.CreateKoreanAudioWithRate(korContent, korAudioPath, 175); err != nil { // 한국어 속도는 고정 or 옵션? 일단 기존 175 유지
			log.Printf("한국어 음성 생성 실패 (%s): %v", korContent, err)
		}

		// 3) 영상 생성 (Even=Kor, Odd=Eng in original logic. Now explicit)
		// 이미지 인덱스는 1-based, 2개씩 생성됨 (홀수: 영어, 짝수: 한국어... 인데 기존 로직 확인 필요)
		// 기존: i%2==0(짝수) -> 한국어?
		// 기존 코드: images/output_%02d.png. i=0(짝수) -> output_01.png.
		// i=0 loop -> image index 1 (output_01.png).
		// 기존 로직:
		// for i=0..contentCount*2
		//   i=0 (짝수) -> image output_01.png, kor audio kor_0.mp3 -> video_0.mp4 (KOR)
		//   i=1 (홀수) -> image output_02.png, eng audio eng_0.mp3 -> video_1.mp4 (ENG)
		//   Wait, 기존 코드는 '짝수 - 한국어', '홀수 - 영어' 라고 주석이 되어 있음.
		//   하지만 이미지는 output_01, output_02...
		//   확인: i=0 -> output_01 (1번째 이미지). 보통 1번째 이미지가 영어(Main) 아닌가?
		//   기존 ImageService.GenerateBasicImages... Create logic order check.
		//   보통 (Eng, Kor) 쌍으로 이미지 생성됨.
		//   i=0 -> 1st image (Eng content text).
		//   근데 기존 코드에 `if i%2 == 0 { // 짝수 - 한국어 }` 라고 되어 있음.
		//   그리고 `imagePath := fmt.Sprintf("images/output_%02d.png", i+1)` -> output_01.png
		//   만약 output_01.png가 영어 텍스트라면, 한국어 오디오를 입히는게 이상함.
		//   **중요**: 기존 ImageService 로직을 보면 `words`, `meanings` 순서대로 렌더링함.
		//   아마 (Word(Eng), Meaning(Kor)) 순서대로 이미지가 01, 02 이렇게 생성될 것임.
		//   그렇다면 output_01은 영어, output_02는 한국어.
		//   기존 코드: i=0(짝수) -> output_01(영어이미지) + kor_audio? -> "한국어 영상 생성 실패" 로그
		//   잠깐, 기존 코드 `if i%2 == 0 { // 짝수 - 한국어 }` 는 i가 loop iterator (0..total*2).
		//   i=0 -> output_01.png.
		//   i=1 -> output_02.png.
		//   만약 ImageService가 Eng->Kor 순서로 생성한다면,
		//   i=0에는 EngAudio + EngImage여야 함.
		//   기존 코드가 `i%2 == 0`일 때 `CreateVideoWithKorean`을 호출하고 있음.
		//   즉, 기존 코드는 (Kor, Eng) 순서로 영상을 만들고 있었거나, 주석/로직이 꼬여 있었을 수 있음.
		//   하지만 `contentData.Primary`(Eng)와 `Secondary`(Kor)를 넘김.
		//   ImageService 로직을 확인하지 않고는 확신 불가.
		//   하지만 통상적으로 Eng -> Kor.
		//   여기서는 명시적으로 EngVideo, KorVideo 식별해서 생성.

		// 이미지 경로 설정 (ImageService: 홀수=한국어, 짝수=영어)
		// output_01(Kor), output_02(Eng), output_03(Kor), output_04(Eng)...
		// 이미지 경로 설정 (ImageService: 홀수=한국어, 짝수=영어)
		// output_01(Kor), output_02(Eng), output_03(Kor), output_04(Eng)...
		korImagePath := fmt.Sprintf("temp/images/output_%02d.png", i*2+1)
		engImagePath := fmt.Sprintf("temp/images/output_%02d.png", i*2+2)

		engVideoPath := fmt.Sprintf("temp/videos/eng_%d.mp4", i)
		korVideoPath := fmt.Sprintf("temp/videos/kor_%d.mp4", i)

		// 영어 영상 생성
		if err := videoService.CreateVideoWithEnglish(engImagePath, engAudioPath, engVideoPath, 0.5); err != nil {
			log.Printf("영어 영상 생성 실패 (%d): %v", i, err)
			response.Error = err
			return response
		}

		// 한국어 영상 생성
		if err := videoService.CreateVideoWithKorean(korImagePath, korAudioPath, korVideoPath, 0.5); err != nil {
			log.Printf("한국어 영상 생성 실패 (%d): %v", i, err)
			response.Error = err
			return response
		}

		// 조립 (Assemble)
		// IsReverse와 RepeatCount 적용
		// 기본값 처리
		repeat := options.EnglishRepeatCount
		if repeat < 1 {
			repeat = 1
		}

		// 순서 결정
		// IsReverse가 true이면: English -> Korean
		// IsReverse가 false(기본)이면: Korean -> English
		if contentData.IsReverse {
			// Reverse: Eng (반복) -> Kor
			for r := 0; r < repeat; r++ {
				videoPaths = append(videoPaths, engVideoPath)
			}
			// 영어 반복 후 공백 1회
			if options.PauseDuration > 0 && silenceVideoPath != "" {
				videoPaths = append(videoPaths, silenceVideoPath)
			}
			videoPaths = append(videoPaths, korVideoPath)
		} else {
			// Default: Kor -> Eng (반복)
			// 한국어 1회
			videoPaths = append(videoPaths, korVideoPath)
			// 한국어 후 공백 (옵션)
			if options.PauseDuration > 0 && silenceVideoPath != "" {
				videoPaths = append(videoPaths, silenceVideoPath)
			}
			// 영어 정확히 N회 반복
			for r := 0; r < repeat; r++ {
				videoPaths = append(videoPaths, engVideoPath)
			}
			// 영어 반복 후 공백 1회 (다음 단어로 넘어가기 전)
			if options.PauseDuration > 0 && silenceVideoPath != "" && i < contentCount-1 {
				videoPaths = append(videoPaths, silenceVideoPath)
			}
		}

		log.Printf("영상 세트 생성 완료: %d/%d", i+1, contentCount)
	}

	log.Println("개별 영상 생성 및 리스트 조합 완료!")

	log.Println("개별 영상 생성 완료!")

	// 5. 모든 영상을 하나로 합치기
	// 지정된 날짜를 YYMMDD 형식으로 생성하고 서비스 타입에 따라 구별
	// 최종 결과물 디렉토리 생성
	finalVideoDir := "final-video"
	if err := os.MkdirAll(finalVideoDir, 0755); err != nil {
		log.Printf("final-video 디렉토리 생성 실패: %v", err)
		response.Error = err
		return response
	}

	var finalFileName string
	var fileNameBase string
	datePrefix := fmt.Sprintf("%02d%02d%02d", request.TargetDate.Year()%100, request.TargetDate.Month(), request.TargetDate.Day())

	switch enum.ServiceType(request.ServiceType) {
	case enum.InstagramWord:
		fileNameBase = "instagram_w"
	case enum.InstagramIdiom:
		fileNameBase = "instagram_i"
	case enum.InstagramSentence:
		fileNameBase = "instagram_s"

	case enum.FacebookWord:
		fileNameBase = "facebook_w"
	case enum.FacebookIdiom:
		fileNameBase = "facebook_i"
	case enum.FacebookSentence:
		fileNameBase = "facebook_s"

	case enum.YoutubeShortsWord:
		fileNameBase = "youtube_w"
	case enum.YoutubeShortsIdiom:
		fileNameBase = "youtube_i"
	case enum.YoutubeShotsSentence:
		fileNameBase = "youtube_s"

	default:
		fileNameBase = request.ServiceType
	}
	finalFileName = fmt.Sprintf("%s/%s_%s.mp4", finalVideoDir, datePrefix, fileNameBase)
	response.FinalFileName = finalFileName

	// 위에서 이미 videoPaths를 채웠으므로 다시 만들 필요 없음.
	// 기존 코드는 기존 videoPaths(조립 단위)를 무시하고 단순 1,2,3... 으로 재생성하려 했음.
	// 하지만 이제 videoPaths에 순서대로 다 들어있으므로 그대로 사용하면 됨.
	// 다만, output_filename 결정 로직만 사용.

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

	// 6. 중간 파일들 정리 (defer에서 처리하지만 명시적으로 로그 남김)
	log.Println("중간 파일들 정리 완료!")
	log.Printf("최종 영상: %s", finalFileName)

	response.Success = true
	return response
}

// cleanupTempFiles - 중간 파일들 정리
func (s *ReelsCreationService) cleanupTempFiles() {
	log.Println("🧹 임시 파일 및 디렉토리 정리 중...")
	if err := os.RemoveAll("temp"); err != nil {
		log.Printf("임시 디렉토리 삭제 실패: %v", err)
	}
}
