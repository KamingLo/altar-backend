package models

import (
	"altar/utils"
	"time"

	"gorm.io/gorm"
)

type Semester struct {
	ID           string         `gorm:"primaryKey;type:varchar(36)" json:"id"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`
	
	TahunAjaran  string         `json:"tahun_ajaran"`  // cth: "2026/2027"
	TipeSemester string         `json:"tipe_semester"` // cth: "Ganjil", "Genap", "Pendek"
}

func init() {
	RegisterModel(&Semester{})
}

func (o *Semester) BeforeCreate(tx *gorm.DB) (err error) {
	o.ID = utils.GenerateCustomID("sms", 6)
	return nil
}