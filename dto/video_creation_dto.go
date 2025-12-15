package dto

import "time"

// VideoCreationRequest - 비디오 생성 요청 DTO
type VideoCreationRequest struct {
	TargetDate  time.Time
	ServiceType string
}

// ContentData - 컨텐츠 데이터 DTO (영어 단어/숙어 공통)
type ContentData struct {
	Primary        []string // 영어 단어 또는 숙어
	PrimaryLine2   []string // 영어 두 번째 줄 (SS 타입 전용, english_sentence_2)
	Secondary      []string // 한국어 번역 또는 의미
	SecondaryLine2 []string // 한국어 두 번째 줄 (SS 타입 전용, korean_sentence_2)
	Tertiary       []string // 발음 또는 예문
	Count          int      // 컨텐츠 개수
}

// TemplateConfig - 템플릿 설정 DTO
type TemplateConfig struct {
	BaseTemplate  string // 기본 템플릿 경로 (word.png, idiom.png)
	CountTemplate string // 개수 표시 템플릿 경로 (wordCount, idiomCount)
	TextColor     string // 텍스트 색상 ("white" 또는 "black")
}

// VideoCreationResponse - 비디오 생성 결과 DTO
type VideoCreationResponse struct {
	FinalFileName string
	ContentCount  int
	Success       bool
	Error         error
}
