package services

import (
	"altar/config"
	"altar/models"
)

type SemesterSummary struct {
	ID           string `json:"id"`
	TahunAjaran  string `json:"tahun_ajaran"`
	TipeSemester string `json:"tipe_semester"`
}

func CreateSemester(semester *models.Semester) error {
	return config.DB.Create(semester).Error
}

func GetAllSemesters(page int, limit int, search string) ([]SemesterSummary, int64, error) {
	var semesters []SemesterSummary
	var total int64

	offset := (page - 1) * limit
	query := config.DB.Model(&models.Semester{})

	if search != "" {
		query = query.Where("tahun_ajaran ILIKE ? OR tipe_semester ILIKE ?", "%"+search+"%", "%"+search+"%")
	}

	err := query.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	err = query.Limit(limit).Offset(offset).Find(&semesters).Error
	if err != nil {
		return nil, 0, err
	}

	return semesters, total, nil
}

func GetSemesterByID(id string) (models.Semester, error) {
	var semester models.Semester
	err := config.DB.First(&semester, "id = ?", id).Error
	return semester, err
}

func UpdateSemester(id string, data map[string]interface{}) (models.Semester, error) {
	var semester models.Semester
	if err := config.DB.First(&semester, "id = ?", id).Error; err != nil {
		return semester, err
	}

	err := config.DB.Model(&semester).Updates(data).Error
	return semester, err
}

func DeleteSemester(id string) error {
	return config.DB.Delete(&models.Semester{}, "id = ?", id).Error
}
