package models

import (
	"altar/utils"
	"time"

	"gorm.io/gorm"
)

type JadwalUtama struct {
	ID            string         `gorm:"primaryKey;type:varchar(36)" json:"id"`
	IDKelas       string         `gorm:"type:varchar(36);not null" json:"id_kelas"`
	IDMk          string         `gorm:"type:varchar(36);not null" json:"id_mk"`
	IDSemester    string         `gorm:"type:varchar(36);not null" json:"id_semester"`
	IDRuangan     string         `gorm:"type:varchar(36);not null" json:"id_ruangan"`
	IDAsdos1      *string        `gorm:"type:varchar(36)" json:"id_asdos1"`
	IDAsdos2      *string        `gorm:"type:varchar(36)" json:"id_asdos2"`
	IDDosen       *string        `gorm:"type:varchar(36)" json:"id_dosen"`
	KelasMulai    time.Time      `gorm:"not null" json:"kelas_mulai"`
	KelasBerakhir time.Time      `gorm:"not null" json:"kelas_berakhir"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
	DeletedAt     gorm.DeletedAt `gorm:"index" json:"-"`
}

func init() {
	RegisterModel(&JadwalUtama{})
}

func (j *JadwalUtama) BeforeCreate(tx *gorm.DB) (err error) {
	j.ID = utils.GenerateCustomID("ses", 6)
	return nil
}

func (j *JadwalUtama) IsDosenSession() bool {
	return j.IDDosen != nil
}

func (j *JadwalUtama) IsAsdosSession() bool {
	return j.IDAsdos1 != nil
}
