package service

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

// VideoService 비디오 생성 서비스
type VideoService struct {
	imageService *ImageService
}

// NewVideoService 새로운 비디오 서비스 생성
func NewVideoService(imageService *ImageService) *VideoService {
	return &VideoService{
		imageService: imageService,
	}
}

// CreateSilentVideo 이미지만으로 무음 영상을 생성합니다
func (s *VideoService) CreateSilentVideo(
	imagePath string,
	outputPath string,
	duration float64,
) error {
	// ffmpeg 명령어 구성
	// -loop 1: 이미지를 반복
	// -i imagePath: 입력 이미지
	// -c:v libx264: 비디오 코덱
	// -t duration: 지속 시간 설정
	cmd := exec.Command("ffmpeg",
		"-loop", "1",
		"-i", imagePath,
		"-c:v", "libx264",
		"-preset", "fast",
		"-profile:v", "baseline",
		"-level", "3.0",
		"-crf", "25",
		"-vf", "scale=1080:1920,fps=30",
		"-f", "lavfi",
		"-i", fmt.Sprintf("anullsrc=channel_layout=stereo:sample_rate=44100"),
		"-c:a", "aac",
		"-b:a", "128k",
		"-ar", "44100",
		"-t", fmt.Sprintf("%.2f", duration),
		"-y", // 기존 파일 덮어쓰기
		outputPath,
	)

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

// CreateVideoWithAudio 이미지와 음성을 합쳐서 영상을 생성합니다
func (s *VideoService) CreateVideoWithAudio(
	imagePath string,
	audioPath string,
	outputPath string,
	duration float64,
) error {
	// ffmpeg 명령어 구성
	// -loop 1: 이미지를 반복
	// -i imagePath: 입력 이미지
	// -i audioPath: 입력 오디오
	// -c:v libx264: 비디오 코덱
	// -c:a aac: 오디오 코덱
	// -shortest: 오디오 길이에 맞춰 비디오 종료
	// -t duration: 지속 시간 설정
	cmd := exec.Command("ffmpeg",
		"-loop", "1",
		"-i", imagePath,
		"-i", audioPath,
		"-c:v", "libx264",
		"-preset", "fast",
		"-profile:v", "baseline",
		"-level", "3.0",
		"-crf", "25",
		"-vf", "scale=1080:1920,fps=30",
		"-c:a", "aac",
		"-b:a", "128k",
		"-ar", "44100",
		"-shortest",
		"-avoid_negative_ts", "make_zero",
		"-fflags", "+genpts",
		"-movflags", "+faststart",
		"-t", fmt.Sprintf("%.2f", duration),
		"-y", // 기존 파일 덮어쓰기
		outputPath,
	)

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

// CreateVideoWithKoreanAndEnglish 한국어 한 번 + 0.5초 쉬고 + 영어 2번 + 0.5초 쉬는 영상을 생성합니다
func (s *VideoService) CreateVideoWithKoreanAndEnglish(
	imagePath string,
	koreanAudioPath string,
	englishAudioPath string,
	outputPath string,
) error {
	// 임시 오디오 파일들 생성
	tempKoreanPath := koreanAudioPath[:len(koreanAudioPath)-4] + "_temp.mp3"
	tempEnglishPath := englishAudioPath[:len(englishAudioPath)-4] + "_temp.mp3"

	// 한국어 오디오에 0.5초 무음 추가
	koreanCmd := exec.Command("ffmpeg",
		"-i", koreanAudioPath,
		"-af", "apad=pad_dur=0.5",
		"-y",
		tempKoreanPath,
	)

	koreanCmd.Stdout = os.Stdout
	koreanCmd.Stderr = os.Stderr

	if err := koreanCmd.Run(); err != nil {
		return fmt.Errorf("한국어 오디오 처리 실패: %v", err)
	}

	// 영어 오디오를 2번 반복
	englishCmd := exec.Command("ffmpeg",
		"-i", englishAudioPath,
		"-filter_complex", "[0:a]aloop=loop=-1:size=2e+09[a]",
		"-map", "[a]",
		"-y",
		tempEnglishPath,
	)

	englishCmd.Stdout = os.Stdout
	englishCmd.Stderr = os.Stderr

	if err := englishCmd.Run(); err != nil {
		return fmt.Errorf("영어 오디오 처리 실패: %v", err)
	}

	// 두 오디오를 연결
	concatPath := outputPath[:len(outputPath)-4] + "_concat.mp3"
	concatCmd := exec.Command("ffmpeg",
		"-i", tempKoreanPath,
		"-i", tempEnglishPath,
		"-filter_complex", "[0:a][1:a]concat=n=2:v=0:a=1[a]",
		"-map", "[a]",
		"-y",
		concatPath,
	)

	concatCmd.Stdout = os.Stdout
	concatCmd.Stderr = os.Stderr

	if err := concatCmd.Run(); err != nil {
		return fmt.Errorf("오디오 연결 실패: %v", err)
	}

	// 비디오 생성
	cmd := exec.Command("ffmpeg",
		"-loop", "1",
		"-i", imagePath,
		"-i", concatPath,
		"-c:v", "libx264",
		"-c:a", "aac",
		"-shortest",
		"-y",
		outputPath,
	)

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("비디오 생성 실패: %v", err)
	}

	// 임시 파일들 삭제
	os.Remove(tempKoreanPath)
	os.Remove(tempEnglishPath)
	os.Remove(concatPath)

	return nil
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
		"-vf", "scale=1080:1920,format=yuv420p,fps=30",
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
		"-vf", "scale=1080:1920,format=yuv420p,fps=30",
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

// GenerateVideosFromNumberedFiles 1부터 시작하는 번호가 매겨진 파일들로부터 영상을 생성합니다
func (s *VideoService) GenerateVideosFromNumberedFiles(
	imageDir string,
	audioDir string,
	outputDir string,
	imageFormat string,
	startNumber int,
	endNumber int,
) error {
	// 출력 디렉토리 생성
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("출력 디렉토리 생성 실패: %v", err)
	}

	for i := startNumber; i <= endNumber; i++ {
		imagePath := filepath.Join(imageDir, fmt.Sprintf("%d.%s", i, imageFormat))
		outputPath := filepath.Join(outputDir, fmt.Sprintf("%d.mp4", i))

		if i%2 == 1 { // 홀수 번호 - 한국어
			koreanAudioPath := filepath.Join(audioDir, fmt.Sprintf("%d_ko.mp3", i))
			if err := s.CreateVideoWithKorean(imagePath, koreanAudioPath, outputPath); err != nil {
				return fmt.Errorf("한국어 영상 생성 실패 (%d): %v", i, err)
			}
		} else { // 짝수 번호 - 영어
			englishAudioPath := filepath.Join(audioDir, fmt.Sprintf("%d_en.mp3", i))
			if err := s.CreateVideoWithEnglish(imagePath, englishAudioPath, outputPath); err != nil {
				return fmt.Errorf("영어 영상 생성 실패 (%d): %v", i, err)
			}
		}
	}

	return nil
}

// CreateVideoFromImages 여러 이미지로부터 개별 영상을 생성합니다
func (s *VideoService) CreateVideoFromImages(
	imagePrefix string,
	audioPrefix string,
	outputPrefix string,
	duration float64,
	count int,
) error {
	// videos 디렉토리 생성
	videosDir := "videos"
	if err := os.MkdirAll(videosDir, 0755); err != nil {
		return fmt.Errorf("videos 디렉토리 생성 실패: %v", err)
	}

	// 각 이미지에 대해 개별 영상 생성
	for i := 0; i < count; i++ {
		imagePath := fmt.Sprintf("%s_%d.png", imagePrefix, i)
		audioPath := fmt.Sprintf("%s_%d.mp3", audioPrefix, i)
		outputPath := filepath.Join(videosDir, fmt.Sprintf("%s_%d.mp4", outputPrefix, i))

		// 이미지 파일이 존재하는지 확인
		if _, err := os.Stat(imagePath); os.IsNotExist(err) {
			return fmt.Errorf("이미지 파일이 존재하지 않습니다: %s", imagePath)
		}

		// 오디오 파일이 존재하는지 확인
		if _, err := os.Stat(audioPath); os.IsNotExist(err) {
			return fmt.Errorf("오디오 파일이 존재하지 않습니다: %s", audioPath)
		}

		fmt.Printf("영상 생성 중: %d/%d (이미지: %s, 오디오: %s)\n", i+1, count, imagePath, audioPath)
		if err := s.CreateVideoWithAudio(imagePath, audioPath, outputPath, duration); err != nil {
			return fmt.Errorf("영상 생성 실패 (%d): %v", i, err)
		}
	}

	return nil
}

// ConcatenateVideos 여러 영상을 하나로 합칩니다
func (s *VideoService) ConcatenateVideos(
	videoPrefix string,
	outputPath string,
	count int,
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
	for i := 0; i < count; i++ {
		videoPath := fmt.Sprintf("%s_%d.mp4", videoPrefix, i)
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

// GenerateAudioFromText 텍스트로부터 음성을 생성합니다
func (s *VideoService) GenerateAudioFromText(
	text string,
	outputPath string,
) error {
	// 임시 aiff 파일 경로
	tempAiffPath := outputPath[:len(outputPath)-4] + ".aiff"

	// macOS의 say 명령어를 사용하여 aiff 음성 생성
	cmd := exec.Command("say",
		"-v", "Alex", // 영어 음성
		"-o", tempAiffPath,
		text,
	)

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("음성 생성 실패: %v", err)
	}

	// aiff를 mp3로 변환
	convertCmd := exec.Command("ffmpeg",
		"-i", tempAiffPath,
		"-acodec", "libmp3lame",
		"-ab", "128k",
		"-y", // 기존 파일 덮어쓰기
		outputPath,
	)

	convertCmd.Stdout = os.Stdout
	convertCmd.Stderr = os.Stderr

	if err := convertCmd.Run(); err != nil {
		return fmt.Errorf("mp3 변환 실패: %v", err)
	}

	// 임시 aiff 파일 삭제
	os.Remove(tempAiffPath)

	return nil
}

// GenerateKoreanAudioFromText 한국어 텍스트로부터 음성을 생성합니다
func (s *VideoService) GenerateKoreanAudioFromText(
	text string,
	outputPath string,
) error {
	// 임시 aiff 파일 경로
	tempAiffPath := outputPath[:len(outputPath)-4] + ".aiff"

	// macOS의 say 명령어를 사용하여 aiff 음성 생성
	cmd := exec.Command("say",
		"-v", "Yuna", // 한국어 음성 (Yuna는 한국어 음성)
		"-o", tempAiffPath,
		text,
	)

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("음성 생성 실패: %v", err)
	}

	// aiff를 mp3로 변환
	convertCmd := exec.Command("ffmpeg",
		"-i", tempAiffPath,
		"-acodec", "libmp3lame",
		"-ab", "128k",
		"-y", // 기존 파일 덮어쓰기
		outputPath,
	)

	convertCmd.Stdout = os.Stdout
	convertCmd.Stderr = os.Stderr

	if err := convertCmd.Run(); err != nil {
		return fmt.Errorf("mp3 변환 실패: %v", err)
	}

	// 임시 aiff 파일 삭제
	os.Remove(tempAiffPath)

	return nil
}

// GenerateNativeEnglishAudio 원어민 수준의 영어 음성을 생성합니다
func (s *VideoService) GenerateNativeEnglishAudio(text, outputPath string) error {
	// Python 스크립트로 고품질 영어 음성 생성
	scriptContent := fmt.Sprintf(`#!/usr/bin/env python3
from gtts import gTTS
import os

def generate_native_english_audio(text, output_path):
    try:
        # 고품질 영어 음성 설정
        tts = gTTS(text=text, lang='en', slow=False, lang_check=True)
        tts.save(output_path)
        print(f"✅ 원어민 영어 음성 생성 완료: {output_path}")
        return True
    except Exception as e:
        print(f"❌ 영어 음성 생성 실패: {e}")
        return False

# 영어 텍스트
text = "%s"
output_file = "%s"

generate_native_english_audio(text, output_file)
`, text, outputPath)

	// 임시 스크립트 파일 생성
	scriptFile := "temp_english_audio.py"
	err := os.WriteFile(scriptFile, []byte(scriptContent), 0644)
	if err != nil {
		return fmt.Errorf("영어 음성 스크립트 파일 생성 실패: %v", err)
	}
	defer os.Remove(scriptFile)

	// Python 스크립트 실행
	cmd := exec.Command("python3", scriptFile)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("영어 음성 생성 스크립트 실행 실패: %v, 출력: %s", err, string(output))
	}

	return nil
}

// GenerateAllNativeEnglishAudio 모든 영어 단어에 대해 원어민 음성을 생성합니다
func (s *VideoService) GenerateAllNativeEnglishAudio(englishWords []string, outputPrefix string) error {
	fmt.Println("🎤 원어민 영어 음성 파일들을 생성합니다...")

	for i, word := range englishWords {
		outputFile := fmt.Sprintf("%s_eng_%02d.mp3", outputPrefix, i+1)
		err := s.GenerateNativeEnglishAudio(word, outputFile)
		if err != nil {
			fmt.Printf("⚠️ 영어 음성 생성 실패 (%s): %v\n", word, err)
		}
	}

	fmt.Println("✅ 모든 영어 음성 파일 생성 완료!")
	return nil
}

// GenerateAzureEnglishAudio Azure Cognitive Services를 사용한 고품질 영어 음성을 생성합니다
func (s *VideoService) GenerateAzureEnglishAudio(text, outputPath string) error {
	// Azure Cognitive Services 사용 (API 키가 필요한 경우)
	scriptContent := fmt.Sprintf(`#!/usr/bin/env python3
import requests
import json
import os

def generate_azure_english_audio(text, output_path):
    try:
        # Azure Cognitive Services 설정
        subscription_key = "YOUR_AZURE_KEY"  # 실제 사용시 API 키 필요
        region = "eastus"
        
        # 음성 설정 (원어민 수준)
        voice_name = "en-US-JennyNeural"  # 자연스러운 여성 음성
        # voice_name = "en-US-GuyNeural"  # 자연스러운 남성 음성
        
        url = f"https://{region}.tts.speech.microsoft.com/cognitiveservices/v1"
        
        headers = {
            "Ocp-Apim-Subscription-Key": subscription_key,
            "Content-Type": "application/ssml+xml",
            "X-Microsoft-OutputFormat": "audio-16khz-128kbitrate-mono-mp3"
        }
        
        # SSML (Speech Synthesis Markup Language) 사용
        ssml = f'''<speak version="1.0" xmlns="http://www.w3.org/2001/10/synthesis" xml:lang="en-US">
            <voice name="{voice_name}">
                <prosody rate="medium" pitch="medium">
                    {text}
                </prosody>
            </voice>
        </speak>'''
        
        response = requests.post(url, headers=headers, data=ssml.encode('utf-8'))
        
        if response.status_code == 200:
            with open(output_path, 'wb') as f:
                f.write(response.content)
            print(f"✅ Azure 영어 음성 생성 완료: {output_path}")
            return True
        else:
            print(f"❌ Azure API 오류: {response.status_code}")
            return False
            
    except Exception as e:
        print(f"❌ Azure 음성 생성 실패: {e}")
        return False

# 영어 텍스트
text = "%s"
output_file = "%s"

generate_azure_english_audio(text, output_file)
`, text, outputPath)

	// 임시 스크립트 파일 생성
	scriptFile := "temp_azure_audio.py"
	err := os.WriteFile(scriptFile, []byte(scriptContent), 0644)
	if err != nil {
		return fmt.Errorf("Azure 음성 스크립트 파일 생성 실패: %v", err)
	}
	defer os.Remove(scriptFile)

	// Python 스크립트 실행
	cmd := exec.Command("python3", scriptFile)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("Azure 음성 생성 스크립트 실행 실패: %v, 출력: %s", err, string(output))
	}

	return nil
}
