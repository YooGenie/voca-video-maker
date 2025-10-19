package service

import (
	"auto-video-service/config"
	"os"
	"testing"
)

func TestSetTitleOnImage(t *testing.T) {
	// Setup: Initialize config
	config.InitConfig("../config/config.json")

	// Test
	service := NewImageService()
	title := "이번주 영어 단어 100개"
	imagePath := "../template/long.png"
	outpath := "../template/titleImage.png"

	err := service.SetTitleOnImage(title, imagePath, outpath)
	if err != nil {
		t.Fatalf("SetTitleOnImage failed: %v", err)
	}

	// Verify
	if _, err := os.Stat(outpath); os.IsNotExist(err) {
		t.Errorf("Expected output image '%s' to be created, but it was not", outpath)
	}
}