package db

import (
	"github.com/jmoiron/sqlx"
	"github.com/workos-code-challenge/matthew-ault/models"
)

func NewPostgres(dbConn *sqlx.DB) *Postgres {
	return &Postgres{dbConn: dbConn}
}

// Postgres implements the Storer interface
type Postgres struct {
	dbConn *sqlx.DB
}

// id, name, deleted, real_name, tz, profile_status_text, profile_status_emoji, profile_image_512

func (p *Postgres) CreateUser(user models.User) error {
	_, err := p.dbConn.NamedExec(`INSERT INTO users (id, name, deleted, real_name, tz, profile_status_text, profile_status_emoji, profile_image_512) VALUES (:id, :name, :deleted, :real_name, :tz, :profile_status_text, :profile_status_emoji, :profile_image_512)`, user)
	return err
}

func (p *Postgres) CreateUsers(users []models.User) error {
	// Want to do an upsert as if the service has been down the db may be stale
	// ON CONFLICT does not seem to work with sqlx namedexec
	//_, err := p.dbConn.NamedExec(`INSERT INTO users (id, name, deleted, real_name, tz, profile_status_text, profile_status_emoji, profile_image_512) VALUES (:id, :name, :deleted, :real_name, :tz, :profile_status_text, :profile_status_emoji, :profile_image_512) ON CONFLICT (id) DO UPDATE SET name=:name, deleted=:deleted, real_name=:real_name, tz=:tz, profile_status_text=:profile_status_text, profile_status_emoji=:profile_status_emoji, profile_image_512=:profile_image_512`, users)
	var err error
	for _, user := range users {
		// this is inefficient but is only run at app start
		// TODO: optimise
		_, err = p.dbConn.Exec(`INSERT INTO users VALUES ($1,$2,$3,$4,$5,$6,$7,$8) ON CONFLICT (id) DO UPDATE SET name=$9, deleted=$10, real_name=$11, tz=$12, profile_status_text=$13, profile_status_emoji=$14, profile_image_512=$15`,
			user.ID, user.Name, user.Deleted, user.RealName, user.Tz, user.ProfileStatusText, user.ProfileStatusEmoji, user.ProfileImage512, user.Name, user.Deleted, user.RealName, user.Tz, user.ProfileStatusText, user.ProfileStatusEmoji, user.ProfileImage512)
		if err != nil {
			return err
		}
	}
	return nil
}

func (p *Postgres) UpdateUser(user models.User) error {
	_, err := p.dbConn.NamedExec(`UPDATE users SET name=:name, deleted=:deleted, real_name=:real_name, tz=:tz, profile_status_text=:profile_status_text, profile_status_emoji=:profile_status_emoji, profile_image_512=:profile_image_512 WHERE id=:id`, user)
	return err
}

// may need pagination here
func (p *Postgres) GetAllUsers() ([]models.User, error) {
	var users []models.User
	err := p.dbConn.Select(&users, "SELECT * FROM users")
	return users, err
}
