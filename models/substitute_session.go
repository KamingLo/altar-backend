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
	ID                string                  `gorm:"primaryKey;type:varchar(36)"              json:"id"`
	IDSession         string                  `gorm:"type:varchar(36);not null"                json:"id_session"`
	IDRuangan         string                  `gorm:"type:varchar(36);not null"                json:"id_ruangan"`
	IDDosen           *string                 `gorm:"type:varchar(36)"                         json:"id_dosen"`
	IDAsdos1          *string                 `gorm:"type:varchar(36)"                         json:"id_asdos1"`
	IDAsdos2          *string                 `gorm:"type:varchar(36)"                         json:"id_asdos2"`
	Reason            string                  `gorm:"type:text;not null"                       json:"reason"`
	CoordinatorReason *string                 `gorm:"type:text"                                json:"coordinator_reason"` // nullable — the coordinator's note on approval/rejection
	Status            SubstituteSessionStatus `gorm:"type:varchar(20);not null;default:'PENDING'" json:"status"`
	SubstituteDate    time.Time               `gorm:"not null"                                 json:"substitute_date"` // actual replacement date (YYYY-MM-DD)
	OriginalDate      time.Time               `gorm:"not null;type:date"                       json:"original_date"`   // cancelled regular schedule date (YYYY-MM-DD)
	KelasMulai        time.Time               `gorm:"not null"                                 json:"kelas_mulai"`
	KelasBerakhir     time.Time               `gorm:"not null"                                 json:"kelas_berakhir"`
	CreatedAt         time.Time               `json:"created_at"`
	UpdatedAt         time.Time               `json:"updated_at"`
	DeletedAt         gorm.DeletedAt          `gorm:"index"                                    json:"-"`

	// Relations (preload only — no FK constraint enforced by GORM tags here)
	Session *JadwalUtama  `gorm:"foreignKey:IDSession" json:"session,omitempty"`
	Ruangan *Ruangan      `gorm:"foreignKey:IDRuangan" json:"ruangan,omitempty"`
	Dosen   *Dosen        `gorm:"foreignKey:IDDosen"   json:"dosen,omitempty"`
	Asdos1  *AsistenDosen `gorm:"foreignKey:IDAsdos1"  json:"asdos1,omitempty"`
	Asdos2  *AsistenDosen `gorm:"foreignKey:IDAsdos2"  json:"asdos2,omitempty"`
}

func init() {
	RegisterModel(&SubstituteSession{})
}

func (s *SubstituteSession) BeforeCreate(tx *gorm.DB) (err error) {
	s.ID = utils.GenerateCustomID("sub", 6)
	s.Status = StatusPending
	return nil
}
