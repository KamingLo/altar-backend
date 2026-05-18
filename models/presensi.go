package models

import (
	"altar/utils"
	"time"

	"gorm.io/gorm"
)

type TipeAbsensi string

const (
	AbsensiQR   TipeAbsensi = "qr"
	AbsensiLink TipeAbsensi = "link"
)

type Presensi struct {
	IDPresensi        string         `gorm:"primaryKey;type:varchar(36)" json:"id_presensi"`
	IDSesi            string         `gorm:"type:varchar(36);not null" json:"id_sesi"`
	IDSesiPengganti   *string        `gorm:"type:varchar(36);nullable" json:"id_sesi_pengganti"`
	IDAsdosPelaksana  string         `gorm:"type:varchar(36);not null" json:"id_asdos_pelaksana"`
	IDAsdosRekan      *string        `gorm:"type:varchar(36);nullable" json:"id_asdos_rekan"`
	TipeAbsensi       TipeAbsensi    `gorm:"type:varchar(10);not null;default:'qr'" json:"tipe_absensi"`
	LinkVideo         *string        `gorm:"type:text;nullable" json:"link_video"`
	Menggantikan      bool           `gorm:"type:boolean;not null;default:false" json:"menggantikan"`
	WaktuCheckIn      time.Time      `gorm:"not null" json:"waktu_checkin"`
	WaktuCheckOut     *time.Time     `gorm:"nullable" json:"waktu_checkout"`
	TanggalMengajar   time.Time      `gorm:"type:date;not null" json:"tanggal_mengajar"`
	DeskripsiMateri   *string        `gorm:"type:text;nullable" json:"deskripsi_materi"`
	IsVerified        bool           `gorm:"default:false" json:"is_verified"`
	CreatedAt         time.Time      `json:"created_at"`
	UpdatedAt         time.Time      `json:"updated_at"`
	DeletedAt         gorm.DeletedAt `gorm:"index" json:"-"`

	// Eager Loading Relational Fields
	JadwalUtama       JadwalUtama        `gorm:"foreignKey:IDSesi" json:"jadwal_utama,omitempty"`
	SubstituteSession *SubstituteSession `gorm:"foreignKey:IDSesiPengganti" json:"substitute_session,omitempty"`
	AsdosPelaksana    AsistenDosen       `gorm:"foreignKey:IDAsdosPelaksana" json:"asdos_pelaksana,omitempty"`
	AsdosRekan        *AsistenDosen      `gorm:"foreignKey:IDAsdosRekan" json:"asdos_rekan,omitempty"`
}

func init() {
	RegisterModel(&Presensi{})
}

func (p *Presensi) BeforeCreate(tx *gorm.DB) (err error) {
	p.IDPresensi = utils.GenerateCustomID("att", 6)
	return nil
}
