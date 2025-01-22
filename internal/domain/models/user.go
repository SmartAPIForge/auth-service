package models

type User struct {
	ID       int64
	Email    string
	Password []byte
	Role     int64 `db:"role_id"`
}
