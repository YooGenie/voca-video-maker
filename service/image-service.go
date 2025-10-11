package service

import (
	"auto-video-service/config"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"os"

	"golang.org/x/image/font"
	"golang.org/x/image/font/opentype"
	"golang.org/x/image/math/fixed"
)

// ImageService 이미지 생성 서비스
type ImageService struct{}

// NewImageService 새로운 이미지 서비스 생성
func NewImageService() *ImageService {
	return &ImageService{}
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
		Size:    120, // 폰트 크기 한글
		DPI:     72,  // DPI (Dots Per Inch)
		Hinting: font.HintingNone,
	})
	if err != nil {
		return fmt.Errorf("폰트 페이스 생성 실패: %v", err)
	}
	defer face.Close()

	// 3. 배열 길이 검증
	if len(eng) == 0 || len(kor) == 0 || len(pronounce) == 0 {
		return fmt.Errorf("입력 배열이 비어있습니다: eng=%d, kor=%d, pronounce=%d", len(eng), len(kor), len(pronounce))
	}
	
	expectedLength := count / 2
	if len(eng) < expectedLength || len(kor) < expectedLength || len(pronounce) < expectedLength {
		return fmt.Errorf("배열 길이가 부족합니다: 필요=%d, eng=%d, kor=%d, pronounce=%d", 
			expectedLength, len(eng), len(kor), len(pronounce))
	}

	// 4. 이미지들 생성
	textColor := color.RGBA{R: 255, G: 255, B: 255, A: 255} // 흰색
	fmt.Println("=======>",count)
	for i := 0; i < count; i++ {
		// 원본 이미지 복사
		rgba := image.NewRGBA(img.Bounds())
		draw.Draw(rgba, rgba.Bounds(), img, image.Point{}, draw.Src)

		var text string
		var secondText string
		if i%2 == 0 { // 홀수 번째 (0, 2, 4, ...) - 영어
			text = kor[i/2]
		} else { // 짝수 번째 (1, 3, 5, ...) - 한국어 + 발음
			text = eng[i/2]
			secondText = "( " + pronounce[i/2] + " )"
		}

		// 텍스트 위치 계산
		textBounds, _ := font.BoundString(face, text)
		textWidth := (textBounds.Max.X - textBounds.Min.X).Ceil()
		textHeight := (textBounds.Max.Y - textBounds.Min.Y).Ceil()

		imgWidth := rgba.Bounds().Dx()
		imgHeight := rgba.Bounds().Dy()

		pointX := (imgWidth - textWidth) / 2
		pointY := (imgHeight + textHeight) / 2

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

		// 두 번째 텍스트가 있으면 아래에 그리기 (작은 폰트 사용)
		if secondText != "" {
			// 작은 폰트로 두 번째 텍스트 그리기
			smallFace, err := opentype.NewFace(parsedFont, &opentype.FaceOptions{
				Size:    75, // 작은 폰트 크기
				DPI:     72, // DPI (Dots Per Inch)
				Hinting: font.HintingNone,
			})
			if err != nil {
				return fmt.Errorf("작은 폰트 페이스 생성 실패: %v", err)
			}
			defer smallFace.Close()

			smallDrawer := &font.Drawer{
				Dst:  rgba,
				Src:  image.NewUniform(textColor),
				Face: smallFace,
			}

			secondTextBounds, _ := font.BoundString(smallFace, secondText)
			secondTextWidth := (secondTextBounds.Max.X - secondTextBounds.Min.X).Ceil()

			secondPointX := (imgWidth - secondTextWidth) / 2
			secondPointY := pointY + textHeight + 20 // 첫 번째 텍스트 아래 20픽셀 간격

			secondPoint := fixed.Point26_6{
				X: fixed.I(secondPointX),
				Y: fixed.I(secondPointY),
			}

			smallDrawer.Dot = secondPoint
			smallDrawer.DrawString(secondText)
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

// GenerateSingleImage 단일 이미지를 생성합니다
func (s *ImageService) GenerateSingleImage(
	imagePath string,
	text string,
	outputPath string,
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
		Size:    120, // 폰트 크기
		DPI:     72,  // DPI (Dots Per Inch)
		Hinting: font.HintingNone,
	})
	if err != nil {
		return fmt.Errorf("폰트 페이스 생성 실패: %v", err)
	}
	defer face.Close()

	// 3. 이미지 생성
	rgba := image.NewRGBA(img.Bounds())
	draw.Draw(rgba, rgba.Bounds(), img, image.Point{}, draw.Src)

	textColor := color.RGBA{R: 255, G: 255, B: 255, A: 255} // 흰색

	// 텍스트 위치 계산
	textBounds, _ := font.BoundString(face, text)
	textWidth := (textBounds.Max.X - textBounds.Min.X).Ceil()
	textHeight := (textBounds.Max.Y - textBounds.Min.Y).Ceil()

	imgWidth := rgba.Bounds().Dx()
	imgHeight := rgba.Bounds().Dy()

	pointX := (imgWidth - textWidth) / 2
	pointY := (imgHeight + textHeight) / 2

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

	// 이미지 저장
	outputFile, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("출력 파일을 생성할 수 없습니다: %v", err)
	}
	defer outputFile.Close()

	err = png.Encode(outputFile, rgba)
	if err != nil {
		return fmt.Errorf("이미지 인코딩 실패: %v", err)
	}

	fmt.Printf("이미지가 성공적으로 생성되었습니다: %s\n", outputPath)
	return nil
}

// GenerateImagesFromText 텍스트 배열로부터 이미지들을 생성합니다 (인스타그램 릴스용)
func (s *ImageService) GenerateImagesFromText(
	imagePath string,
	textData []string,
	outputPrefix string,
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

	// 3. 이미지들 생성
	textColor := color.RGBA{R: 255, G: 255, B: 255, A: 255} // 흰색

	for i, text := range textData {
		// 원본 이미지 복사
		rgba := image.NewRGBA(img.Bounds())
		draw.Draw(rgba, rgba.Bounds(), img, image.Point{}, draw.Src)

		// 폰트 옵션 설정 (인스타그램 릴스용으로 크기 조정)
		face, err := opentype.NewFace(parsedFont, &opentype.FaceOptions{
			Size:    80, // 인스타그램 릴스용으로 폰트 크기 조정
			DPI:     72,
			Hinting: font.HintingNone,
		})
		if err != nil {
			return fmt.Errorf("폰트 페이스 생성 실패: %v", err)
		}
		defer face.Close()

		// 텍스트 위치 계산
		textBounds, _ := font.BoundString(face, text)
		textWidth := (textBounds.Max.X - textBounds.Min.X).Ceil()
		textHeight := (textBounds.Max.Y - textBounds.Min.Y).Ceil()

		imgWidth := rgba.Bounds().Dx()
		imgHeight := rgba.Bounds().Dy()

		pointX := (imgWidth - textWidth) / 2
		pointY := (imgHeight + textHeight) / 2

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

	fmt.Printf("모든 %d장의 이미지가 성공적으로 생성되었습니다.\n", len(textData))
	return nil
}

// GenerateOptionalImage wordCount 값을 이미지에 표시하는 이미지를 생성합니다
func (s *ImageService) GenerateOptionalImage(
	imagePath string,
	wordCountText string,
	outputPrefix string,
	serviceType string,
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
		Size:    100, // wordCount용 폰트 크기 (80에서 100으로 증가)
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
	switch serviceType {
	case "W":
		textColor = color.RGBA{R: 173, G: 216, B: 230, A: 255} // #ADD8E6 (연한 파란색)
	case "I":
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
	pointY := textHeight + 485

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

// GenerateTitleImage creates an image with a centered title.
func (s *ImageService) GenerateTitleImage(title, imagePath, outputPath string) error {
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

	// 2. Load font
	fontBytes, err := os.ReadFile(config.Config.FontPath)
	if err != nil {
		return fmt.Errorf("could not read font file: %v", err)
	}

	parsedFont, err := opentype.Parse(fontBytes)
	if err != nil {
		return fmt.Errorf("failed to parse font: %v", err)
	}

	// Set font options
	face, err := opentype.NewFace(parsedFont, &opentype.FaceOptions{
		Size:    120, // Font size
		DPI:     72,  // DPI (Dots Per Inch)
		Hinting: font.HintingNone,
	})
	if err != nil {
		return fmt.Errorf("failed to create font face: %v", err)
	}
	defer face.Close()

	// 3. Create image
	rgba := image.NewRGBA(img.Bounds())
	draw.Draw(rgba, rgba.Bounds(), img, image.Point{}, draw.Src)

	textColor := color.RGBA{R: 255, G: 255, B: 255, A: 255} // White

	// Calculate text position
	textBounds, _ := font.BoundString(face, title)
	textWidth := (textBounds.Max.X - textBounds.Min.X).Ceil()
	textHeight := (textBounds.Max.Y - textBounds.Min.Y).Ceil()

	imgWidth := rgba.Bounds().Dx()
	imgHeight := rgba.Bounds().Dy()

	pointX := (imgWidth - textWidth) / 2
	pointY := (imgHeight + textHeight) / 2

	point := fixed.Point26_6{
		X: fixed.I(pointX),
		Y: fixed.I(pointY),
	}

	// Draw text on image
	d := &font.Drawer{
		Dst:  rgba,
		Src:  image.NewUniform(textColor),
		Face: face,
		Dot:  point,
	}
	d.DrawString(title)

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
