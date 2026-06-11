package services

import (
	"errors"
	"sort"
	"strconv"
	"strings"
	"time"

	"school-schedule-api/dto"
	"school-schedule-api/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ScheduleService struct {
	DB *gorm.DB
}

type ScheduleFilters struct {
	ClassCode  string
	TeacherNIK string
	Date       string
	StartDate  string
	EndDate    string
	Page       int
	Limit      int
}

func NewScheduleService(db *gorm.DB) *ScheduleService {
	return &ScheduleService{DB: db}
}

func ValidateScheduleRequest(req dto.ScheduleRequest) []string {
	var errs []string

	if strings.TrimSpace(req.ClassCode) == "" {
		errs = append(errs, "class_code is required")
	}
	if strings.TrimSpace(req.ClassName) == "" {
		errs = append(errs, "class_name is required")
	}
	if strings.TrimSpace(req.SubjectCode) == "" {
		errs = append(errs, "subject_code is required")
	}
	if strings.TrimSpace(req.TeacherNIK) == "" {
		errs = append(errs, "teacher_nik is required")
	}
	if strings.TrimSpace(req.TeacherName) == "" {
		errs = append(errs, "teacher_name is required")
	}
	if strings.TrimSpace(req.Date) == "" {
		errs = append(errs, "date is required")
	}
	if req.JamKe <= 0 {
		errs = append(errs, "jam_ke must be greater than 0")
	}
	if strings.TrimSpace(req.TimeStart) == "" {
		errs = append(errs, "time_start is required")
	}
	if strings.TrimSpace(req.TimeEnd) == "" {
		errs = append(errs, "time_end is required")
	}

	lengthChecks := map[string]struct {
		value string
		max   int
	}{
		"class_code":   {req.ClassCode, 10},
		"class_name":   {req.ClassName, 10},
		"subject_code": {req.SubjectCode, 10},
		"teacher_nik":  {req.TeacherNIK, 20},
		"teacher_name": {req.TeacherName, 100},
	}
	for field, check := range lengthChecks {
		if len(check.value) > check.max {
			errs = append(errs, field+" maximum length is "+strconv.Itoa(check.max))
		}
	}

	if req.Date != "" {
		if _, err := time.Parse("2006-01-02", req.Date); err != nil {
			errs = append(errs, "date must be YYYY-MM-DD")
		}
	}

	start, startErr := time.Parse("15:04:05", req.TimeStart)
	end, endErr := time.Parse("15:04:05", req.TimeEnd)
	if req.TimeStart != "" && startErr != nil {
		errs = append(errs, "time_start must be HH:mm:ss")
	}
	if req.TimeEnd != "" && endErr != nil {
		errs = append(errs, "time_end must be HH:mm:ss")
	}
	if startErr == nil && endErr == nil && !end.After(start) {
		errs = append(errs, "time_end must be greater than time_start")
	}

	return errs
}

func ValidateDateRange(startDate string, endDate string) []string {
	var errs []string
	start, startErr := time.Parse("2006-01-02", startDate)
	end, endErr := time.Parse("2006-01-02", endDate)

	if startDate == "" {
		errs = append(errs, "start_date is required")
	} else if startErr != nil {
		errs = append(errs, "start_date must be YYYY-MM-DD")
	}
	if endDate == "" {
		errs = append(errs, "end_date is required")
	} else if endErr != nil {
		errs = append(errs, "end_date must be YYYY-MM-DD")
	}
	if startErr == nil && endErr == nil && end.Before(start) {
		errs = append(errs, "end_date must not be earlier than start_date")
	}

	return errs
}

func RequestToSchedule(req dto.ScheduleRequest) models.Schedule {
	return models.Schedule{
		ClassCode:   strings.TrimSpace(req.ClassCode),
		ClassName:   strings.TrimSpace(req.ClassName),
		SubjectCode: strings.TrimSpace(req.SubjectCode),
		TeacherNIK:  strings.TrimSpace(req.TeacherNIK),
		TeacherName: strings.TrimSpace(req.TeacherName),
		Date:        strings.TrimSpace(req.Date),
		JamKe:       req.JamKe,
		TimeStart:   strings.TrimSpace(req.TimeStart),
		TimeEnd:     strings.TrimSpace(req.TimeEnd),
	}
}

func (s *ScheduleService) Create(req dto.ScheduleRequest) (models.Schedule, []string, error) {
	schedule := RequestToSchedule(req)
	conflicts, err := s.CheckConflicts(schedule, nil)
	if err != nil || len(conflicts) > 0 {
		return schedule, conflicts, err
	}

	return schedule, nil, s.DB.Create(&schedule).Error
}

func (s *ScheduleService) FindAll(filters ScheduleFilters) ([]models.Schedule, int64, error) {
	var schedules []models.Schedule
	var total int64

	query := s.applyFilters(s.DB.Model(&models.Schedule{}), filters)
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	page, limit := normalizePagination(filters.Page, filters.Limit)
	err := query.Order("date ASC, time_start ASC, jam_ke ASC").
		Limit(limit).
		Offset((page - 1) * limit).
		Find(&schedules).Error
	return schedules, total, err
}

func (s *ScheduleService) FindByID(id string) (models.Schedule, error) {
	var schedule models.Schedule
	parsedID, err := uuid.Parse(id)
	if err != nil {
		return schedule, gorm.ErrRecordNotFound
	}

	return schedule, s.DB.First(&schedule, "id = ?", parsedID).Error
}

func (s *ScheduleService) Update(id string, req dto.ScheduleRequest) (models.Schedule, []string, error) {
	existing, err := s.FindByID(id)
	if err != nil {
		return existing, nil, err
	}

	updated := RequestToSchedule(req)
	updated.ID = existing.ID
	conflicts, err := s.CheckConflicts(updated, &existing.ID)
	if err != nil || len(conflicts) > 0 {
		return existing, conflicts, err
	}

	err = s.DB.Model(&existing).Updates(updated).Error
	return existing, nil, err
}

func (s *ScheduleService) Delete(id string) error {
	schedule, err := s.FindByID(id)
	if err != nil {
		return err
	}

	return s.DB.Delete(&schedule).Error
}

func (s *ScheduleService) StudentSchedule(classCode string, date string) ([]models.Schedule, error) {
	var schedules []models.Schedule
	err := s.DB.Where("class_code = ? AND date = ?", classCode, date).
		Order("jam_ke ASC, time_start ASC").
		Find(&schedules).Error
	return schedules, err
}

func (s *ScheduleService) TeacherSchedule(teacherNIK string, startDate string, endDate string) ([]models.Schedule, error) {
	var schedules []models.Schedule
	err := s.DB.Where("teacher_nik = ? AND date BETWEEN ? AND ?", teacherNIK, startDate, endDate).
		Order("date ASC, time_start ASC, jam_ke ASC").
		Find(&schedules).Error
	return schedules, err
}

type foundationRecapRow struct {
	TeacherNIK  string
	TeacherName string
	ClassCode   string
	ClassName   string
	JumlahJP    int64
}

func (s *ScheduleService) RecapJP(startDate string, endDate string) (dto.FoundationRecapResponse, error) {
	response := dto.FoundationRecapResponse{
		Periode: dto.PeriodResponse{
			StartDate: startDate,
			EndDate:   endDate,
		},
		Rekap: []dto.FoundationRecapTeacher{},
	}

	var rows []foundationRecapRow
	err := s.DB.Model(&models.Schedule{}).
		Select("teacher_nik, teacher_name, class_code, class_name, COUNT(*) AS jumlah_jp").
		Where("date >= ? AND date <= ?", startDate, endDate).
		Group("teacher_nik, teacher_name, class_code, class_name").
		Order("teacher_name ASC, class_name ASC, class_code ASC").
		Scan(&rows).Error
	if err != nil {
		return response, err
	}

	teacherIndex := make(map[string]int)
	for _, row := range rows {
		key := row.TeacherNIK + "|" + row.TeacherName
		index, exists := teacherIndex[key]
		if !exists {
			response.Rekap = append(response.Rekap, dto.FoundationRecapTeacher{
				TeacherNIK:  row.TeacherNIK,
				TeacherName: row.TeacherName,
				Detail:      []dto.FoundationRecapClassDetail{},
			})
			index = len(response.Rekap) - 1
			teacherIndex[key] = index
		}

		response.Rekap[index].TotalJP += row.JumlahJP
		response.Rekap[index].Detail = append(response.Rekap[index].Detail, dto.FoundationRecapClassDetail{
			ClassCode: row.ClassCode,
			ClassName: row.ClassName,
			JumlahJP:  row.JumlahJP,
		})
		response.Rekap[index].TotalKelas = len(response.Rekap[index].Detail)
	}

	response.TotalPengajar = len(response.Rekap)
	sort.SliceStable(response.Rekap, func(i, j int) bool {
		if response.Rekap[i].TotalJP == response.Rekap[j].TotalJP {
			return response.Rekap[i].TeacherName < response.Rekap[j].TeacherName
		}
		return response.Rekap[i].TotalJP > response.Rekap[j].TotalJP
	})

	return response, nil
}

func (s *ScheduleService) CheckConflicts(schedule models.Schedule, ignoreID *uuid.UUID) ([]string, error) {
	var conflicts []string
	base := s.DB.Model(&models.Schedule{}).
		Where("date = ? AND time_start < ? AND time_end > ?", schedule.Date, schedule.TimeEnd, schedule.TimeStart)
	if ignoreID != nil {
		base = base.Where("id <> ?", *ignoreID)
	}

	var classCount int64
	if err := base.Session(&gorm.Session{}).Where("class_code = ?", schedule.ClassCode).Count(&classCount).Error; err != nil {
		return nil, err
	}
	if classCount > 0 {
		conflicts = append(conflicts, "Class "+schedule.ClassCode+" already has schedule at the selected time")
	}

	var teacherCount int64
	if err := base.Session(&gorm.Session{}).Where("teacher_nik = ?", schedule.TeacherNIK).Count(&teacherCount).Error; err != nil {
		return nil, err
	}
	if teacherCount > 0 {
		conflicts = append(conflicts, "Teacher "+schedule.TeacherNIK+" already has schedule at the selected time")
	}

	return conflicts, nil
}

func (s *ScheduleService) applyFilters(query *gorm.DB, filters ScheduleFilters) *gorm.DB {
	if filters.ClassCode != "" {
		query = query.Where("class_code = ?", filters.ClassCode)
	}
	if filters.TeacherNIK != "" {
		query = query.Where("teacher_nik = ?", filters.TeacherNIK)
	}
	if filters.Date != "" {
		query = query.Where("date = ?", filters.Date)
	}
	if filters.StartDate != "" {
		query = query.Where("date >= ?", filters.StartDate)
	}
	if filters.EndDate != "" {
		query = query.Where("date <= ?", filters.EndDate)
	}

	return query
}

func NormalizePagination(page int, limit int) (int, int) {
	return normalizePagination(page, limit)
}

func normalizePagination(page int, limit int) (int, int) {
	if page <= 0 {
		page = 1
	}
	if limit <= 0 {
		limit = 10
	}
	if limit > 100 {
		limit = 100
	}
	return page, limit
}

func IsNotFound(err error) bool {
	return errors.Is(err, gorm.ErrRecordNotFound)
}
