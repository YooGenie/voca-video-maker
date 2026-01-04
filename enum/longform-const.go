package enum

// =============================================================================
// 롱폼 이미지 관련 상수 (GenerateLongformImages)
// =============================================================================

const (
	// 롱폼 텍스트 영역
	LongformMaxTextWidthRatio = 0.8 // 가로형이므로 80%

	// 롱폼 폰트 크기
	LongformMaxFontSize  = 120.0 // 메인 텍스트 최대 폰트 크기
	LongformMinFontSize  = 20.0  // 메인 텍스트 최소 폰트 크기
	LongformFontSizeStep = 10.0  // 폰트 크기 감소 단위

	// 롱폼 위치 오프셋
	LongformYOffset = 100 // Y축 위로 이동

	// 롱폼 그림자/외곽선 (메인 텍스트)
	LongformShadowOffset  = 8 // 그림자 오프셋
	LongformOutlineOffset = 5 // 외곽선 굵기

	// 롱폼 발음 텍스트
	PronounceMaxFontSize   = 75.0 // 발음 최대 폰트 크기
	PronounceSpacing       = 20   // 메인 텍스트와 발음 간격
	PronounceShadowOffset  = 3    // 발음 그림자 오프셋
	PronounceOutlineOffset = 2    // 발음 외곽선 굵기
)
