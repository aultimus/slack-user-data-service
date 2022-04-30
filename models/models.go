package models

type User struct {
	Deleted            bool   `json:"deleted" db:"deleted"`
	ID                 string `json:"id" db:"id"`
	Name               string `json:"name" db:"name"`
	ProfileImage512    string `json:"image_512" db:"profile_image_512"`
	ProfileStatusEmoji string `json:"status_emoji" db:"profile_status_emoji"`
	ProfileStatusText  string `json:"status_text" db:"profile_status_text"`
	RealName           string `json:"real_name" db:"real_name"`
	Tz                 string `json:"tz" db:"tz"`
}

// Event represents a slack event data type which is received by the webhooks
// endpoint. This struct was initally generated using https://mholt.github.io/json-to-go/
// id, name, deleted, real_name, tz, profile object (status_text, status_emoji, image_512).
type Event struct {
	APIAppID       string `json:"api_app_id"`
	Authorizations []struct {
		EnterpriseID        interface{} `json:"enterprise_id"`
		IsBot               bool        `json:"is_bot"`
		IsEnterpriseInstall bool        `json:"is_enterprise_install"`
		TeamID              string      `json:"team_id"`
		UserID              string      `json:"user_id"`
	} `json:"authorizations"`
	Event struct {
		CacheTs int    `json:"cache_ts"`
		EventTs string `json:"event_ts"`
		Type    string `json:"type"`
		User    User   `json:"user"`
	} `json:"event"`
	EventID            string `json:"event_id"`
	EventTime          int    `json:"event_time"`
	IsExtSharedChannel bool   `json:"is_ext_shared_channel"`
	TeamID             string `json:"team_id"`
	Token              string `json:"token"`
	Type               string `json:"type"`
}
