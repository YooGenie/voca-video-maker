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
	// YAML 설정 파일에서 기본값 로드
	cliCfg, err := config.LoadCliConfig("config/config.yaml")
	if err != nil {
		log.Fatalf("설정 파일을 읽는 중 에러 발생: %v", err)
	}

	// 날짜 기본값 처리: config.yaml에 'today'라고 되어있거나 비어있으면 오늘 날짜로 설정
	defaultDate := cliCfg.Video.Date
	if defaultDate == "today" || defaultDate == "" {
		defaultDate = time.Now().Format("20060102")
	}

	// 명령행 인자 정의 (기본값으로 YAML 설정값 사용)
	dateFlag := flag.String("date", defaultDate, "날짜 지정 (YYYYMMDD 형식). 미입력 시 config.yaml 또는 오늘 날짜로 자동 설정됩니다.")
	typeFlag := flag.String("type", cliCfg.Video.Type, "서비스 타입 (W, I, SS, EK, L, START 중 하나). 필수 입력입니다.")
	flag.Parse()

	// 타입 플래그 유효성 검사
	allowedTypes := map[string]bool{"W": true, "I": true, "SS": true, "EK": true, "L": true, "START": true}
	if *typeFlag == "" || !allowedTypes[*typeFlag] {
		log.Println("에러: -type 플래그는 W, I, SS, EK, L, START 중 하나를 필수로 입력해야 합니다.")
		os.Exit(1)
	}

	// 날짜 형식 유효성 검사
	date := *dateFlag
	_, err = time.Parse("20060102", date)
	if err != nil {
		log.Printf("에러: 날짜 형식이 잘못되었습니다. YYYYMMDD 형식으로 입력해주세요. (입력값: %s)", date)
		os.Exit(1)
	}
	log.Printf("정보: 최종 설정된 날짜는 %s 입니다.", date)
	

	// 설정 파일 로드
	config.InitConfig("config/config.json")

	// 디비 연결
	config.ConfigureDatabase() //DB 설정

	ctx := context.Background()

	videoFactory := factory.NewVideoServiceFactory()
	videoFactory.CreateVideo(ctx, &date, typeFlag)
}