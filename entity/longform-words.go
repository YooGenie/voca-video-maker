package entity

type LongformWord struct {
	Id              int64  `xorm:"id pk autoincr"`
	Word            string `xorm:"word notnull"`
	Meaning         string `xorm:"meaning notnull"`
	PronunciationKr string `xorm:"pronunciation_kr"`
	PhoneticSymbol  string `xorm:"phonetic_symbol"`
	CreatedDate     string `xorm:"created_date notnull"`
	ShortsDate      string `xorm:"shorts_date"`  // 유튜브 숏폼 생성 기준일
	ContentType     string `xorm:"content_type"` // 콘텐츠 종류: word, idiom, sentence
}

func (LongformWord) TableName() string {
	return "longform_words"
}
