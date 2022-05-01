package models

type User struct {
	Deleted            bool   `json:"deleted" db:"deleted"`
	ID                 string `json:"id" db:"id"`
	Name               string `json:"name" db:"name"`
	ProfileImage512    string `json:"image_512" db:"profile_image_512"`
	ProfileStatusEmoji string `json:"status_emoji" db:"profile_status_emoji"`
	ProfileStatusText  string `json:"status_text" db:"profile_status_text"`
	RealName           string `json:"real_name" db:"real_name"`
	TZ                 string `json:"tz" db:"tz"`
}
