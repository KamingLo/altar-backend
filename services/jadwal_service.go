package services

import (
	"altar/config"
	"altar/models"
	"fmt"
)

func CreateSession(session *models.JadwalUtama) error {
	tx := config.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 1. Validate Mandatory Entity Existence
	if err := tx.First(&models.Kelas{}, "id = ?", session.IDKelas).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("class with ID %s not found", session.IDKelas)
	}
	if err := tx.First(&models.MataKuliah{}, "id = ?", session.IDMk).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("course with ID %s not found", session.IDMk)
	}
	if err := tx.First(&models.Semester{}, "id = ?", session.IDSemester).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("semester with ID %s not found", session.IDSemester)
	}
	if err := tx.First(&models.Ruangan{}, "id = ?", session.IDRuangan).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("room with ID %s not found", session.IDRuangan)
	}

	// 2. Validate Optional Entity Existence
	if session.IDDosen != nil {
		if err := tx.First(&models.Dosen{}, "id = ?", *session.IDDosen).Error; err != nil {
			tx.Rollback()
			return fmt.Errorf("lecturer with ID %s not found", *session.IDDosen)
		}
	}
	if session.IDAsdos1 != nil {
		if err := tx.First(&models.AsistenDosen{}, "id = ?", *session.IDAsdos1).Error; err != nil {
			tx.Rollback()
			return fmt.Errorf("assistant lecturer 1 with ID %s not found", *session.IDAsdos1)
		}
	}
	if session.IDAsdos2 != nil {
		if err := tx.First(&models.AsistenDosen{}, "id = ?", *session.IDAsdos2).Error; err != nil {
			tx.Rollback()
			return fmt.Errorf("assistant lecturer 2 with ID %s not found", *session.IDAsdos2)
		}
	}

	// 3. Clash Detection (Fail Fast)
	var count int64

	// A. Room
	tx.Model(&models.JadwalUtama{}).Where("id_ruangan = ? AND (kelas_mulai < ? AND kelas_berakhir > ?)", 
		session.IDRuangan, session.KelasBerakhir, session.KelasMulai).Count(&count)
	if count > 0 {
		tx.Rollback()
		return fmt.Errorf("room is already occupied during this time range")
	}

	// B. Lecturer
	if session.IDDosen != nil {
		tx.Model(&models.JadwalUtama{}).Where("id_dosen = ? AND (kelas_mulai < ? AND kelas_berakhir > ?)", 
			*session.IDDosen, session.KelasBerakhir, session.KelasMulai).Count(&count)
		if count > 0 {
			tx.Rollback()
			return fmt.Errorf("lecturer already has another session at this time")
		}
	}

	// C. Assistant Lecturer 1
	if session.IDAsdos1 != nil {
		tx.Model(&models.JadwalUtama{}).Where("(id_asdos1 = ? OR id_asdos2 = ?) AND (kelas_mulai < ? AND kelas_berakhir > ?)", 
			*session.IDAsdos1, *session.IDAsdos1, session.KelasBerakhir, session.KelasMulai).Count(&count)
		if count > 0 {
			tx.Rollback()
			return fmt.Errorf("assistant lecturer 1 already has another session at this time")
		}
	}

	// D. Assistant Lecturer 2
	if session.IDAsdos2 != nil {
		tx.Model(&models.JadwalUtama{}).Where("(id_asdos1 = ? OR id_asdos2 = ?) AND (kelas_mulai < ? AND kelas_berakhir > ?)", 
			*session.IDAsdos2, *session.IDAsdos2, session.KelasBerakhir, session.KelasMulai).Count(&count)
		if count > 0 {
			tx.Rollback()
			return fmt.Errorf("assistant lecturer 2 already has another session at this time")
		}
	}

	// 4. Create Session
	if err := tx.Create(session).Error; err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit().Error
}
