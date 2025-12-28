package enum

// ContentType 콘텐츠 종류 (단어, 숙어, 문장)
type ContentType string

const (
	ContentWord     ContentType = "word"
	ContentIdiom    ContentType = "idiom"
	ContentSentence ContentType = "sentence"
)
