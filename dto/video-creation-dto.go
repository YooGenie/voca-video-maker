package dto

import (
	"auto-video-service/enum"
	"time"
)

// VideoCreationRequest - 비디오 생성 요청 DTO
type VideoCreationRequest struct {
	TargetDate  time.Time
	ServiceType string
	ContentType enum.ContentType
}

// ContentData - 컨텐츠 데이터 DTO (영어 단어/숙어 공통)
type ContentData struct {
	Primary        []string // 영어 단어 또는 숙어
	PrimaryLine2   []string // 영어 두 번째 줄 (SS 타입 전용, english_sentence_2)
	Secondary      []string // 한국어 번역 또는 의미
	SecondaryLine2 []string // 한국어 두 번째 줄 (SS 타입 전용, korean_sentence_2)
	Tertiary       []string // 발음 또는 예문
	Count          int      // 컨텐츠 개수
	IsReverse      bool     // true면 영어->한국어, false면(기본) 한국어->영어
}

// TemplateConfig - 템플릿 설정 DTO
type TemplateConfig struct {
	BaseTemplate string
	TextColor    enum.TextColor
}

// VideoCreationResponse - 비디오 생성 결과 DTO
type VideoCreationResponse struct {
	FinalFileName string
	ContentCount  int
	Success       bool
	Error         error
}

// VideoCreationOptions - 비디오 생성 옵션 DTO (플랫폼, 길이, 반복 등)
type VideoCreationOptions struct {
	Platform           enum.Platform
	VideoLength        enum.VideoLength
	EnglishRepeatCount int               // 영어 반복 횟수 (예: 기본 1, 페이스북 3)
	SpeakSpeed         float64           // TTS/Video 속도 계수 (예: 기본 1.0, 느리게 0.8)
	PauseDuration      float64           // 문장 간 공백 (초 단위)
	LangOrder          string            // "kor_eng", "eng_kor" (Legacy, will be superseded by IsReverse logic if preferred, or used together)
	TemplateType       enum.TemplateType // individual(각각), common(공통)
}

// ContentDataResult - DB 조회 결과 DTO
type ContentDataResult struct {
	Primary        []string // 영어 (단어/숙어/문장1)
	PrimaryLine2   []string // 영어 2번째 줄 (문장 전용)
	Secondary      []string // 한국어
	SecondaryLine2 []string // 한국어 2번째 줄 (문장 전용)
	Tertiary       []string // 발음
}
