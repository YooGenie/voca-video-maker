package main

import (
	"context"
	"log"
	"os"
	"time"

	"auto-video-service/config"
	"auto-video-service/enum"
	"auto-video-service/factory"
)

func main() {
	// YAML 설정 파일에서 설정 로드
	cliCfg, err := config.LoadCliConfig("config/config.yaml")
	if err != nil {
		log.Fatalf("설정 파일을 읽는 중 에러 발생: %v", err)
	}

	// 날짜 처리: 'today'이거나 비어있으면 오늘 날짜로 설정
	date := cliCfg.Video.Date
	if date == "today" || date == "" {
		date = time.Now().Format("20060102")
	}

	// 서비스 타입 가져오기
	serviceType := cliCfg.Video.Type

	// 서비스 타입 유효성 검사
	allowedTypes := map[string]bool{
		string(enum.InstagramWord): true, string(enum.InstagramIdiom): true, string(enum.InstagramSentence): true,
		string(enum.YoutubeLongform): true, string(enum.Start): true,
		string(enum.FacebookWord): true, string(enum.FacebookIdiom): true, string(enum.FacebookSentence): true,
		string(enum.YoutubeShortsWord): true, string(enum.YoutubeShortsIdiom): true, string(enum.YoutubeShotsSentence): true,
	}
	if serviceType == "" || !allowedTypes[serviceType] {
		log.Println("에러: config.yaml의 type 값이 올바르지 않습니다.")
		log.Printf("허용된 타입: iw, ii, is, fw, fi, fs, ysw, ysi, yss, yl, start")
		os.Exit(1)
	}

	// 날짜 형식 유효성 검사
	_, err = time.Parse("20060102", date)
	if err != nil {
		log.Printf("에러: 날짜 형식이 잘못되었습니다. YYYYMMDD 형식으로 입력해주세요. (입력값: %s)", date)
		os.Exit(1)
	}

	log.Printf("📹 영상 생성 시작: 타입=%s, 날짜=%s", serviceType, date)

	// 설정 파일 로드
	config.InitConfig("config/config.json")

	// 디비 연결
	config.ConfigureDatabase()

	ctx := context.Background()

	videoFactory := factory.NewVideoServiceFactory()
	videoFactory.CreateVideo(ctx, date, serviceType)
}
