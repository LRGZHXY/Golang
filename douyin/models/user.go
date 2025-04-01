package models

type User struct {
	ID       string `gorm:"primaryKey" json:"id"`
	Nickname string `gorm:"not null;size:255" json:"nickname"`
	Password string `gorm:"not null;size:255" json:"-"`
	Email    string `gorm:"unique;not null;size:255" json:"email"`
	//Verified bool   `gorm:"default:false" json:"verified"`
}
