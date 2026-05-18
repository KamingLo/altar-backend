package services

import (
	"altar/config"
	"altar/models"
)

type ClassSummary struct {
	ID          string `json:"id"`
	NamaKelas   string `json:"nama_kelas"`
	Jurusan     string `json:"jurusan"`
	JumlahSiswa int    `json:"jumlah_siswa"`
}

func CreateClass(class *models.Kelas) error {
	return config.DB.Create(class).Error
}

func GetAllClasses(page int, limit int, search string) ([]ClassSummary, int64, error) {
	var classes []ClassSummary
	var total int64

	offset := (page - 1) * limit
	query := config.DB.Model(&models.Kelas{})

	if search != "" {
		query = query.Where("nama_kelas ILIKE ? OR jurusan ILIKE ?", "%"+search+"%", "%"+search+"%")
	}

	err := query.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	err = query.Limit(limit).Offset(offset).Find(&classes).Error
	if err != nil {
		return nil, 0, err
	}

	return classes, total, nil
}

func GetClassByID(id string) (models.Kelas, error) {
	var class models.Kelas
	err := config.DB.First(&class, "id = ?", id).Error
	return class, err
}

func UpdateClass(id string, data map[string]interface{}) (models.Kelas, error) {
	var class models.Kelas
	if err := config.DB.First(&class, "id = ?", id).Error; err != nil {
		return class, err
	}

	err := config.DB.Model(&class).Updates(data).Error
	return class, err
}

func DeleteClass(id string) error {
	return config.DB.Delete(&models.Kelas{}, "id = ?", id).Error
}
