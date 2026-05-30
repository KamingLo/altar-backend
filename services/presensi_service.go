package services

import (
	"altar/config"
	"altar/models"
	"altar/utils"
	"errors"
	"os"
	"time"

	"gorm.io/gorm"
)

func ValidateQRToken(tokenString string) (string, error) {
	claims, err := utils.ValidateToken(tokenString, os.Getenv("JWT_SECRET"))
	if err != nil {
		return "", err
	}

	coordinatorID, ok := claims["coordinator_id"].(string)
	if !ok {
		return "", errors.New("invalid QR token payload")
	}

	return coordinatorID, nil
}

func CheckIn(asdosID string, input models.Presensi) (models.Presensi, error) {
	// Business Rules:
	// Regular: Menggantikan = false, IDSesi = ID Jadwal Utama, IDSesiPengganti = NULL
	// KP: Menggantikan = true, IDSesi = ID Jadwal Utama, IDSesiPengganti = KP ID

	input.IDAsdosPelaksana = asdosID
	input.WaktuCheckIn = time.Now()
	input.TanggalMengajar = time.Now()
	input.TipeAbsensi = models.AbsensiQR

	if err := config.DB.Create(&input).Error; err != nil {
		return models.Presensi{}, err
	}

	// Reload with preloads
	config.DB.
		Preload("JadwalUtama", func(db *gorm.DB) *gorm.DB { return db.Unscoped() }).
		Preload("JadwalUtama.MataKuliah", func(db *gorm.DB) *gorm.DB { return db.Unscoped() }).
		Preload("JadwalUtama.Kelas", func(db *gorm.DB) *gorm.DB { return db.Unscoped() }).
		Preload("JadwalUtama.Ruangan", func(db *gorm.DB) *gorm.DB { return db.Unscoped() }).
		Preload("SubstituteSession", func(db *gorm.DB) *gorm.DB { return db.Unscoped() }).
		Preload("SubstituteSession.Session", func(db *gorm.DB) *gorm.DB { return db.Unscoped() }).
		Preload("SubstituteSession.Session.MataKuliah", func(db *gorm.DB) *gorm.DB { return db.Unscoped() }).
		Preload("SubstituteSession.Session.Kelas", func(db *gorm.DB) *gorm.DB { return db.Unscoped() }).
		Preload("SubstituteSession.Ruangan", func(db *gorm.DB) *gorm.DB { return db.Unscoped() }).
		Preload("AsdosPelaksana", func(db *gorm.DB) *gorm.DB { return db.Unscoped() }).
		Preload("AsdosPelaksana.User").
		Preload("AsdosRekan", func(db *gorm.DB) *gorm.DB { return db.Unscoped() }).
		Preload("AsdosRekan.User").
		First(&input, "id_presensi = ?", input.IDPresensi)

	return input, nil
}

func CheckOut(asdosID string, presensiID string, deskripsi string) (models.Presensi, error) {
	var presensi models.Presensi
	if err := config.DB.Where("id_presensi = ?", presensiID).First(&presensi).Error; err != nil {
		return models.Presensi{}, errors.New("attendance record not found")
	}

	if presensi.IDAsdosPelaksana != asdosID {
		return models.Presensi{}, errors.New("unauthorized: you are not the one who checked in")
	}

	if presensi.WaktuCheckOut != nil {
		return models.Presensi{}, errors.New("already checked out")
	}

	now := time.Now()
	presensi.WaktuCheckOut = &now
	presensi.DeskripsiMateri = &deskripsi

	if err := config.DB.Save(&presensi).Error; err != nil {
		return models.Presensi{}, err
	}

	// Reload with preloads
	config.DB.
		Preload("JadwalUtama", func(db *gorm.DB) *gorm.DB { return db.Unscoped() }).
		Preload("JadwalUtama.MataKuliah", func(db *gorm.DB) *gorm.DB { return db.Unscoped() }).
		Preload("JadwalUtama.Kelas", func(db *gorm.DB) *gorm.DB { return db.Unscoped() }).
		Preload("JadwalUtama.Ruangan", func(db *gorm.DB) *gorm.DB { return db.Unscoped() }).
		Preload("SubstituteSession", func(db *gorm.DB) *gorm.DB { return db.Unscoped() }).
		Preload("SubstituteSession.Session", func(db *gorm.DB) *gorm.DB { return db.Unscoped() }).
		Preload("SubstituteSession.Session.MataKuliah", func(db *gorm.DB) *gorm.DB { return db.Unscoped() }).
		Preload("SubstituteSession.Session.Kelas", func(db *gorm.DB) *gorm.DB { return db.Unscoped() }).
		Preload("SubstituteSession.Ruangan", func(db *gorm.DB) *gorm.DB { return db.Unscoped() }).
		Preload("AsdosPelaksana", func(db *gorm.DB) *gorm.DB { return db.Unscoped() }).
		Preload("AsdosPelaksana.User").
		Preload("AsdosRekan", func(db *gorm.DB) *gorm.DB { return db.Unscoped() }).
		Preload("AsdosRekan.User").
		First(&presensi, "id_presensi = ?", presensi.IDPresensi)

	return presensi, nil
}

func OnlineAttendance(asdosID string, input models.Presensi, startTime, endTime time.Time) (models.Presensi, error) {
	input.IDAsdosPelaksana = asdosID
	input.WaktuCheckIn = startTime
	input.WaktuCheckOut = &endTime
	input.TanggalMengajar = time.Now()
	input.TipeAbsensi = models.AbsensiLink
	input.IsVerified = false // Link based needs coordinator verification

	if err := config.DB.Create(&input).Error; err != nil {
		return models.Presensi{}, err
	}

	// Reload with preloads
	config.DB.
		Preload("JadwalUtama", func(db *gorm.DB) *gorm.DB { return db.Unscoped() }).
		Preload("JadwalUtama.MataKuliah", func(db *gorm.DB) *gorm.DB { return db.Unscoped() }).
		Preload("JadwalUtama.Kelas", func(db *gorm.DB) *gorm.DB { return db.Unscoped() }).
		Preload("JadwalUtama.Ruangan", func(db *gorm.DB) *gorm.DB { return db.Unscoped() }).
		Preload("SubstituteSession", func(db *gorm.DB) *gorm.DB { return db.Unscoped() }).
		Preload("SubstituteSession.Session", func(db *gorm.DB) *gorm.DB { return db.Unscoped() }).
		Preload("SubstituteSession.Session.MataKuliah", func(db *gorm.DB) *gorm.DB { return db.Unscoped() }).
		Preload("SubstituteSession.Session.Kelas", func(db *gorm.DB) *gorm.DB { return db.Unscoped() }).
		Preload("SubstituteSession.Ruangan", func(db *gorm.DB) *gorm.DB { return db.Unscoped() }).
		Preload("AsdosPelaksana", func(db *gorm.DB) *gorm.DB { return db.Unscoped() }).
		Preload("AsdosPelaksana.User").
		Preload("AsdosRekan", func(db *gorm.DB) *gorm.DB { return db.Unscoped() }).
		Preload("AsdosRekan.User").
		First(&input, "id_presensi = ?", input.IDPresensi)

	return input, nil
}

func GetAllPresensi(isVerified *bool, tipe *string, idUser *string, idSemester *string, isPaid *bool) ([]models.Presensi, error) {
	var presensi []models.Presensi
	query := config.DB.
		Preload("JadwalUtama", func(db *gorm.DB) *gorm.DB { return db.Unscoped() }).
		Preload("JadwalUtama.MataKuliah", func(db *gorm.DB) *gorm.DB { return db.Unscoped() }).
		Preload("JadwalUtama.Kelas", func(db *gorm.DB) *gorm.DB { return db.Unscoped() }).
		Preload("JadwalUtama.Ruangan", func(db *gorm.DB) *gorm.DB { return db.Unscoped() }).
		Preload("SubstituteSession", func(db *gorm.DB) *gorm.DB { return db.Unscoped() }).
		Preload("SubstituteSession.Session", func(db *gorm.DB) *gorm.DB { return db.Unscoped() }).
		Preload("SubstituteSession.Session.MataKuliah", func(db *gorm.DB) *gorm.DB { return db.Unscoped() }).
		Preload("SubstituteSession.Session.Kelas", func(db *gorm.DB) *gorm.DB { return db.Unscoped() }).
		Preload("SubstituteSession.Ruangan", func(db *gorm.DB) *gorm.DB { return db.Unscoped() }).
		Preload("AsdosPelaksana", func(db *gorm.DB) *gorm.DB { return db.Unscoped() }).
		Preload("AsdosPelaksana.User").
		Preload("AsdosRekan", func(db *gorm.DB) *gorm.DB { return db.Unscoped() }).
		Preload("AsdosRekan.User")

	if isVerified != nil {
		query = query.Where("presensis.is_verified = ?", *isVerified)
	}
	if tipe != nil {
		query = query.Where("presensis.tipe_absensi = ?", *tipe)
	}
	if isPaid != nil {
		query = query.Where("presensis.is_paid = ?", *isPaid)
	}

	if idUser != nil {
		query = query.Joins("JOIN asisten_dosens ON presensis.id_asdos_pelaksana = asisten_dosens.id").
			Where("asisten_dosens.id = ?", *idUser)
	}

	if idSemester != nil {
		query = query.Joins("JOIN jadwal_utamas ON presensis.id_sesi = jadwal_utamas.id").
			Where("jadwal_utamas.id_semester = ?", *idSemester)
	}

	if err := query.Find(&presensi).Error; err != nil {
		return nil, err
	}
	return presensi, nil
}

func GetAllMyPresensi(asdosID string) ([]models.Presensi, error) {
	var presensi []models.Presensi
	err := config.DB.
		Preload("JadwalUtama", func(db *gorm.DB) *gorm.DB { return db.Unscoped() }).
		Preload("JadwalUtama.MataKuliah", func(db *gorm.DB) *gorm.DB { return db.Unscoped() }).
		Preload("JadwalUtama.Kelas", func(db *gorm.DB) *gorm.DB { return db.Unscoped() }).
		Preload("JadwalUtama.Ruangan", func(db *gorm.DB) *gorm.DB { return db.Unscoped() }).
		Preload("SubstituteSession", func(db *gorm.DB) *gorm.DB { return db.Unscoped() }).
		Preload("SubstituteSession.Session", func(db *gorm.DB) *gorm.DB { return db.Unscoped() }).
		Preload("SubstituteSession.Session.MataKuliah", func(db *gorm.DB) *gorm.DB { return db.Unscoped() }).
		Preload("SubstituteSession.Session.Kelas", func(db *gorm.DB) *gorm.DB { return db.Unscoped() }).
		Preload("SubstituteSession.Ruangan", func(db *gorm.DB) *gorm.DB { return db.Unscoped() }).
		Preload("AsdosPelaksana", func(db *gorm.DB) *gorm.DB { return db.Unscoped() }).
		Preload("AsdosPelaksana.User").
		Preload("AsdosRekan", func(db *gorm.DB) *gorm.DB { return db.Unscoped() }).
		Preload("AsdosRekan.User").
		Where("id_asdos_pelaksana = ? OR id_asdos_rekan = ?", asdosID, asdosID).
		Order("waktu_check_in DESC").
		Find(&presensi).Error

	return presensi, err
}

func VerifyPresensi(id string, verified bool) error {
	result := config.DB.Model(&models.Presensi{}).Where("id_presensi = ?", id).Update("is_verified", verified)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("attendance record not found")
	}
	return nil
}

func UpdatePaymentStatus(ids []string, isPaid bool) error {
	result := config.DB.Model(&models.Presensi{}).Where("id_presensi IN ?", ids).Update("is_paid", isPaid)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("no attendance records found to update")
	}
	return nil
}
