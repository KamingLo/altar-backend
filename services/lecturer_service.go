package services

import (
	"altar/config"
	"altar/models"
	"errors"
)

// LecturerSummary menambahkan tag gorm:"column:nip" agar fitur pencarian (GET All) berjalan mulus
type LecturerSummary struct {
	ID   string `json:"id"`
	NIP  string `gorm:"column:nip" json:"nip"`
	Nama string `json:"nama"`
}

func CreateLecturer(lecturer *models.Dosen) error {
	// Pengecekan NIP duplikat agar error lebih ramah untuk frontend
	var existing models.Dosen
	if err := config.DB.Where("nip = ?", lecturer.NIP).First(&existing).Error; err == nil {
		return errors.New("NIP sudah terdaftar di sistem")
	}

	return config.DB.Create(lecturer).Error
}

func GetAllLecturers(page int, limit int, search string) ([]LecturerSummary, int64, error) {
	var lecturers []LecturerSummary
	var total int64

	offset := (page - 1) * limit
	query := config.DB.Model(&models.Dosen{})

	if search != "" {
		query = query.Where("nip ILIKE ? OR nama ILIKE ?", "%"+search+"%", "%"+search+"%")
	}

	err := query.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	err = query.Limit(limit).Offset(offset).Find(&lecturers).Error
	if err != nil {
		return nil, 0, err
	}

	return lecturers, total, nil
}

func GetLecturerByID(id string) (models.Dosen, error) {
	var lecturer models.Dosen
	err := config.DB.First(&lecturer, "id = ?", id).Error
	return lecturer, err
}

func UpdateLecturer(id string, data map[string]interface{}) (models.Dosen, error) {
	var lecturer models.Dosen

	// 1. Pastikan dosen yang ingin diupdate memang ada
	if err := config.DB.First(&lecturer, "id = ?", id).Error; err != nil {
		return lecturer, errors.New("dosen tidak ditemukan")
	}

	// 2. Jika payload update mengandung perubahan NIP, cek apakah NIP tersebut bentrok dengan dosen LAIN
	if newNIP, exists := data["nip"].(string); exists {
		var existing models.Dosen
		if err := config.DB.Where("nip = ? AND id != ?", newNIP, id).First(&existing).Error; err == nil {
			return lecturer, errors.New("NIP sudah terdaftar pada dosen lain")
		}
	}

	// 3. Lakukan update parsial menggunakan map
	err := config.DB.Model(&lecturer).Updates(data).Error
	return lecturer, err
}

func DeleteLecturer(id string) error {
	return config.DB.Delete(&models.Dosen{}, "id = ?", id).Error
}
