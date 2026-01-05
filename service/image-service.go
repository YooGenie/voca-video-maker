package service

import (
	"auto-video-service/config"
	"auto-video-service/enum"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"os"

	"golang.org/x/image/font"
	"golang.org/x/image/font/opentype"
	"golang.org/x/image/math/fixed"

	"github.com/disintegration/imaging"
)

// =============================================================================
// 타이틀 이미지 관련 상수 (SetTitleOnImage)
// =============================================================================
const (
	// 타이틀 영역 여백
	titleLeftMargin  = 830 // 왼쪽 여백 (그림 영역 피하기)
	titleRightMargin = 150 // 오른쪽 여백

	// 타이틀 폰트 크기
	titleMaxFontSize  = 120.0 // 최대 폰트 크기
	titleMinFontSize  = 10.0  // 최소 폰트 크기
	titleFontSizeStep = 5.0   // 폰트 크기 감소 단위

	// 타이틀 위치 오프셋
	titleYOffset = 150 // Y축 위로 이동

	// 타이틀 그림자/외곽선
	titleShadowOffset  = 8   // 그림자 오프셋
	titleOutlineOffset = 5   // 외곽선 굵기
	titleBlurSigma     = 4.0 // 그림자 블러 강도

	// 서브타이틀 관련
	subtitleFontRatio  = 0.9  // 타이틀 대비 서브타이틀 폰트 비율
	subtitleMaxFont    = 70.0 // 서브타이틀 최대 폰트 크기
	subtitleMinFont    = 25.0 // 서브타이틀 최소 폰트 크기
	subtitleSpacing    = 30   // 타이틀과 서브타이틀 간격
	subtitleShadowOff  = 8    // 서브타이틀 그림자 오프셋
	subtitleOutlineOff = 5    // 서브타이틀 외곽선 굵기
	subtitleBlurSigma  = 3.0  // 서브타이틀 블러 강도
)

// ImageService 이미지 생성 서비스
type ImageService struct{}

// NewImageService 새로운 이미지 서비스 생성
func NewImageService() *ImageService {
	return &ImageService{}
}

// TextRenderOptions 텍스트 렌더링 옵션
type TextRenderOptions struct {
	ShadowOffset  int
	OutlineOffset int
	MainColor     color.RGBA
	OutlineColor  color.RGBA
	ShadowColor   color.RGBA
}

// drawTextWithShadowAndOutline 그림자, 외곽선, 메인 텍스트를 한번에 그리는 헬퍼 함수
func drawTextWithShadowAndOutline(
	dst *image.RGBA,
	face font.Face,
	text string,
	pointX, pointY int,
	opts TextRenderOptions,
) {
	// 1. 그림자 그리기
	shadowDrawer := &font.Drawer{
		Dst:  dst,
		Src:  image.NewUniform(opts.ShadowColor),
		Face: face,
		Dot:  fixed.Point26_6{X: fixed.I(pointX + opts.ShadowOffset), Y: fixed.I(pointY + opts.ShadowOffset)},
	}
	shadowDrawer.DrawString(text)

	// 2. 외곽선 그리기 (8방향)
	offsets := []struct{ dx, dy int }{
		{-opts.OutlineOffset, -opts.OutlineOffset}, {0, -opts.OutlineOffset}, {opts.OutlineOffset, -opts.OutlineOffset},
		{-opts.OutlineOffset, 0}, {opts.OutlineOffset, 0},
		{-opts.OutlineOffset, opts.OutlineOffset}, {0, opts.OutlineOffset}, {opts.OutlineOffset, opts.OutlineOffset},
	}
	for _, off := range offsets {
		outlineDrawer := &font.Drawer{
			Dst:  dst,
			Src:  image.NewUniform(opts.OutlineColor),
			Face: face,
			Dot:  fixed.Point26_6{X: fixed.I(pointX + off.dx), Y: fixed.I(pointY + off.dy)},
		}
		outlineDrawer.DrawString(text)
	}

	// 3. 메인 텍스트 그리기
	mainDrawer := &font.Drawer{
		Dst:  dst,
		Src:  image.NewUniform(opts.MainColor),
		Face: face,
		Dot:  fixed.Point26_6{X: fixed.I(pointX), Y: fixed.I(pointY)},
	}
	mainDrawer.DrawString(text)
}

// GenerateBasicImages 단어 학습용 이미지들을 생성합니다
func (s *ImageService) GenerateBasicImages(
	imagePath string,
	eng []string,
	kor []string,
	pronounce []string,
	outputPrefix string,
	count int,
) error {
	return s.GenerateBasicImagesWithFontSize(imagePath, eng, []string{}, kor, []string{}, pronounce, outputPrefix, count, 120, enum.TextColorBeige)
}

// GenerateBasicImagesWithFontSize 단어 학습용 이미지들을 폰트 크기를 지정하여 생성합니다
func (s *ImageService) GenerateBasicImagesWithFontSize(
	imagePath string,
	eng []string,
	engLine2 []string, // 영어 두 번째 줄 (SS 타입 전용)
	kor []string,
	korLine2 []string, // 한국어 두 번째 줄 (SS 타입 전용)
	pronounce []string,
	outputPrefix string,
	count int,
	fontSize float64, // This will be treated as the maximum font size
	textColorEnum enum.TextColor, // 텍스트 색상 (white 또는 black)
) error {
	// 1. 이미지 불러오기
	existingImageFile, err := os.Open(imagePath)
	if err != nil {
		return fmt.Errorf("이미지 파일을 열 수 없습니다: %v", err)
	}
	defer existingImageFile.Close()

	img, err := png.Decode(existingImageFile)
	if err != nil {
		return fmt.Errorf("이미지 디코딩 실패: %v", err)
	}

	// 2. 폰트 불러오기
	fontBytes, err := os.ReadFile(config.Config.FontPath)
	if err != nil {
		return fmt.Errorf("폰트 파일을 읽을 수 없습니다: %v", err)
	}
	parsedFont, err := opentype.Parse(fontBytes)
	if err != nil {
		return fmt.Errorf("폰트 파싱 실패: %v", err)
	}

	// 3. 배열 길이 검증
	if len(eng) == 0 || len(kor) == 0 || len(pronounce) == 0 {
		return fmt.Errorf("입력 배열이 비어있습니다: eng=%d, kor=%d, pronounce=%d", len(eng), len(kor), len(pronounce))
	}

	expectedLength := count / 2
	if len(eng) < expectedLength || len(kor) < expectedLength || len(pronounce) < expectedLength {
		return fmt.Errorf("배열 길이가 부족합니다: 필요=%d, eng=%d, kor=%d, pronounce=%d",
			expectedLength, len(eng), len(kor), len(pronounce))
	}

	// 4. 텍스트 색상 결정
	var textColor color.RGBA

	switch textColorEnum {
	case enum.TextColorBlack:
		textColor = color.RGBA{R: 0, G: 0, B: 0, A: 255} // 검정색
	case enum.TextColorBeige:
		textColor = color.RGBA{R: 245, G: 245, B: 220, A: 255} // 베이지색 (#F5F5DC)
	default:
		textColor = color.RGBA{R: 255, G: 255, B: 255, A: 255} // 흰색 (기본값)
	}

	// 5. 이미지들 생성
	for i := 0; i < count; i++ {
		// 원본 이미지 복사
		rgba := image.NewRGBA(img.Bounds())
		draw.Draw(rgba, rgba.Bounds(), img, image.Point{}, draw.Src)

		var text string
		var secondText string
		var thirdText string

		if i%2 == 0 { // 짝수 번째 (0, 2, 4, ...) - 한국어
			text = kor[i/2]
			// SS 타입: korLine2가 있으면 두 번째 줄로 표시
			if len(korLine2) > i/2 && korLine2[i/2] != "" {
				secondText = korLine2[i/2]
			}
		} else { // 홀수 번째 (1, 3, 5, ...) - 영어
			text = eng[i/2]
			// SS 타입: engLine2가 있으면 두 번째 줄로 표시
			if len(engLine2) > i/2 && engLine2[i/2] != "" {
				secondText = engLine2[i/2]
			}
			// 발음은 항상 세 번째 줄
			thirdText = "( " + pronounce[i/2] + " )"
		}

		// ===== 글자 길이에 따른 동적 폰트 크기 조절 로직 시작 =====
		var face font.Face
		currentFontSize := fontSize // 제공된 폰트 크기를 최대 크기로 시작
		imgWidth := rgba.Bounds().Dx()
		imgHeight := rgba.Bounds().Dy()

		// 비디오 방향(가로/세로)에 따라 최대 텍스트 너비 조정
		var maxTextWidth int
		if imgWidth > imgHeight { // 가로형 비디오
			maxTextWidth = int(float64(imgWidth) * 0.8) // 너비의 80%
		} else { // 세로형 비디오
			maxTextWidth = int(float64(imgWidth) * 0.9) // 너비의 90%
		}

		for {
			var faceErr error
			face, faceErr = opentype.NewFace(parsedFont, &opentype.FaceOptions{
				Size:    currentFontSize,
				DPI:     72,
				Hinting: font.HintingFull,
			})
			if faceErr != nil {
				return fmt.Errorf("폰트 페이스 생성 실패: %v", faceErr)
			}

			textBounds, _ := font.BoundString(face, text)
			textWidth := (textBounds.Max.X - textBounds.Min.X).Ceil()

			if textWidth > maxTextWidth {
				face.Close() // 중요: 더 이상 사용하지 않을 face는 닫아줍니다.
				currentFontSize -= 10
				if currentFontSize < 20 { // 최소 폰트 크기 제한
					break
				}
				continue
			}
			break // 텍스트가 너비에 맞으면 루프 종료
		}
		// ===== 동적 폰트 크기 조절 로직 끝 =====

		// 텍스트 위치 계산
		textBounds, _ := font.BoundString(face, text)
		textWidth := (textBounds.Max.X - textBounds.Min.X).Ceil()
		textHeight := (textBounds.Max.Y - textBounds.Min.Y).Ceil()

		pointX := (imgWidth - textWidth) / 2
		// 비디오 방향에 따라 Y 오프셋 조정 (가로형: 아래쪽, 세로형: 위쪽)
		var yOffset int
		if imgWidth > imgHeight { // 가로형 비디오
			yOffset = -100
		} else { // 세로형 비디오
			yOffset = -180
		}
		pointY := (imgHeight+textHeight)/2 + yOffset

		// 이미지에 텍스트 그리기
		d := &font.Drawer{
			Dst:  rgba,
			Src:  image.NewUniform(textColor),
			Face: face,
			Dot:  fixed.Point26_6{X: fixed.I(pointX), Y: fixed.I(pointY)},
		}
		d.DrawString(text)
		face.Close() // 텍스트를 그린 후 face를 닫아줍니다.

		// 두 번째 텍스트가 있으면 아래에 그리기
		if secondText != "" {
			var secondFace font.Face
			var secondFontSize float64 = fontSize // 최대 폰트 크기로 시작

			// 두 번째 텍스트도 동적 폰트 크기 조절
			for {
				var faceErr error
				secondFace, faceErr = opentype.NewFace(parsedFont, &opentype.FaceOptions{
					Size:    secondFontSize,
					DPI:     72,
					Hinting: font.HintingFull,
				})
				if faceErr != nil {
					return fmt.Errorf("두 번째 줄 폰트 페이스 생성 실패: %v", faceErr)
				}

				secondTextBounds, _ := font.BoundString(secondFace, secondText)
				secondTextWidth := (secondTextBounds.Max.X - secondTextBounds.Min.X).Ceil()

				if secondTextWidth > maxTextWidth {
					secondFace.Close()
					secondFontSize -= 10
					if secondFontSize < 20 {
						break
					}
					continue
				}
				break
			}

			secondDrawer := &font.Drawer{
				Dst:  rgba,
				Src:  image.NewUniform(textColor),
				Face: secondFace,
			}

			var secondTextWidth, secondTextHeight int
			secondTextBounds, _ := font.BoundString(secondFace, secondText)
			secondTextWidth = (secondTextBounds.Max.X - secondTextBounds.Min.X).Ceil()
			secondTextHeight = (secondTextBounds.Max.Y - secondTextBounds.Min.Y).Ceil()

			secondPointX := (imgWidth - secondTextWidth) / 2
			secondPointY := pointY + textHeight + 20 // 첫 번째 텍스트 아래 20픽셀 간격

			secondDrawer.Dot = fixed.Point26_6{X: fixed.I(secondPointX), Y: fixed.I(secondPointY)}
			secondDrawer.DrawString(secondText)

			// 세 번째 텍스트(발음)가 있으면 아래에 그리기 - 동적 폰트 크기 조절
			if thirdText != "" {
				var thirdFace font.Face
				thirdFontSize := 75.0 // 최대 폰트 크기

				for {
					var faceErr error
					thirdFace, faceErr = opentype.NewFace(parsedFont, &opentype.FaceOptions{
						Size:    thirdFontSize,
						DPI:     72,
						Hinting: font.HintingFull,
					})
					if faceErr != nil {
						secondFace.Close()
						return fmt.Errorf("세 번째 줄 폰트 페이스 생성 실패: %v", faceErr)
					}

					thirdTextBounds, _ := font.BoundString(thirdFace, thirdText)
					thirdTextWidth := (thirdTextBounds.Max.X - thirdTextBounds.Min.X).Ceil()

					if thirdTextWidth > maxTextWidth {
						thirdFace.Close()
						thirdFontSize -= 10
						if thirdFontSize < 20 {
							break
						}
						continue
					}
					break
				}

				thirdTextBounds, _ := font.BoundString(thirdFace, thirdText)
				thirdTextWidth := (thirdTextBounds.Max.X - thirdTextBounds.Min.X).Ceil()

				thirdDrawer := &font.Drawer{
					Dst:  rgba,
					Src:  image.NewUniform(textColor),
					Face: thirdFace,
				}

				thirdPointX := (imgWidth - thirdTextWidth) / 2
				thirdPointY := secondPointY + secondTextHeight + 20 // 두 번째 텍스트 아래 20픽셀 간격

				thirdDrawer.Dot = fixed.Point26_6{X: fixed.I(thirdPointX), Y: fixed.I(thirdPointY)}
				thirdDrawer.DrawString(thirdText)
				thirdFace.Close()
			}

			secondFace.Close()
		} else if thirdText != "" {
			// 두 번째 텍스트가 없고 세 번째 텍스트(발음)만 있는 경우
			// 발음을 두 번째 줄 위치에 표시 - 동적 폰트 크기 조절
			var thirdFace font.Face
			thirdFontSize := 75.0 // 최대 폰트 크기

			for {
				var faceErr error
				thirdFace, faceErr = opentype.NewFace(parsedFont, &opentype.FaceOptions{
					Size:    thirdFontSize,
					DPI:     72,
					Hinting: font.HintingFull,
				})
				if faceErr != nil {
					return fmt.Errorf("발음 폰트 페이스 생성 실패: %v", faceErr)
				}

				thirdTextBounds, _ := font.BoundString(thirdFace, thirdText)
				thirdTextWidth := (thirdTextBounds.Max.X - thirdTextBounds.Min.X).Ceil()

				if thirdTextWidth > maxTextWidth {
					thirdFace.Close()
					thirdFontSize -= 10
					if thirdFontSize < 20 {
						break
					}
					continue
				}
				break
			}

			thirdTextBounds, _ := font.BoundString(thirdFace, thirdText)
			thirdTextWidth := (thirdTextBounds.Max.X - thirdTextBounds.Min.X).Ceil()

			thirdDrawer := &font.Drawer{
				Dst:  rgba,
				Src:  image.NewUniform(textColor),
				Face: thirdFace,
			}

			thirdPointX := (imgWidth - thirdTextWidth) / 2
			thirdPointY := pointY + textHeight + 20 // 첫 번째 텍스트 아래 20픽셀 간격

			thirdDrawer.Dot = fixed.Point26_6{X: fixed.I(thirdPointX), Y: fixed.I(thirdPointY)}
			thirdDrawer.DrawString(thirdText)
			thirdFace.Close()
		}

		// 이미지 저장
		outputFileName := fmt.Sprintf("%s_%02d.png", outputPrefix, i+1)
		outputFile, err := os.Create(outputFileName)
		if err != nil {
			return fmt.Errorf("출력 파일을 생성할 수 없습니다: %v", err)
		}

		err = png.Encode(outputFile, rgba)
		if err != nil {
			outputFile.Close()
			return fmt.Errorf("이미지 인코딩 실패: %v", err)
		}
		outputFile.Close()

		fmt.Printf("이미지 %d 생성 완료: %s\n", i+1, outputFileName)
	}

	fmt.Printf("모든 %d장의 이미지가 성공적으로 생성되었습니다.\n", count)
	return nil
}

// GenerateEKImages 단어 학습용 이미지들을 영어 -> 한국어 순서로 생성합니다
func (s *ImageService) GenerateEKImages(
	imagePath string,
	eng []string,
	kor []string,
	pronounce []string,
	outputPrefix string,
	count int,
) error {
	return s.GenerateEKImagesWithFontSize(imagePath, eng, kor, pronounce, outputPrefix, count, 120)
}

// GenerateEKImagesWithFontSize 단어 학습용 이미지들을 영어 -> 한국어 순서로, 폰트 크기를 지정하여 생성합니다
func (s *ImageService) GenerateEKImagesWithFontSize(
	imagePath string,
	eng []string,
	kor []string,
	pronounce []string,
	outputPrefix string,
	count int,
	fontSize float64,
) error {
	// 1. 이미지 불러오기
	existingImageFile, err := os.Open(imagePath)
	if err != nil {
		return fmt.Errorf("이미지 파일을 열 수 없습니다: %v", err)
	}
	defer existingImageFile.Close()

	// PNG 이미지 디코딩
	img, err := png.Decode(existingImageFile)
	if err != nil {
		return fmt.Errorf("이미지 디코딩 실패: %v", err)
	}

	// 2. 폰트 불러오기
	fontBytes, err := os.ReadFile(config.Config.FontPath)
	if err != nil {
		return fmt.Errorf("폰트 파일을 읽을 수 없습니다: %v", err)
	}

	parsedFont, err := opentype.Parse(fontBytes)
	if err != nil {
		return fmt.Errorf("폰트 파싱 실패: %v", err)
	}

	// 폰트 옵션 설정
	face, err := opentype.NewFace(parsedFont, &opentype.FaceOptions{
		Size:    fontSize, // 폰트 크기 (변수로 처리)
		DPI:     72,       // DPI (Dots Per Inch)
		Hinting: font.HintingNone,
	})
	if err != nil {
		return fmt.Errorf("폰트 페이스 생성 실패: %v", err)
	}
	defer face.Close()

	// 3. 배열 길이 검증
	if len(eng) == 0 || len(kor) == 0 {
		return fmt.Errorf("입력 배열이 비어있습니다: eng=%d, kor=%d", len(eng), len(kor))
	}

	expectedLength := count / 2
	if len(eng) < expectedLength || len(kor) < expectedLength {
		return fmt.Errorf("배열 길이가 부족합니다: 필요=%d, eng=%d, kor=%d",
			expectedLength, len(eng), len(kor))
	}

	// 4. 이미지들 생성
	textColor := color.RGBA{R: 245, G: 245, B: 220, A: 255} // 베이지색 (#F5F5DC)
	for i := 0; i < count; i++ {
		// 원본 이미지 복사
		rgba := image.NewRGBA(img.Bounds())
		draw.Draw(rgba, rgba.Bounds(), img, image.Point{}, draw.Src)

		var text string
		var secondText string

		isFirstImageOfPair := i%2 == 0

		// EK 타입은 항상 영어 -> 한국어 순서
		if isFirstImageOfPair {
			text = eng[i/2]
			if pronounce != nil && len(pronounce) > i/2 {
				secondText = "( " + pronounce[i/2] + " )"
			}
		} else {
			text = kor[i/2]
		}

		// 텍스트 위치 계산
		textBounds, _ := font.BoundString(face, text)
		textWidth := (textBounds.Max.X - textBounds.Min.X).Ceil()
		textHeight := (textBounds.Max.Y - textBounds.Min.Y).Ceil()

		imgWidth := rgba.Bounds().Dx()
		imgHeight := rgba.Bounds().Dy()

		pointX := (imgWidth - textWidth) / 2
		pointY := (imgHeight+textHeight)/2 - 180

		point := fixed.Point26_6{
			X: fixed.I(pointX),
			Y: fixed.I(pointY),
		}

		// 이미지에 텍스트 그리기
		d := &font.Drawer{
			Dst:  rgba,
			Src:  image.NewUniform(textColor),
			Face: face,
			Dot:  point,
		}
		d.DrawString(text)

		// 두 번째 텍스트(발음)가 있으면 아래에 그리기 - 동적 폰트 크기 조절
		if secondText != "" {
			var smallFace font.Face
			smallFontSize := 75.0 // 최대 폰트 크기

			// 비디오 방향에 따라 최대 텍스트 너비 조정
			var maxTextWidth int
			if imgWidth > imgHeight {
				maxTextWidth = int(float64(imgWidth) * 0.8)
			} else {
				maxTextWidth = int(float64(imgWidth) * 0.9)
			}

			for {
				var faceErr error
				smallFace, faceErr = opentype.NewFace(parsedFont, &opentype.FaceOptions{
					Size:    smallFontSize,
					DPI:     72,
					Hinting: font.HintingNone,
				})
				if faceErr != nil {
					return fmt.Errorf("발음 폰트 페이스 생성 실패: %v", faceErr)
				}

				secondTextBounds, _ := font.BoundString(smallFace, secondText)
				secondTextWidth := (secondTextBounds.Max.X - secondTextBounds.Min.X).Ceil()

				if secondTextWidth > maxTextWidth {
					smallFace.Close()
					smallFontSize -= 10
					if smallFontSize < 20 {
						break
					}
					continue
				}
				break
			}

			secondTextBounds, _ := font.BoundString(smallFace, secondText)
			secondTextWidth := (secondTextBounds.Max.X - secondTextBounds.Min.X).Ceil()

			smallDrawer := &font.Drawer{
				Dst:  rgba,
				Src:  image.NewUniform(textColor),
				Face: smallFace,
			}

			secondPointX := (imgWidth - secondTextWidth) / 2
			secondPointY := pointY + textHeight + 20 // 첫 번째 텍스트 아래 20픽셀 간격

			smallDrawer.Dot = fixed.Point26_6{X: fixed.I(secondPointX), Y: fixed.I(secondPointY)}
			smallDrawer.DrawString(secondText)
			smallFace.Close()
		}

		// 이미지 저장
		outputFileName := fmt.Sprintf("%s_%02d.png", outputPrefix, i+1)
		outputFile, err := os.Create(outputFileName)
		if err != nil {
			return fmt.Errorf("출력 파일을 생성할 수 없습니다: %v", err)
		}

		err = png.Encode(outputFile, rgba)
		if err != nil {
			outputFile.Close()
			return fmt.Errorf("이미지 인코딩 실패: %v", err)
		}
		outputFile.Close()

		fmt.Printf("이미지 %d 생성 완료: %s\n", i+1, outputFileName)
	}

	fmt.Printf("모든 %d장의 이미지가 성공적으로 생성되었습니다.\n", count)
	return nil
}

// SetWordCountOnImage wordCount 값을 이미지에 표시하는 이미지를 생성합니다
func (s *ImageService) SetWordCountOnImage(
	imagePath string,
	wordCountText string,
	outputPrefix string,
	contentType enum.ContentType,
) error {
	// 1. 이미지 불러오기
	existingImageFile, err := os.Open(imagePath)
	if err != nil {
		return fmt.Errorf("이미지 파일을 열 수 없습니다: %v", err)
	}
	defer existingImageFile.Close()

	// PNG 이미지 디코딩
	img, err := png.Decode(existingImageFile)
	if err != nil {
		return fmt.Errorf("이미지 디코딩 실패: %v", err)
	}

	// 2. 폰트 불러오기
	fontBytes, err := os.ReadFile(config.Config.FontPath)
	if err != nil {
		return fmt.Errorf("폰트 파일을 읽을 수 없습니다: %v", err)
	}

	parsedFont, err := opentype.Parse(fontBytes)
	if err != nil {
		return fmt.Errorf("폰트 파싱 실패: %v", err)
	}

	// 폰트 옵션 설정 (wordCount용으로 크기 조정)
	face, err := opentype.NewFace(parsedFont, &opentype.FaceOptions{
		Size:    90, // wordCount용 폰트 크기 (80에서 100으로 증가)
		DPI:     72,
		Hinting: font.HintingNone,
	})
	if err != nil {
		return fmt.Errorf("폰트 페이스 생성 실패: %v", err)
	}
	defer face.Close()

	// 3. wordCount 이미지 생성
	rgba := image.NewRGBA(img.Bounds())
	draw.Draw(rgba, rgba.Bounds(), img, image.Point{}, draw.Src)

	// wordCount 텍스트 설정
	text := wordCountText
	var textColor color.RGBA

	// 서비스 타입에 따라 글자색 설정
	switch contentType {
	case enum.ContentWord:
		textColor = color.RGBA{R: 173, G: 216, B: 230, A: 255} // #ADD8E6 (연한 파란색)
	case enum.ContentIdiom:
		textColor = color.RGBA{R: 248, G: 202, B: 204, A: 255} // #F8CACC (연한 분홍색)
	default:
		textColor = color.RGBA{R: 173, G: 216, B: 230, A: 255} // 기본값: #ADD8E6
	}

	// 텍스트 위치 계산
	textBounds, _ := font.BoundString(face, text)
	textWidth := (textBounds.Max.X - textBounds.Min.X).Ceil()
	textHeight := (textBounds.Max.Y - textBounds.Min.Y).Ceil()

	imgWidth := rgba.Bounds().Dx()

	// 이미지 우측 상단에 배치
	pointX := imgWidth - textWidth - 720
	pointY := textHeight + 405

	point := fixed.Point26_6{
		X: fixed.Int26_6(pointX * 64),
		Y: fixed.Int26_6(pointY * 64),
	}

	// 텍스트 그리기
	d := &font.Drawer{
		Dst:  rgba,
		Src:  image.NewUniform(textColor),
		Face: face,
		Dot:  point,
	}
	d.DrawString(text)

	// 4. 이미지 저장
	outputPath := fmt.Sprintf("%s.png", outputPrefix)
	outputFile, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("출력 파일 생성 실패: %v", err)
	}
	defer outputFile.Close()

	if err := png.Encode(outputFile, rgba); err != nil {
		return fmt.Errorf("이미지 인코딩 실패: %v", err)
	}

	return nil
}

// SetTitleOnImage creates an image with a centered title and subtitle.
func (s *ImageService) SetTitleOnImage(title, subTitle, imagePath, outputPath string) error {
	// 1. Load image
	existingImageFile, err := os.Open(imagePath)
	if err != nil {
		return fmt.Errorf("could not open image file: %v", err)
	}
	defer existingImageFile.Close()

	// Decode PNG image
	img, err := png.Decode(existingImageFile)
	if err != nil {
		return fmt.Errorf("failed to decode image: %v", err)
	}

	// 2. Load title font
	fontBytes, err := os.ReadFile(config.Config.TitleFontPath)
	if err != nil {
		return fmt.Errorf("could not read font file: %v", err)
	}

	parsedFont, err := opentype.Parse(fontBytes)
	if err != nil {
		return fmt.Errorf("failed to parse font: %v", err)
	}

	// 3. Create image
	rgba := image.NewRGBA(img.Bounds())
	draw.Draw(rgba, rgba.Bounds(), img, image.Point{}, draw.Src)

	imgWidth := rgba.Bounds().Dx()
	imgHeight := rgba.Bounds().Dy()

	// 텍스트 영역 정의 (왼쪽 그림 영역 피하기, 양쪽 여백 확보)
	rightAreaEnd := imgWidth - titleRightMargin    // 텍스트 영역 끝점
	maxTextWidth := rightAreaEnd - titleLeftMargin // 동적으로 텍스트 영역 너비 계산

	// 색상 정의
	mainColor := color.RGBA{R: 0x8F, G: 0x5B, B: 0x34, A: 255} // #8F5B34 갈색
	outlineColor := color.RGBA{R: 255, G: 255, B: 255, A: 255} // 흰색 외곽선

	// === 타이틀 동적 폰트 크기 조절 ===
	var titleFace font.Face
	titleFontSize := titleMaxFontSize

	for {
		var faceErr error
		titleFace, faceErr = opentype.NewFace(parsedFont, &opentype.FaceOptions{
			Size:    titleFontSize,
			DPI:     72,
			Hinting: font.HintingFull,
		})
		if faceErr != nil {
			return fmt.Errorf("failed to create font face: %v", faceErr)
		}

		textBounds, _ := font.BoundString(titleFace, title)
		textWidth := (textBounds.Max.X - textBounds.Min.X).Ceil()

		if textWidth > maxTextWidth {
			titleFace.Close()
			titleFontSize -= titleFontSizeStep
			if titleFontSize < titleMinFontSize {
				break
			}
			continue
		}
		break
	}
	defer titleFace.Close()

	// Calculate title text position
	textBounds, _ := font.BoundString(titleFace, title)
	textWidth := (textBounds.Max.X - textBounds.Min.X).Ceil()
	textHeight := (textBounds.Max.Y - textBounds.Min.Y).Ceil()

	// 타이틀 왼쪽 정렬
	pointX := titleLeftMargin
	// 오른쪽 경계를 넘지 않도록 제한
	if pointX+textWidth > rightAreaEnd {
		pointX = rightAreaEnd - textWidth
	}
	pointY := (imgHeight+textHeight)/2 - titleYOffset

	// === 그림자 그리기 (블러 적용) ===
	shadowLayer := image.NewRGBA(rgba.Bounds())
	shadowDrawer := &font.Drawer{
		Dst:  shadowLayer,
		Src:  image.NewUniform(color.RGBA{R: 0, G: 0, B: 0, A: 180}),
		Face: titleFace,
		Dot:  fixed.Point26_6{X: fixed.I(pointX + titleShadowOffset), Y: fixed.I(pointY + titleShadowOffset)},
	}
	shadowDrawer.DrawString(title)

	// 블러 처리
	blurredShadow := imaging.Blur(shadowLayer, titleBlurSigma)

	// 원본에 합성
	draw.Draw(rgba, rgba.Bounds(), blurredShadow, image.Point{}, draw.Over)

	// === 외곽선 그리기 (8방향) ===
	offsets := []struct{ dx, dy int }{
		{-titleOutlineOffset, -titleOutlineOffset}, {0, -titleOutlineOffset}, {titleOutlineOffset, -titleOutlineOffset},
		{-titleOutlineOffset, 0}, {titleOutlineOffset, 0},
		{-titleOutlineOffset, titleOutlineOffset}, {0, titleOutlineOffset}, {titleOutlineOffset, titleOutlineOffset},
	}
	for _, off := range offsets {
		outlineDrawer := &font.Drawer{
			Dst:  rgba,
			Src:  image.NewUniform(outlineColor),
			Face: titleFace,
			Dot:  fixed.Point26_6{X: fixed.I(pointX + off.dx), Y: fixed.I(pointY + off.dy)},
		}
		outlineDrawer.DrawString(title)
	}

	// === 메인 타이틀 그리기 ===
	mainDrawer := &font.Drawer{
		Dst:  rgba,
		Src:  image.NewUniform(mainColor),
		Face: titleFace,
		Dot:  fixed.Point26_6{X: fixed.I(pointX), Y: fixed.I(pointY)},
	}
	mainDrawer.DrawString(title)

	// Draw subtitle below title if provided
	if subTitle != "" {
		// 서브타이틀용 폰트 로딩 (NanumGothicExtraBold)
		subFontBytes, err := os.ReadFile(config.Config.BoldFontPath)
		if err != nil {
			return fmt.Errorf("could not read subtitle font file: %v", err)
		}
		subParsedFont, err := opentype.Parse(subFontBytes)
		if err != nil {
			return fmt.Errorf("failed to parse subtitle font: %v", err)
		}

		// === 서브타이틀 동적 폰트 크기 조절 ===
		var subFace font.Face
		subFontSize := titleFontSize * subtitleFontRatio
		if subFontSize > subtitleMaxFont {
			subFontSize = subtitleMaxFont
		}

		for {
			var faceErr error
			subFace, faceErr = opentype.NewFace(subParsedFont, &opentype.FaceOptions{
				Size:    subFontSize,
				DPI:     72,
				Hinting: font.HintingFull,
			})
			if faceErr != nil {
				return fmt.Errorf("failed to create subtitle font face: %v", faceErr)
			}

			subTextBounds, _ := font.BoundString(subFace, subTitle)
			subTextWidth := (subTextBounds.Max.X - subTextBounds.Min.X).Ceil()

			if subTextWidth > maxTextWidth {
				subFace.Close()
				subFontSize -= titleFontSizeStep
				if subFontSize < subtitleMinFont {
					break
				}
				continue
			}
			break
		}
		defer subFace.Close()

		subTitleBounds, _ := font.BoundString(subFace, subTitle)
		subTitleWidth := (subTitleBounds.Max.X - subTitleBounds.Min.X).Ceil()

		// 서브타이틀은 타이틀 너비를 기준으로 가운데 정렬
		subTitlePointX := pointX + (textWidth-subTitleWidth)/2
		subTitlePointY := pointY + textHeight + subtitleSpacing

		// 서브타이틀 그림자 (블러 적용)
		subShadowLayer := image.NewRGBA(rgba.Bounds())
		subShadowDrawer := &font.Drawer{
			Dst:  subShadowLayer,
			Src:  image.NewUniform(color.RGBA{R: 0, G: 0, B: 0, A: 180}),
			Face: subFace,
			Dot:  fixed.Point26_6{X: fixed.I(subTitlePointX + subtitleShadowOff), Y: fixed.I(subTitlePointY + subtitleShadowOff)},
		}
		subShadowDrawer.DrawString(subTitle)
		subBlurredShadow := imaging.Blur(subShadowLayer, subtitleBlurSigma)
		draw.Draw(rgba, rgba.Bounds(), subBlurredShadow, image.Point{}, draw.Over)

		// 서브타이틀 외곽선
		smallOffsets := []struct{ dx, dy int }{
			{-subtitleOutlineOff, -subtitleOutlineOff}, {0, -subtitleOutlineOff}, {subtitleOutlineOff, -subtitleOutlineOff},
			{-subtitleOutlineOff, 0}, {subtitleOutlineOff, 0},
			{-subtitleOutlineOff, subtitleOutlineOff}, {0, subtitleOutlineOff}, {subtitleOutlineOff, subtitleOutlineOff},
		}
		for _, off := range smallOffsets {
			subOutlineDrawer := &font.Drawer{
				Dst:  rgba,
				Src:  image.NewUniform(outlineColor),
				Face: subFace,
				Dot:  fixed.Point26_6{X: fixed.I(subTitlePointX + off.dx), Y: fixed.I(subTitlePointY + off.dy)},
			}
			subOutlineDrawer.DrawString(subTitle)
		}

		// 서브타이틀 메인
		subMainDrawer := &font.Drawer{
			Dst:  rgba,
			Src:  image.NewUniform(mainColor),
			Face: subFace,
			Dot:  fixed.Point26_6{X: fixed.I(subTitlePointX), Y: fixed.I(subTitlePointY)},
		}
		subMainDrawer.DrawString(subTitle)
	}

	// Save image
	outputFile, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("could not create output file: %v", err)
	}
	defer outputFile.Close()

	err = png.Encode(outputFile, rgba)
	if err != nil {
		return fmt.Errorf("failed to encode image: %v", err)
	}

	fmt.Printf("Title image created successfully: %s\n", outputPath)
	return nil
}

// GenerateLongformImages 롱폼 전용 이미지 생성 (그림자, 외곽선, 메인 텍스트)
func (s *ImageService) GenerateLongformImages(
	imagePath string,
	eng []string,
	kor []string,
	pronounce []string,
	outputPrefix string,
	count int,
) error {
	// 1. 이미지 불러오기
	existingImageFile, err := os.Open(imagePath)
	if err != nil {
		return fmt.Errorf("이미지 파일을 열 수 없습니다: %v", err)
	}
	defer existingImageFile.Close()

	img, err := png.Decode(existingImageFile)
	if err != nil {
		return fmt.Errorf("이미지 디코딩 실패: %v", err)
	}

	// 2. 폰트 불러오기 (롱폼 전용 볼드 폰트)
	fontBytes, err := os.ReadFile(config.Config.BoldFontPath)
	if err != nil {
		return fmt.Errorf("폰트 파일을 읽을 수 없습니다: %v", err)
	}
	parsedFont, err := opentype.Parse(fontBytes)
	if err != nil {
		return fmt.Errorf("폰트 파싱 실패: %v", err)
	}

	// 3. 색상 정의
	mainColor := color.RGBA{R: 0x4E, G: 0x32, B: 0x15, A: 255} // #4E3215 갈색
	outlineColor := color.RGBA{R: 255, G: 255, B: 255, A: 255} // 흰색 외곽선
	shadowColor := color.RGBA{R: 0, G: 0, B: 0, A: 64}         // #000000 그림자 (불투명도 25%)

	// 4. 배열 길이 검증
	if len(eng) == 0 || len(kor) == 0 || len(pronounce) == 0 {
		return fmt.Errorf("입력 배열이 비어있습니다: eng=%d, kor=%d, pronounce=%d", len(eng), len(kor), len(pronounce))
	}

	expectedLength := count / 2
	if len(eng) < expectedLength || len(kor) < expectedLength || len(pronounce) < expectedLength {
		return fmt.Errorf("배열 길이가 부족합니다: 필요=%d, eng=%d, kor=%d, pronounce=%d",
			expectedLength, len(eng), len(kor), len(pronounce))
	}

	imgWidth := img.Bounds().Dx()
	imgHeight := img.Bounds().Dy()
	maxTextWidth := int(float64(imgWidth) * enum.LongformMaxTextWidthRatio)

	// 5. 이미지들 생성
	for i := 0; i < count; i++ {
		// 원본 이미지 복사
		rgba := image.NewRGBA(img.Bounds())
		draw.Draw(rgba, rgba.Bounds(), img, image.Point{}, draw.Src)

		var text string
		var secondText string // 발음

		if i%2 == 0 { // 짝수 번째 - 한국어
			text = kor[i/2]
		} else { // 홀수 번째 - 영어
			text = eng[i/2]
			secondText = "( " + pronounce[i/2] + " )"
		}

		// 동적 폰트 크기 조절
		var face font.Face
		currentFontSize := enum.LongformMaxFontSize

		for {
			var faceErr error
			face, faceErr = opentype.NewFace(parsedFont, &opentype.FaceOptions{
				Size:    currentFontSize,
				DPI:     72,
				Hinting: font.HintingFull,
			})
			if faceErr != nil {
				return fmt.Errorf("폰트 페이스 생성 실패: %v", faceErr)
			}

			textBounds, _ := font.BoundString(face, text)
			textWidth := (textBounds.Max.X - textBounds.Min.X).Ceil()

			if textWidth > maxTextWidth {
				face.Close()
				currentFontSize -= enum.LongformFontSizeStep
				if currentFontSize < enum.LongformMinFontSize {
					break
				}
				continue
			}
			break
		}

		// 텍스트 위치 계산
		textBounds, _ := font.BoundString(face, text)
		textWidth := (textBounds.Max.X - textBounds.Min.X).Ceil()
		textHeight := (textBounds.Max.Y - textBounds.Min.Y).Ceil()

		pointX := (imgWidth - textWidth) / 2
		pointY := (imgHeight+textHeight)/2 - enum.LongformYOffset

		// === 텍스트 렌더링 (그림자 + 외곽선 + 메인) ===
		drawTextWithShadowAndOutline(rgba, face, text, pointX, pointY, TextRenderOptions{
			ShadowOffset:  enum.LongformShadowOffset,
			OutlineOffset: enum.LongformOutlineOffset,
			MainColor:     mainColor,
			OutlineColor:  outlineColor,
			ShadowColor:   shadowColor,
		})
		face.Close()

		// 발음 텍스트 (두 번째 줄)
		if secondText != "" {
			var smallFace font.Face
			smallFontSize := enum.PronounceMaxFontSize

			// 동적 폰트 크기 조절
			for {
				var faceErr error
				smallFace, faceErr = opentype.NewFace(parsedFont, &opentype.FaceOptions{
					Size:    smallFontSize,
					DPI:     72,
					Hinting: font.HintingFull,
				})
				if faceErr != nil {
					return fmt.Errorf("발음 폰트 페이스 생성 실패: %v", faceErr)
				}

				secondTextBounds, _ := font.BoundString(smallFace, secondText)
				secondTextWidth := (secondTextBounds.Max.X - secondTextBounds.Min.X).Ceil()

				if secondTextWidth > maxTextWidth {
					smallFace.Close()
					smallFontSize -= enum.LongformFontSizeStep
					if smallFontSize < enum.LongformMinFontSize {
						break
					}
					continue
				}
				break
			}

			secondTextBounds, _ := font.BoundString(smallFace, secondText)
			secondTextWidth := (secondTextBounds.Max.X - secondTextBounds.Min.X).Ceil()

			secondPointX := (imgWidth - secondTextWidth) / 2
			secondPointY := pointY + textHeight + enum.PronounceSpacing

			// === 발음 텍스트 렌더링 (그림자 + 외곽선 + 메인) ===
			drawTextWithShadowAndOutline(rgba, smallFace, secondText, secondPointX, secondPointY, TextRenderOptions{
				ShadowOffset:  enum.PronounceShadowOffset,
				OutlineOffset: enum.PronounceOutlineOffset,
				MainColor:     mainColor,
				OutlineColor:  outlineColor,
				ShadowColor:   shadowColor,
			})
			smallFace.Close()
		}

		// 이미지 저장
		outputFileName := fmt.Sprintf("%s_%02d.png", outputPrefix, i+1)
		outputFile, err := os.Create(outputFileName)
		if err != nil {
			return fmt.Errorf("출력 파일을 생성할 수 없습니다: %v", err)
		}

		err = png.Encode(outputFile, rgba)
		if err != nil {
			outputFile.Close()
			return fmt.Errorf("이미지 인코딩 실패: %v", err)
		}
		outputFile.Close()

		fmt.Printf("이미지 %d 생성 완료: %s\n", i+1, outputFileName)
	}

	fmt.Printf("모든 %d장의 이미지가 성공적으로 생성되었습니다.\n", count)
	return nil
}
