package models

import (
	"altar/utils"
	"time"

	"gorm.io/gorm"
)

type Koordinator struct {
	ID        string         `gorm:"primaryKey;type:varchar(36)" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	UserID string `gorm:"type:varchar(36)" json:"user_id"`
	NIP    string `json:"nip"`

	User User `gorm:"foreignKey:UserID" json:"user"`
}

func init() {
	RegisterModel(&Koordinator{})
}

func (k *Koordinator) BeforeCreate(tx *gorm.DB) (err error) {
	k.ID = utils.GenerateCustomID("koo", 6)
	return nil
}
