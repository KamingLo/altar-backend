package models

import (
	"altar/utils"
	"time"

	"gorm.io/gorm"
)

type Kelas struct {
	ID        string         `gorm:"primaryKey;type:varchar(36)" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	NamaKelas   string `json:"nama_kelas"` // cth: "TI A"
	Jurusan     string `json:"jurusan"`    // cth: "Teknik Informatika"
	JumlahSiswa int    `json:"jumlah_siswa"`
}

func init() {
	RegisterModel(&Kelas{})
}

func (o *Kelas) BeforeCreate(tx *gorm.DB) (err error) {
	o.ID = utils.GenerateCustomID("kls", 6)
	return nil
}
