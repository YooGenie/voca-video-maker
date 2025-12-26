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
	title := "왕초보 영어단어 500개"
	subTitle := "Day 1"
	imagePath := "../template/title.png"
	outPath := "../template/titleImage.png"

	err := service.SetTitleOnImage(title, subTitle, imagePath, outPath)
	if err != nil {
		t.Fatalf("SetTitleOnImage failed: %v", err)
	}

	// Verify
	if _, err := os.Stat(outPath); os.IsNotExist(err) {
		t.Errorf("Expected output image '%s' to be created, but it was not", outPath)
	}
}
