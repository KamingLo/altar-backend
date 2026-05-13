package services

import (
	"altar/config"
	"altar/models"
	"errors"
	"fmt"
	"time"
)

// ─────────────────────────────────────────────
// DTO: SubstituteSessionInput (Create)
// ─────────────────────────────────────────────

type SubstituteSessionInput struct {
	IDSession      string `json:"id_session"`
	IDRuangan      string `json:"id_ruangan"`
	SubstituteDate string `json:"substitute_date"` // YYYY-MM-DD
	SlotOption     int    `json:"slot_option"`     // 1–7
	Reason         string `json:"reason"`
}

// ─────────────────────────────────────────────
// DTO: UpdateSubstituteStatusInput (Verify/Reject)
// ─────────────────────────────────────────────

type UpdateSubstituteStatusInput struct {
	Status models.SubstituteSessionStatus `json:"status"`
}

// ─────────────────────────────────────────────
// DTO: SubstituteSessionResponse
// ─────────────────────────────────────────────

type SubstituteSessionResponse struct {
	ID             string                         `json:"id"`
	Status         models.SubstituteSessionStatus `json:"status"`
	Reason         string                         `json:"reason"`
	SubstituteDate string                         `json:"substitute_date"` // YYYY-MM-DD
	TimeSlot       string                         `json:"time_slot"`       // "HH:mm – HH:mm"
	Room           string                         `json:"room"`            // "Name (Floor X)"
	Session        *SessionResponse               `json:"session,omitempty"`
	CreatedAt      time.Time                      `json:"created_at"`
	UpdatedAt      time.Time                      `json:"updated_at"`
}

// ─────────────────────────────────────────────
// Internal: translateSubstituteSchedule
// Converts a concrete date string (YYYY-MM-DD) + slot option into
// two WIB time.Time values (KelasMulai, KelasBerakhir).
// The date preserves the real calendar day of the substitute class.
// ─────────────────────────────────────────────

func translateSubstituteSchedule(dateStr string, slotOption int) (time.Time, time.Time, error) {
	type timeSlot struct{ startH, startM, endH, endM int }
	slotMap := map[int]timeSlot{
		1: {7, 30, 9, 10},
		2: {9, 30, 11, 10},
		3: {11, 30, 13, 10},
		4: {13, 30, 15, 10},
		5: {15, 30, 17, 10},
		6: {17, 40, 19, 15},
		7: {19, 30, 21, 0},
	}
	slot, ok := slotMap[slotOption]
	if !ok {
		return time.Time{}, time.Time{}, fmt.Errorf("invalid slot option: %d (valid: 1–7)", slotOption)
	}

	date, err := time.ParseInLocation("2006-01-02", dateStr, wib)
	if err != nil {
		return time.Time{}, time.Time{}, fmt.Errorf("invalid substitute_date format, expected YYYY-MM-DD: %w", err)
	}

	startTime := time.Date(date.Year(), date.Month(), date.Day(), slot.startH, slot.startM, 0, 0, wib)
	endTime := time.Date(date.Year(), date.Month(), date.Day(), slot.endH, slot.endM, 0, 0, wib)
	return startTime, endTime, nil
}

// ─────────────────────────────────────────────
// Internal: checkSubstituteClash
// Detects room conflicts across both JadwalUtama and SubstituteSession.
// For JadwalUtama: checks weekly time-of-day overlap regardless of week.
// For SubstituteSession: checks only PENDING or VERIFIED entries (same room + exact date).
// ─────────────────────────────────────────────

func checkSubstituteClash(roomID string, startTime, endTime time.Time, excludeID string) error {
	db := config.DB
	var count int64

	// A. Conflict against existing main sessions (JadwalUtama).
	//    JadwalUtama stores anchor dates in January 2024 but the time-of-day is
	//    what matters for a weekly repeating schedule. We compare the week day
	//    and the time window by checking the stored kelas_mulai/kelas_berakhir
	//    times that fall on the same weekday AND overlap in clock time.
	//    Simplest portable approach: compare weekday number and clock-time overlap.
	weekday := int(startTime.Weekday()) // 0=Sun…6=Sat; JadwalUtama day 1=Mon → weekday 1
	if err := db.Model(&models.JadwalUtama{}).
		Where("id_ruangan = ?", roomID).
		Where("DAYOFWEEK(kelas_mulai) = ?", weekday+1). // MySQL: 1=Sun, 2=Mon…
		Where("TIME(kelas_mulai) < TIME(?) AND TIME(kelas_berakhir) > TIME(?)",
			endTime, startTime).
		Count(&count).Error; err != nil {
		return fmt.Errorf("failed to check main session room conflict: %w", err)
	}
	if count > 0 {
		return errors.New("room is already used by a regular session on this weekday and time slot")
	}

	// B. Conflict against other SubstituteSession entries (PENDING or VERIFIED only).
	query := db.Model(&models.SubstituteSession{}).
		Where("id_ruangan = ?", roomID).
		Where("status IN (?)", []models.SubstituteSessionStatus{models.StatusPending, models.StatusVerified}).
		Where("kelas_mulai < ? AND kelas_berakhir > ?", endTime, startTime)
	if excludeID != "" {
		query = query.Where("id != ?", excludeID)
	}
	count = 0
	if err := query.Count(&count).Error; err != nil {
		return fmt.Errorf("failed to check substitute session room conflict: %w", err)
	}
	if count > 0 {
		return errors.New("room already has a pending or verified substitute session that overlaps with this time slot")
	}

	return nil
}

// ─────────────────────────────────────────────
// Internal: buildSubstituteResponse
// Builds the response DTO from a fully preloaded SubstituteSession.
// ─────────────────────────────────────────────

func buildSubstituteResponse(sub *models.SubstituteSession) SubstituteSessionResponse {
	roomStr := ""
	if sub.Ruangan != nil {
		roomStr = fmt.Sprintf("%s (Lantai %d)", sub.Ruangan.NamaRuangan, sub.Ruangan.Lantai)
	}

	timeSlot := fmt.Sprintf("%s – %s",
		sub.KelasMulai.In(wib).Format("15:04"),
		sub.KelasBerakhir.In(wib).Format("15:04"),
	)

	dateStr := sub.SubstituteDate.In(wib).Format("2006-01-02")

	var sessionResp *SessionResponse
	if sub.Session != nil {
		r, err := loadSessionRelations(sub.Session)
		if err == nil {
			sessionResp = &r
		}
	}

	return SubstituteSessionResponse{
		ID:             sub.ID,
		Status:         sub.Status,
		Reason:         sub.Reason,
		SubstituteDate: dateStr,
		TimeSlot:       timeSlot,
		Room:           roomStr,
		Session:        sessionResp,
		CreatedAt:      sub.CreatedAt,
		UpdatedAt:      sub.UpdatedAt,
	}
}

// ─────────────────────────────────────────────
// Service: CreateSubstituteSession
// Validates room availability, then saves a new PENDING substitute request.
// ─────────────────────────────────────────────

func CreateSubstituteSession(input *SubstituteSessionInput) (SubstituteSessionResponse, error) {
	db := config.DB

	// 1. Ensure the referenced session exists
	var session models.JadwalUtama
	if err := db.First(&session, "id = ?", input.IDSession).Error; err != nil {
		return SubstituteSessionResponse{}, fmt.Errorf("session with ID %s not found", input.IDSession)
	}

	// 2. Ensure the room exists
	var ruangan models.Ruangan
	if err := db.First(&ruangan, "id = ?", input.IDRuangan).Error; err != nil {
		return SubstituteSessionResponse{}, fmt.Errorf("room with ID %s not found", input.IDRuangan)
	}

	// 3. Translate date + slot into concrete time.Time values
	startTime, endTime, err := translateSubstituteSchedule(input.SubstituteDate, input.SlotOption)
	if err != nil {
		return SubstituteSessionResponse{}, err
	}

	// 4. Clash detection against JadwalUtama and other SubstituteSessions
	if err := checkSubstituteClash(input.IDRuangan, startTime, endTime, ""); err != nil {
		return SubstituteSessionResponse{}, err
	}

	// 5. Parse substitute date (date-only, midnight WIB)
	substituteDate, _ := time.ParseInLocation("2006-01-02", input.SubstituteDate, wib)

	// 6. Persist
	sub := models.SubstituteSession{
		IDSession:     input.IDSession,
		IDRuangan:     input.IDRuangan,
		Reason:        input.Reason,
		Status:        models.StatusPending,
		SubstituteDate: substituteDate,
		KelasMulai:    startTime,
		KelasBerakhir: endTime,
	}
	if err := db.Create(&sub).Error; err != nil {
		return SubstituteSessionResponse{}, fmt.Errorf("failed to create substitute session: %w", err)
	}

	// 7. Reload with preloads for the response
	if err := db.Preload("Session").Preload("Session.Kelas").
		Preload("Session.MataKuliah").Preload("Ruangan").
		First(&sub, "id = ?", sub.ID).Error; err != nil {
		return SubstituteSessionResponse{}, fmt.Errorf("failed to reload substitute session: %w", err)
	}

	return buildSubstituteResponse(&sub), nil
}

// ─────────────────────────────────────────────
// Service: GetAllSubstituteSessions
// Returns all substitute sessions, optionally filtered by status.
// ─────────────────────────────────────────────

func GetAllSubstituteSessions(statusFilter string) ([]SubstituteSessionResponse, error) {
	db := config.DB
	var subs []models.SubstituteSession

	query := db.Preload("Session").Preload("Session.Kelas").
		Preload("Session.MataKuliah").Preload("Ruangan")

	if statusFilter != "" {
		query = query.Where("status = ?", statusFilter)
	}

	if err := query.Order("created_at DESC").Find(&subs).Error; err != nil {
		return nil, fmt.Errorf("failed to fetch substitute sessions: %w", err)
	}

	responses := make([]SubstituteSessionResponse, 0, len(subs))
	for i := range subs {
		responses = append(responses, buildSubstituteResponse(&subs[i]))
	}
	return responses, nil
}

// ─────────────────────────────────────────────
// Service: UpdateSubstituteStatus
// Coordinator approves (VERIFIED) or rejects (REJECTED) a pending request.
// ─────────────────────────────────────────────

func UpdateSubstituteStatus(id string, input *UpdateSubstituteStatusInput) (SubstituteSessionResponse, error) {
	db := config.DB

	// Validate allowed transitions
	if input.Status != models.StatusVerified && input.Status != models.StatusRejected {
		return SubstituteSessionResponse{}, errors.New("status must be either 'VERIFIED' or 'REJECTED'")
	}

	var sub models.SubstituteSession
	if err := db.Preload("Session").Preload("Session.Kelas").
		Preload("Session.MataKuliah").Preload("Ruangan").
		First(&sub, "id = ?", id).Error; err != nil {
		return SubstituteSessionResponse{}, fmt.Errorf("substitute session with ID %s not found", id)
	}

	if sub.Status == models.StatusVerified || sub.Status == models.StatusRejected {
		return SubstituteSessionResponse{}, fmt.Errorf(
			"substitute session is already finalised with status '%s' and cannot be changed", sub.Status,
		)
	}

	if err := db.Model(&sub).Update("status", input.Status).Error; err != nil {
		return SubstituteSessionResponse{}, fmt.Errorf("failed to update substitute session status: %w", err)
	}

	// Reload to reflect updated_at
	if err := db.Preload("Session").Preload("Session.Kelas").
		Preload("Session.MataKuliah").Preload("Ruangan").
		First(&sub, "id = ?", id).Error; err != nil {
		return SubstituteSessionResponse{}, fmt.Errorf("failed to reload substitute session after update: %w", err)
	}

	return buildSubstituteResponse(&sub), nil
}
