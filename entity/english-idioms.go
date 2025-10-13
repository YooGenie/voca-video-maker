package entity

type EnglishIdiom struct {
	Id                int64  `xorm:"id pk autoincr"`
	Idiom             string `xorm:"idiom"`
	Meaning           string `xorm:"meaning notnull"`
	PronunciationKr   string `xorm:"pronunciation_kr"`
	PhoneticSymbol    string `xorm:"phonetic_symbol"`
	CreatedDate       string `xorm:"created_date"`
}

func (EnglishIdiom) TableName() string {
	return "english_idioms"
}
