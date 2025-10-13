package entity

type EnglishWord struct {
	Id                int64  `xorm:"id pk autoincr"`
	EnglishWord       string `xorm:"english_word"`
	Meaning           string `xorm:"meaning notnull"`
	PronunciationKr   string `xorm:"pronunciation_kr"`
	PhoneticSymbol    string `xorm:"phonetic_symbol"`
	CreatedDate       string `xorm:"created_date"`
}

func (EnglishWord) TableName() string {
	return "english_words"
}
