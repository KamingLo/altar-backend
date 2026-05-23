package services

import (
	"altar/config"
	"altar/models"
	"time"
)

type RekapPresensiResponse struct {
	TotalHadir      int  `json:"total_hadir"`
	TotalTidakHadir int  `json:"total_tidak_hadir"`
	IsPaid          bool `json:"is_paid"`
}

type AsdosRekapResponse struct {
	AsdosID string                `json:"id_asdos"`
	Nama    string                `json:"nama"`
	Rekap   RekapPresensiResponse `json:"rekap"`
}

func GetAllRekapPresensi(startDate time.Time, endDate time.Time) ([]AsdosRekapResponse, error) {
	var results []AsdosRekapResponse
	var asdosList []models.AsistenDosen

	// Ambil semua data asisten dosen beserta usernamenya
	if err := config.DB.Preload("User").Find(&asdosList).Error; err != nil {
		return nil, err
	}

	for _, asdos := range asdosList {
		rekap, err := GetRekapPresensi(asdos.ID, startDate, endDate)
		if err != nil {
			return nil, err
		}

		results = append(results, AsdosRekapResponse{
			AsdosID: asdos.ID,
			Nama:    asdos.User.Username,
			Rekap:   rekap,
		})
	}

	return results, nil
}

func GetRekapPresensi(asdosID string, startDate time.Time, endDate time.Time) (RekapPresensiResponse, error) {
	var rekap RekapPresensiResponse
	db := config.DB

	// 1. Ambil data Presensi dalam rentang waktu
	var presensiList []models.Presensi
	err := db.Where("(id_asdos_pelaksana = ? OR id_asdos_rekan = ?) AND tanggal_mengajar BETWEEN ? AND ?",
		asdosID, asdosID, startDate, endDate).
		Find(&presensiList).Error
	if err != nil {
		return rekap, err
	}

	// Hitung TotalHadir
	rekap.TotalHadir = len(presensiList)

	// Map untuk pengecekan kehadiran (IDSesi + Tanggal)
	kehadiranMap := make(map[string]bool)
	hasUnpaid := false
	for _, p := range presensiList {
		dateKey := p.TanggalMengajar.Format("2006-01-02")
		kehadiranMap[p.IDSesi+"_"+dateKey] = true

		if !p.IsPaid {
			hasUnpaid = true
		}
	}

	// Tentukan status IsPaid siklus
	if rekap.TotalHadir > 0 && !hasUnpaid {
		rekap.IsPaid = true
	} else {
		rekap.IsPaid = false
	}

	// 2. Kalkulasi TotalTidakHadir
	// Tarik semua Jadwal Utama asdos ini
	var jadwalList []models.JadwalUtama
	err = db.Where("id_asdos1 = ? OR id_asdos2 = ?", asdosID, asdosID).Find(&jadwalList).Error
	if err != nil {
		return rekap, err
	}

	// Batas pengecekan: tidak boleh melebihi 'hari ini'
	now := time.Now().In(wib)
	limitDate := endDate
	if limitDate.After(now) {
		limitDate = now
	}

	// Looping per hari dari startDate sampai limitDate
	for d := startDate; !d.After(limitDate); d = d.AddDate(0, 0, 1) {
		targetDow := int(d.Weekday()) // 0=Sunday, 1=Mon, ..., 6=Sat

		for _, j := range jadwalList {
			// Cek apakah hari ini jadwal mengajar
			if int(j.KelasMulai.In(wib).Weekday()) != targetDow {
				continue
			}

			dateKey := d.Format("2006-01-02")
			
			// Jika sudah ada record presensi, skip
			if kehadiranMap[j.ID+"_"+dateKey] {
				continue
			}

			// Jika waktu kelas sudah lewat dan tidak ada presensi
			classTime := time.Date(
				d.Year(), d.Month(), d.Day(),
				j.KelasMulai.In(wib).Hour(), j.KelasMulai.In(wib).Minute(), 0, 0, wib,
			)

			if now.After(classTime) {
				rekap.TotalTidakHadir++
			}
		}
	}

	return rekap, nil
}
