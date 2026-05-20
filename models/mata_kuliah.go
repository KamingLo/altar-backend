package models

import (
	"altar/utils"
	"time"

	"gorm.io/gorm"
)

type MataKuliah struct {
	ID        string         `gorm:"primaryKey;type:varchar(36)" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	NamaMK string `json:"nama_mk"`
	SKS    int    `json:"sks"`
}

func init() {
	RegisterModel(&MataKuliah{})
}

func (o *MataKuliah) BeforeCreate(tx *gorm.DB) (err error) {
	o.ID = utils.GenerateCustomID("mk", 6)
	return nil
}
