package users

import "time"

type User struct {
	Id        UserId
	Email     Email
	Password  Password
	Name      UserName
	CreatedAt time.Time
}
