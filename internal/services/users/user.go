package users

import (
	"time"

	"github.com/prismelabs/prismeanalytics/internal/secret"
)

type User struct {
	Id        UserId
	Email     Email
	Password  secret.Secret[string]
	Name      UserName
	CreatedAt time.Time
}
