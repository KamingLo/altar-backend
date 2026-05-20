package services

import (
	"altar/config"
	"altar/models"
)

type RoomSummary struct {
	ID          string `json:"id"`
	NamaRuangan string `json:"nama_ruangan"`
	Lantai      int    `json:"lantai"`
	Kapasitas   int    `json:"kapasitas"`
}

func CreateRoom(room *models.Ruangan) error {
	return config.DB.Create(room).Error
}

func GetAllRooms(page int, limit int, search string) ([]RoomSummary, int64, error) {
	var rooms []RoomSummary
	var total int64

	offset := (page - 1) * limit
	query := config.DB.Model(&models.Ruangan{})

	if search != "" {
		query = query.Where("nama_ruangan ILIKE ?", "%"+search+"%")
	}

	err := query.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	err = query.Limit(limit).Offset(offset).Find(&rooms).Error
	if err != nil {
		return nil, 0, err
	}

	return rooms, total, nil
}

func GetRoomByID(id string) (models.Ruangan, error) {
	var room models.Ruangan
	err := config.DB.First(&room, "id = ?", id).Error
	return room, err
}

func UpdateRoom(id string, data map[string]interface{}) (models.Ruangan, error) {
	var room models.Ruangan
	if err := config.DB.First(&room, "id = ?", id).Error; err != nil {
		return room, err
	}

	err := config.DB.Model(&room).Updates(data).Error
	return room, err
}

func DeleteRoom(id string) error {
	return config.DB.Delete(&models.Ruangan{}, "id = ?", id).Error
}
