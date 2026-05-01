package models

import (
	"altar/utils"
	"time"

	"gorm.io/gorm"
)

type AsistenDosen struct {
	ID        string         `gorm:"primaryKey;type:varchar(36)" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	UserID      string `gorm:"type:varchar(36)" json:"user_id"`
	NIM         string `json:"nim"`
	PhoneNumber string `json:"phone_number"`

	User User `gorm:"foreignKey:UserID" json:"user"`
}

func init() {
	RegisterModel(&AsistenDosen{})
}

func (a *AsistenDosen) BeforeCreate(tx *gorm.DB) (err error) {
	a.ID = utils.GenerateCustomID("asd", 6)
	return nil
}
