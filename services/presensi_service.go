package services

import (
	"altar/config"
	"altar/models"
	"altar/utils"
	"errors"
	"os"
	"time"
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
	input.IsVerified = true // Auto verified if via QR

	if err := config.DB.Create(&input).Error; err != nil {
		return models.Presensi{}, err
	}

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

	return presensi, nil
}

func EveningAttendance(asdosID string, input models.Presensi, startTime, endTime time.Time) (models.Presensi, error) {
	input.IDAsdosPelaksana = asdosID
	input.WaktuCheckIn = startTime
	input.WaktuCheckOut = &endTime
	input.TanggalMengajar = time.Now()
	input.TipeAbsensi = models.AbsensiLink
	input.IsVerified = false // Link based needs coordinator verification

	if err := config.DB.Create(&input).Error; err != nil {
		return models.Presensi{}, err
	}

	return input, nil
}

func GetAllPresensi(isVerified *bool, tipe *string) ([]models.Presensi, error) {
	var presensi []models.Presensi
	query := config.DB.Preload("JadwalUtama").Preload("AsdosPelaksana").Preload("AsdosRekan")

	if isVerified != nil {
		query = query.Where("is_verified = ?", *isVerified)
	}
	if tipe != nil {
		query = query.Where("tipe_absensi = ?", *tipe)
	}

	if err := query.Find(&presensi).Error; err != nil {
		return nil, err
	}
	return presensi, nil
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
