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

	// TAMBAHKAN BAGIAN INI AGAR GORM BISA MELAKUKAN PRELOAD
	Kelas      Kelas         `gorm:"foreignKey:IDKelas" json:"kelas,omitempty"`
	MataKuliah MataKuliah    `gorm:"foreignKey:IDMk" json:"mata_kuliah,omitempty"`
	Ruangan    Ruangan       `gorm:"foreignKey:IDRuangan" json:"ruangan,omitempty"`
	Dosen      *Dosen        `gorm:"foreignKey:IDDosen" json:"dosen,omitempty"`
	Asdos1     *AsistenDosen `gorm:"foreignKey:IDAsdos1" json:"asdos1,omitempty"`
	Asdos2     *AsistenDosen `gorm:"foreignKey:IDAsdos2" json:"asdos2,omitempty"`
	Semester   Semester      `gorm:"foreignKey:IDSemester" json:"semester,omitempty"`
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
