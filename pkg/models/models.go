package models

type User struct {
	Id          uint   `json:"id" gorm:"primaryKey;not null"`
	Name        string `json:"name" gorm:"not null"`
	HouseName   string `json:"houseName" gorm:"not null"`
	City        string `json:"city" gorm:"not null"`
	Email       string `json:"email" gorm:"unique; not null"`
	PhoneNumber int    `json:"phoneNumber" gorm:"not null"`
}
