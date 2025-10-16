package dto

import "time"

// VideoCreationRequest - 비디오 생성 요청 DTO
type VideoCreationRequest struct {
	TargetDate  time.Time
	ServiceType string
}

// ContentData - 컨텐츠 데이터 DTO (영어 단어/숙어 공통)
type ContentData struct {
	Primary     []string // 영어 단어 또는 숙어
	Secondary   []string // 한국어 번역 또는 의미
	Tertiary    []string // 발음 또는 예문
	Count       int      // 컨텐츠 개수
}

// TemplateConfig - 템플릿 설정 DTO
type TemplateConfig struct {
	BaseTemplate  string // 기본 템플릿 경로 (word.png, idiom.png)
	CountTemplate string // 개수 표시 템플릿 경로 (wordCount, idiomCount)
}

// VideoCreationResponse - 비디오 생성 결과 DTO
type VideoCreationResponse struct {
	FinalFileName string
	ContentCount  int
	Success       bool
	Error         error
}
