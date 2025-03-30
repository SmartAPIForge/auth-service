package models

type User struct {
	ID       int64
	Username string
	Email    string
	Password []byte
	Role     int64 `db:"role_id"`
}
