package entity

type Title struct {
	Id          int64  `xorm:"id pk autoincr"`
	Title       string `xorm:"title notnull"`
	CreatedDate string `xorm:"created_date notnull"`
	IsUploaded  bool   `xorm:"is_uploaded"`
}

func (Title) TableName() string {
	return "title"
}
