package service

import (
	"fmt"
	"os"
	"os/exec"
)

// AudioService 오디오 생성 서비스
type AudioService struct{}

// NewAudioService 새로운 오디오 서비스 생성
func NewAudioService() *AudioService {
	return &AudioService{}
}

// CreateKoreanAudioWithRate 한국어 텍스트로부터 지정된 속도의 음성을 생성합니다
func (s *AudioService) CreateKoreanAudioWithRate(
	text string,
	outputPath string,
	rate int,
) error {
	// 임시 aiff 파일 경로
	tempAiffPath := outputPath[:len(outputPath)-4] + ".aiff"

	// macOS의 say 명령어를 사용하여 aiff 음성 생성 (속도 조절)
	cmd := exec.Command("say",
		"-v", "Yuna",
		"-r", fmt.Sprintf("%d", rate),
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
		"-y",
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

// CreateNativeEnglishAudio 원어민 수준의 영어 음성을 생성합니다
func (s *AudioService) CreateNativeEnglishAudio(text, outputPath string) error {
	// Python 스크립트로 고품질 영어 음성 생성
	scriptContent := fmt.Sprintf(`#!/usr/bin/env python3
from gtts import gTTS
import os

def generate_native_english_audio(text, output_path):
    try:
        # 고품질 영어 음성 설정
        tts = gTTS(text=text, lang='en', tld='us', slow=False, lang_check=True)
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