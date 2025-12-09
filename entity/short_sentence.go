package entity

import "database/sql"

// ShortSentence represents the structure of the short_sentences table.
type ShortSentence struct {
	Id               int64          `xorm:"id pk autoincr"`
	KoreanSentence1  string         `xorm:"korean_sentence_1 notnull"`
	KoreanSentence2  sql.NullString `xorm:"korean_sentence_2"`
	EnglishSentence1 string         `xorm:"english_sentence_1 notnull"`
	EnglishSentence2 sql.NullString `xorm:"english_sentence_2"`
	Pronunciation    string         `xorm:"pronunciation notnull"`
	CreatedDate      string         `xorm:"created_date notnull"`
}

func (ShortSentence) TableName() string {
	return "short_sentences"
}
