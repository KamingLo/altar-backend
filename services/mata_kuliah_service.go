package services

import (
	"altar/config"
	"altar/models"
)

type CourseSummary struct {
	ID     string `json:"id"`
	NamaMK string `json:"nama_mk"`
	SKS    int    `json:"sks"`
}

func CreateCourse(course *models.MataKuliah) error {
	return config.DB.Create(course).Error
}

func GetAllCourses(page int, limit int, search string) ([]CourseSummary, int64, error) {
	var courses []CourseSummary
	var total int64

	offset := (page - 1) * limit
	query := config.DB.Model(&models.MataKuliah{})

	if search != "" {
		query = query.Where("nama_mk ILIKE ?", "%"+search+"%")
	}

	err := query.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	err = query.Limit(limit).Offset(offset).Find(&courses).Error
	if err != nil {
		return nil, 0, err
	}

	return courses, total, nil
}

func GetCourseByID(id string) (models.MataKuliah, error) {
	var course models.MataKuliah
	err := config.DB.First(&course, "id = ?", id).Error
	return course, err
}

func UpdateCourse(id string, data map[string]interface{}) (models.MataKuliah, error) {
	var course models.MataKuliah
	if err := config.DB.First(&course, "id = ?", id).Error; err != nil {
		return course, err
	}

	err := config.DB.Model(&course).Updates(data).Error
	return course, err
}

func DeleteCourse(id string) error {
	return config.DB.Delete(&models.MataKuliah{}, "id = ?", id).Error
}
