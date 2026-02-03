package entity

import "time"

// User represents a core domain entity without infrastructure concerns.
type User struct {
	ID        int64
	Email     string
	Firstname string
	Lastname  string
	Phone     string
	Gender    string
	AvatarPic string
	CreatedAt time.Time
	UpdatedAt time.Time
}
