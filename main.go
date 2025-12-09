package main

import (
	"context"
	"flag"
	"log"
	"os"
	"time"

	"auto-video-service/config"
	"auto-video-service/factory"
)

func main() {
	// 명령행 인자 정의
	dateFlag := flag.String("date", "", "날짜 지정 (YYYYMMDD 형식). 미입력 시 오늘 날짜로 자동 설정됩니다.")
	typeFlag := flag.String("type", "", "서비스 타입 (W, I, SS, EK, L, START 중 하나). 필수 입력입니다.")
	flag.Parse()

	// 타입 플래그 유효성 검사
	allowedTypes := map[string]bool{"W": true, "I": true, "SS": true, "EK": true, "L": true, "START": true}
	if *typeFlag == "" || !allowedTypes[*typeFlag] {
		log.Println("에러: -type 플래그는 W, I, SS, EK, L, START 중 하나를 필수로 입력해야 합니다.")
		os.Exit(1)
	}

	// 날짜 플래그 처리
	date := *dateFlag
	if date == "" {
		date = time.Now().Format("20060102")
		log.Printf("정보: -date 플래그가 없어 오늘 날짜인 %s로 설정합니다.", date)
	} else {
		// 날짜 형식 유효성 검사
		_, err := time.Parse("20060102", date)
		if err != nil {
			log.Printf("에러: 날짜 형식이 잘못되었습니다. YYYYMMDD 형식으로 입력해주세요. (입력값: %s)", date)
			os.Exit(1)
		}
	}

	// 설정 파일 로드
	config.InitConfig("config/config.json")

	// 디비 연결
	config.ConfigureDatabase() //DB 설정

	ctx := context.Background()

	videoFactory := factory.NewVideoServiceFactory()
	videoFactory.CreateVideo(ctx, &date, typeFlag)
}