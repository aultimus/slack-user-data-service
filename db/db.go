package db

import (
	"errors"

	"github.com/jmoiron/sqlx"
)

// User is the database representation of a user
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

func NewPostgres(dbConn *sqlx.DB) *Postgres {
	return &Postgres{dbConn: dbConn}
}

// Postgres implements the Storer interface
type Postgres struct {
	dbConn *sqlx.DB
}

func (p *Postgres) CreateUsers(users []User) error {
	// Want to do an upsert as if the service has been down the db may be stale
	// ON CONFLICT does not seem to work with sqlx namedexec
	//_, err := p.dbConn.NamedExec(`INSERT INTO users (id, name, deleted, real_name, tz, profile_status_text, profile_status_emoji, profile_image_512) VALUES (:id, :name, :deleted, :real_name, :tz, :profile_status_text, :profile_status_emoji, :profile_image_512) ON CONFLICT (id) DO UPDATE SET name=:name, deleted=:deleted, real_name=:real_name, tz=:tz, profile_status_text=:profile_status_text, profile_status_emoji=:profile_status_emoji, profile_image_512=:profile_image_512`, users)
	tx, err := p.dbConn.Beginx() // put multiple inserts in transaction to speed up
	if err != nil {
		return err
	}
	for _, user := range users {
		_, err = tx.Exec(`INSERT INTO users VALUES ($1,$2,$3,$4,$5,$6,$7,$8) ON CONFLICT (id) DO UPDATE SET name=$9, deleted=$10, real_name=$11, tz=$12, profile_status_text=$13, profile_status_emoji=$14, profile_image_512=$15`,
			user.ID, user.Name, user.Deleted, user.RealName, user.TZ, user.ProfileStatusText, user.ProfileStatusEmoji, user.ProfileImage512, user.Name, user.Deleted, user.RealName, user.TZ, user.ProfileStatusText, user.ProfileStatusEmoji, user.ProfileImage512)
		if err != nil {
			rollbackErr := tx.Rollback()
			if rollbackErr != nil {
				err = errors.New(err.Error() + ":" + rollbackErr.Error())
			}
			return err
		}
	}
	err = tx.Commit()
	return err
}

func (p *Postgres) UpdateUser(user User) error {
	return p.CreateUsers([]User{user})
}

// may need pagination here
func (p *Postgres) GetAllUsers() ([]User, error) {
	var users []User
	// if database gets significantly large then we may not want to load all
	// users into memory at once
	err := p.dbConn.Select(&users, "SELECT * FROM users ORDER BY id")
	return users, err
}
