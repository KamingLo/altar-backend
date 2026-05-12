package models

import (
	"altar/utils"
	"time"

	"gorm.io/gorm"
)

type Dosen struct {
	ID        string         `gorm:"primaryKey;type:varchar(36)" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	NIP  string `gorm:"column:nip" json:"nip"`
	Nama string `json:"nama"`
}

func init() {
	RegisterModel(&Dosen{})
}

func (o *Dosen) BeforeCreate(tx *gorm.DB) (err error) {
	o.ID = utils.GenerateCustomID("dos", 6)
	return nil
}

