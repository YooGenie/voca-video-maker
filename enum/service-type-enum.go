package enum

type ServiceType string

const (
	// 인스타그램 숏폼 (shorts DB 사용)
	InstagramWord     ServiceType = "iw"
	InstagramIdiom    ServiceType = "ii"
	InstagramSentence ServiceType = "is"

	// 페이스북 숏폼 (shorts DB 사용, 느린 속도, 3회 반복)
	FacebookWord     ServiceType = "fw"
	FacebookIdiom    ServiceType = "fi"
	FacebookSentence ServiceType = "fs"

	// 유튜브 숏폼 (longform_words 사용, shorts_date + content_type 조회)
	YoutubeShortsWord    ServiceType = "ysw"
	YoutubeShortsIdiom   ServiceType = "ysi"
	YoutubeShotsSentence ServiceType = "yss"

	// 유튜브 롱폼 (longform_words 사용, date 조회)
	YoutubeLongform ServiceType = "yl"

	// 기타
	Start ServiceType = "start"
)
