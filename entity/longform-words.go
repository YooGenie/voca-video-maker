package entity

type LongformWord struct {
	Id              int64  `xorm:"id pk autoincr"`
	Word            string `xorm:"word notnull"`
	Meaning         string `xorm:"meaning notnull"`
	PronunciationKr string `xorm:"pronunciation_kr"`
	PhoneticSymbol  string `xorm:"phonetic_symbol"`
	Source          string `xorm:"source notnull"`
	CreatedDate     string `xorm:"created_date notnull"`
	ShortsDate      string `xorm:"shorts_date"`
	ContentType     string `xorm:"content_type"`
}

func (LongformWord) TableName() string {
	return "longform_words"
}
