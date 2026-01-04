package service

import (
	"auto-video-service/config"
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

// getProjectRoot returns the project root directory based on the current file's location
func getProjectRoot() string {
	_, currentFile, _, _ := runtime.Caller(0)
	return filepath.Dir(filepath.Dir(currentFile))
}

func TestSetTitleOnImage(t *testing.T) {
	projectRoot := getProjectRoot()

	// Save current directory and restore after test
	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}
	defer func() {
		_ = os.Chdir(originalDir)
	}()

	// Change to project root for relative path compatibility
	if err := os.Chdir(projectRoot); err != nil {
		t.Fatalf("Failed to change directory: %v", err)
	}

	// Setup: Initialize config
	config.InitConfig("config/config.json")

	// Test
	service := NewImageService()
	title := "왕초보 영어단어 500개"
	subTitle := "Day 1"
	imagePath := config.Config.Paths.Templates.Title
	outPath := "template/titleImage.png"

	err = service.SetTitleOnImage(title, subTitle, imagePath, outPath)
	if err != nil {
		t.Fatalf("SetTitleOnImage failed: %v", err)
	}

	// Verify
	if _, err := os.Stat(outPath); os.IsNotExist(err) {
		t.Errorf("Expected output image '%s' to be created, but it was not", outPath)
	}
}
