package controllers

import (
	"altar/models"
	"altar/services"
	"altar/utils"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type RoomRequest struct {
	NamaRuangan string `json:"nama_ruangan" binding:"required"`
	Lantai      int    `json:"lantai" binding:"required"`
	Kapasitas   int    `json:"kapasitas" binding:"required"`
}

func CreateRoom(c *gin.Context) {
	var req RoomRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.SendError(c, http.StatusBadRequest, "Invalid request payload", err)
		return
	}

	room := models.Ruangan{
		NamaRuangan: req.NamaRuangan,
		Lantai:      req.Lantai,
		Kapasitas:   req.Kapasitas,
	}

	if err := services.CreateRoom(&room); err != nil {
		utils.SendError(c, http.StatusInternalServerError, "Failed to create room", err)
		return
	}

	utils.SendSuccess(c, http.StatusCreated, "Room created successfully", room)
}

func GetAllRooms(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	search := c.Query("search")

	data, total, err := services.GetAllRooms(page, limit, search)
	if err != nil {
		utils.SendError(c, http.StatusInternalServerError, "Failed to fetch rooms", err)
		return
	}

	utils.SendSuccess(c, http.StatusOK, "Rooms fetched successfully", gin.H{
		"items":      data,
		"total":      total,
		"page":       page,
		"limit":      limit,
		"total_page": (total + int64(limit) - 1) / int64(limit),
	})
}

func GetRoomByID(c *gin.Context) {
	id := c.Param("id")
	room, err := services.GetRoomByID(id)
	if err != nil {
		utils.SendError(c, http.StatusNotFound, "Room not found", err)
		return
	}

	utils.SendSuccess(c, http.StatusOK, "Room fetched successfully", room)
}

func UpdateRoom(c *gin.Context) {
	id := c.Param("id")
	var req RoomRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.SendError(c, http.StatusBadRequest, "Invalid request payload", err)
		return
	}

	updateData := map[string]interface{}{
		"nama_ruangan": req.NamaRuangan,
		"lantai":       req.Lantai,
		"kapasitas":    req.Kapasitas,
	}

	room, err := services.UpdateRoom(id, updateData)
	if err != nil {
		utils.SendError(c, http.StatusInternalServerError, "Failed to update room", err)
		return
	}

	utils.SendSuccess(c, http.StatusOK, "Room updated successfully", room)
}

func DeleteRoom(c *gin.Context) {
	id := c.Param("id")
	if err := services.DeleteRoom(id); err != nil {
		utils.SendError(c, http.StatusInternalServerError, "Failed to delete room", err)
		return
	}

	utils.SendSuccess(c, http.StatusOK, "Room deleted successfully", nil)
}
