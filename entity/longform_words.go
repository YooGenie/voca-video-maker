package entity

type LongformWord struct {
	Id              int64  `xorm:"id pk autoincr"`
	Word            string `xorm:"word notnull"`
	Meaning         string `xorm:"meaning notnull"`
	PronunciationKr string `xorm:"pronunciation_kr"`
	PhoneticSymbol  string `xorm:"phonetic_symbol"`
	CreatedDate     string `xorm:"created_date notnull"`
}

func (LongformWord) TableName() string {
	return "longform_words"
}