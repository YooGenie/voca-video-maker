package main

import (
	"context"
	"flag"

	"auto-video-service/config"
	"auto-video-service/service"
)

func main() {
	// 명령행 인자 파싱
	var dateFlag = flag.String("date", "", "날짜 지정 (YYYYMMDD 형식, 예: 20250907)")
	var serviceType = flag.String("type", "W", "서비스 타입 (W)")
	flag.Parse()

	// 설정 파일 로드
	config.InitConfig("config/config.json")

	// 디비 연결
	config.ConfigureDatabase() //DB 설정

	ctx := context.Background()

	factory := service.NewVideoServiceFactory()
	factory.CreateVideo(ctx, dateFlag, serviceType)
}