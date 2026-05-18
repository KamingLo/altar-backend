package controllers

import (
	"altar/models"
	"altar/services"
	"altar/utils"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type CheckInRequest struct {
	IDSesi          string  `json:"id_sesi" binding:"required"`
	IDSesiPengganti *string `json:"id_sesi_pengganti"`
	IDAsdosRekan    *string `json:"id_asdos_rekan"`
	Menggantikan    bool    `json:"menggantikan"`
	QRToken         string  `json:"qr_token" binding:"required"`
}

type CheckOutRequest struct {
	IDPresensi      string `json:"id_presensi" binding:"required"`
	DeskripsiMateri string `json:"deskripsi_materi" binding:"required"`
	QRToken         string `json:"qr_token" binding:"required"`
}

type EveningAttendanceRequest struct {
	IDSesi          string  `json:"id_sesi" binding:"required"`
	IDSesiPengganti *string `json:"id_sesi_pengganti"`
	IDAsdosRekan    *string `json:"id_asdos_rekan"`
	Menggantikan    bool    `json:"menggantikan"`
	WaktuMulai      string  `json:"waktu_mulai" binding:"required"`   // Format "15:04"
	WaktuSelesai    string  `json:"waktu_selesai" binding:"required"` // Format "15:04"
	DeskripsiMateri string  `json:"deskripsi_materi" binding:"required"`
}

type PresensiResponseDTO struct {
	IDPresensi      string     `json:"id_presensi"`
	IDSesi          string     `json:"id_sesi"`
	IDSesiPengganti *string    `json:"id_sesi_pengganti,omitempty"`
	NamaMataKuliah  string     `json:"nama_mata_kuliah"`
	NamaKelas       string     `json:"nama_kelas"`
	NamaRuangan     string     `json:"nama_ruangan"`
	NamaAsdos       string     `json:"nama_asdos"`
	NamaAsdosRekan  *string    `json:"nama_asdos_rekan,omitempty"`
	WaktuCheckIn    time.Time  `json:"waktu_checkin"`
	WaktuCheckOut   *time.Time `json:"waktu_checkout,omitempty"`
	TanggalMengajar time.Time  `json:"tanggal_mengajar"`
	DeskripsiMateri *string    `json:"deskripsi_materi,omitempty"`
	TipeAbsensi     string     `json:"tipe_absensi"`
	IsVerified      bool       `json:"is_verified"`
	Menggantikan    bool       `json:"menggantikan"`
	LinkVideo       *string    `json:"link_video,omitempty"`
}

func mapToPresensiDTO(p models.Presensi) PresensiResponseDTO {
	dto := PresensiResponseDTO{
		IDPresensi:      p.IDPresensi,
		IDSesi:          p.IDSesi,
		IDSesiPengganti: p.IDSesiPengganti,
		WaktuCheckIn:    p.WaktuCheckIn,
		WaktuCheckOut:   p.WaktuCheckOut,
		TanggalMengajar: p.TanggalMengajar,
		DeskripsiMateri: p.DeskripsiMateri,
		TipeAbsensi:     string(p.TipeAbsensi),
		IsVerified:      p.IsVerified,
		Menggantikan:    p.Menggantikan,
		LinkVideo:       p.LinkVideo,
		NamaAsdos:       p.AsdosPelaksana.User.Username,
	}

	if p.AsdosRekan != nil {
		dto.NamaAsdosRekan = &p.AsdosRekan.User.Username
	}

	// Determine metadata from regular or substitute session
	if p.SubstituteSession != nil && p.SubstituteSession.Session != nil {
		dto.NamaMataKuliah = p.SubstituteSession.Session.MataKuliah.NamaMK
		dto.NamaKelas = p.SubstituteSession.Session.Kelas.NamaKelas
		dto.NamaRuangan = p.SubstituteSession.Ruangan.NamaRuangan
	} else {
		dto.NamaMataKuliah = p.JadwalUtama.MataKuliah.NamaMK
		dto.NamaKelas = p.JadwalUtama.Kelas.NamaKelas
		dto.NamaRuangan = p.JadwalUtama.Ruangan.NamaRuangan
	}

	return dto
}

func CheckIn(c *gin.Context) {
	// Validate Time Window (Morning: 07:30 - 17:10)
	// now := time.Now()
	// start := time.Date(now.Year(), now.Month(), now.Day(), 7, 30, 0, 0, now.Location())
	// end := time.Date(now.Year(), now.Month(), now.Day(), 17, 10, 0, 0, now.Location())

	// // if now.Before(start) || now.After(end) {
	// // 	utils.SendError(c, http.StatusForbidden, "Morning attendance is only available between 07:30 and 17:10", nil)
	// // 	return
	// // }

	var req CheckInRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.SendError(c, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	// Validate QR Token
	if _, err := services.ValidateQRToken(req.QRToken); err != nil {
		utils.SendError(c, http.StatusUnauthorized, "Invalid or expired QR Token", err)
		return
	}

	asdosID := c.GetString("id_asisten")
	presensi := models.Presensi{
		IDSesi:          req.IDSesi,
		IDSesiPengganti: req.IDSesiPengganti,
		IDAsdosRekan:    req.IDAsdosRekan,
		Menggantikan:    req.Menggantikan,
	}

	res, err := services.CheckIn(asdosID, presensi)
	if err != nil {
		utils.SendError(c, http.StatusInternalServerError, "Failed to check-in", err)
		return
	}

	utils.SendSuccess(c, http.StatusCreated, "Check-in successful", mapToPresensiDTO(res))
}

func CheckOut(c *gin.Context) {
	var req CheckOutRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.SendError(c, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	// Validate QR Token
	if _, err := services.ValidateQRToken(req.QRToken); err != nil {
		utils.SendError(c, http.StatusUnauthorized, "Invalid or expired QR Token", err)
		return
	}

	asdosID := c.GetString("id_asisten")
	res, err := services.CheckOut(asdosID, req.IDPresensi, req.DeskripsiMateri)
	if err != nil {
		utils.SendError(c, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	utils.SendSuccess(c, http.StatusOK, "Check-out successful", mapToPresensiDTO(res))
}

func EveningAttendance(c *gin.Context) {
	// Validate Time Window (Evening: 17:45 - 21:00)
	now := time.Now()
	start := time.Date(now.Year(), now.Month(), now.Day(), 17, 45, 0, 0, now.Location())
	end := time.Date(now.Year(), now.Month(), now.Day(), 21, 0, 0, 0, now.Location())

	if now.Before(start) || now.After(end) {
		utils.SendError(c, http.StatusForbidden, "Evening attendance is only available between 17:45 and 21:00", nil)
		return
	}

	var req EveningAttendanceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.SendError(c, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	// Parse manual times
	tMulai, err1 := time.Parse("15:04", req.WaktuMulai)
	tSelesai, err2 := time.Parse("15:04", req.WaktuSelesai)
	if err1 != nil || err2 != nil {
		utils.SendError(c, http.StatusBadRequest, "Invalid time format. Use HH:mm (e.g., 18:30)", nil)
		return
	}

	// Combine with today's date
	waktuMulai := time.Date(now.Year(), now.Month(), now.Day(), tMulai.Hour(), tMulai.Minute(), 0, 0, now.Location())
	waktuSelesai := time.Date(now.Year(), now.Month(), now.Day(), tSelesai.Hour(), tSelesai.Minute(), 0, 0, now.Location())

	asdosID := c.GetString("id_asisten")
	presensi := models.Presensi{
		IDSesi:          req.IDSesi,
		IDSesiPengganti: req.IDSesiPengganti,
		IDAsdosRekan:    req.IDAsdosRekan,
		Menggantikan:    req.Menggantikan,
		DeskripsiMateri: &req.DeskripsiMateri,
	}

	res, err := services.EveningAttendance(asdosID, presensi, waktuMulai, waktuSelesai)
	if err != nil {
		utils.SendError(c, http.StatusInternalServerError, "Failed to submit evening attendance", err)
		return
	}

	utils.SendSuccess(c, http.StatusCreated, "Evening attendance submitted successfully", mapToPresensiDTO(res))
}

func GetAllPresensi(c *gin.Context) {
	verifiedStr := c.Query("is_verified")
	tipe := c.Query("tipe_absensi")

	var isVerified *bool
	if verifiedStr != "" {
		v := verifiedStr == "true"
		isVerified = &v
	}

	var tipePtr *string
	if tipe != "" {
		tipePtr = &tipe
	}

	res, err := services.GetAllPresensi(isVerified, tipePtr)
	if err != nil {
		utils.SendError(c, http.StatusInternalServerError, "Failed to fetch attendance records", err)
		return
	}

	var dtos []PresensiResponseDTO
	for _, p := range res {
		dtos = append(dtos, mapToPresensiDTO(p))
	}

	utils.SendSuccess(c, http.StatusOK, "Attendance records fetched successfully", dtos)
}

func GetAllMyPresensi(c *gin.Context) {
	asdosID := c.GetString("id_asisten")
	res, err := services.GetAllMyPresensi(asdosID)
	if err != nil {
		utils.SendError(c, http.StatusInternalServerError, "Failed to fetch your attendance records", err)
		return
	}

	var dtos []PresensiResponseDTO
	for _, p := range res {
		dtos = append(dtos, mapToPresensiDTO(p))
	}

	utils.SendSuccess(c, http.StatusOK, "Your attendance records fetched successfully", dtos)
}

type VerifyRequest struct {
	IsVerified bool `json:"is_verified"`
}

func VerifyPresensi(c *gin.Context) {
	id := c.Param("id")
	var req VerifyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.SendError(c, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	if err := services.VerifyPresensi(id, req.IsVerified); err != nil {
		utils.SendError(c, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	msg := "Attendance record verified"
	if !req.IsVerified {
		msg = "Attendance record unverified/rejected"
	}

	utils.SendSuccess(c, http.StatusOK, msg, nil)
}
