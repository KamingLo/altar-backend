package models

import (
	"altar/utils"
	"time"

	"gorm.io/gorm"
)

type Ruangan struct {
	ID          string         `gorm:"primaryKey;type:varchar(36)" json:"id"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
	
	NamaRuangan string         `json:"nama_ruangan"` // cth: "Lab Komputer 1", "R.305"
	Lantai      int            `json:"lantai"`
	Kapasitas   int            `json:"kapasitas"`
}

func init() {
	RegisterModel(&Ruangan{})
}

func (o *Ruangan) BeforeCreate(tx *gorm.DB) (err error) {
	o.ID = utils.GenerateCustomID("rn", 6)
	return nil
}
