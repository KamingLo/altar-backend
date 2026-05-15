package models

import (
	"altar/utils"
	"time"

	"gorm.io/gorm"
)

// ─────────────────────────────────────────────
// Enum: SubstituteSessionStatus
// ─────────────────────────────────────────────

type SubstituteSessionStatus string

const (
	StatusPending  SubstituteSessionStatus = "PENDING"
	StatusVerified SubstituteSessionStatus = "VERIFIED"
	StatusRejected SubstituteSessionStatus = "REJECTED"
)

// ─────────────────────────────────────────────
// Model: SubstituteSession (JadwalPengganti)
// ─────────────────────────────────────────────

type SubstituteSession struct {
	ID                   string                  `gorm:"primaryKey;type:varchar(36)"              json:"id"`
	IDSession            string                  `gorm:"type:varchar(36);not null"                json:"id_session"`
	IDRuangan            string                  `gorm:"type:varchar(36);not null"                json:"id_ruangan"`
	IDAsdosPengganti     *string                 `gorm:"type:varchar(36)"                         json:"id_asdos_pengganti"` // nullable — nil means the original teacher handles the session themselves
	Reason               string                  `gorm:"type:text;not null"                       json:"reason"`
	Status               SubstituteSessionStatus `gorm:"type:varchar(20);not null;default:'PENDING'" json:"status"`
	SubstituteDate       time.Time               `gorm:"not null"                                 json:"substitute_date"`  // actual replacement date (YYYY-MM-DD)
	OriginalDate         time.Time               `gorm:"not null;type:date"                       json:"original_date"`    // cancelled regular schedule date (YYYY-MM-DD)
	KelasMulai           time.Time               `gorm:"not null"                                 json:"kelas_mulai"`
	KelasBerakhir        time.Time               `gorm:"not null"                                 json:"kelas_berakhir"`
	CreatedAt            time.Time               `json:"created_at"`
	UpdatedAt            time.Time               `json:"updated_at"`
	DeletedAt            gorm.DeletedAt          `gorm:"index"                                    json:"-"`

	// Relations (preload only — no FK constraint enforced by GORM tags here)
	Session             *JadwalUtama  `gorm:"foreignKey:IDSession"           json:"session,omitempty"`
	Ruangan             *Ruangan      `gorm:"foreignKey:IDRuangan"           json:"ruangan,omitempty"`
	AsdosPengganti      *AsistenDosen `gorm:"foreignKey:IDAsdosPengganti"    json:"asdos_pengganti,omitempty"` // the asdos who takes over the class
}

func init() {
	RegisterModel(&SubstituteSession{})
}

func (s *SubstituteSession) BeforeCreate(tx *gorm.DB) (err error) {
	s.ID = utils.GenerateCustomID("sub", 6)
	s.Status = StatusPending
	return nil
}
