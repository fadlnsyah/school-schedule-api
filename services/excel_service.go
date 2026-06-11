package services

import (
	"fmt"
	"mime/multipart"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"school-schedule-api/dto"
	"school-schedule-api/models"

	"github.com/google/uuid"
	"github.com/xuri/excelize/v2"
)

type UploadRowError struct {
	Row   int    `json:"row"`
	Error string `json:"error"`
}

type UploadResult struct {
	Inserted int              `json:"inserted"`
	Failed   int              `json:"failed"`
	Errors   []UploadRowError `json:"errors"`
}

type ExcelService struct {
	ScheduleService *ScheduleService
}

type exportTeacherRecap struct {
	TeacherNIK  string
	TeacherName string
	Classes     map[string]string
	Weeks       [5]int
	TotalJP     int
}

func NewExcelService(scheduleService *ScheduleService) *ExcelService {
	return &ExcelService{ScheduleService: scheduleService}
}

func (s *ExcelService) ImportSchedules(file *multipart.FileHeader, savedPath string) (UploadResult, error) {
	result := UploadResult{Errors: []UploadRowError{}}
	if strings.ToLower(filepath.Ext(file.Filename)) != ".xlsx" {
		return result, fmt.Errorf("file must be .xlsx")
	}

	xlsx, err := excelize.OpenFile(savedPath)
	if err != nil {
		return result, err
	}
	defer func() { _ = xlsx.Close() }()

	sheet := xlsx.GetSheetName(0)
	rows, err := xlsx.GetRows(sheet)
	if err != nil {
		return result, err
	}
	if len(rows) < 1 {
		return result, fmt.Errorf("excel file is empty")
	}

	headerMap := mapHeaders(rows[0])
	required := []string{"class_code", "class_name", "subject_code", "teacher_nik", "teacher_name", "date", "jam_ke", "time_start", "time_end"}
	for _, key := range required {
		if _, ok := headerMap[key]; !ok {
			return result, fmt.Errorf("missing required column: %s", key)
		}
	}

	for i := 1; i < len(rows); i++ {
		rowNumber := i + 1
		row := rows[i]
		if isEmptyRow(row) {
			continue
		}

		req, err := rowToScheduleRequest(xlsx, sheet, rowNumber, headerMap)
		if err != nil {
			result.Failed++
			result.Errors = append(result.Errors, UploadRowError{Row: rowNumber, Error: err.Error()})
			continue
		}

		if details := ValidateScheduleRequest(req); len(details) > 0 {
			result.Failed++
			result.Errors = append(result.Errors, UploadRowError{Row: rowNumber, Error: strings.Join(details, ", ")})
			continue
		}

		schedule := RequestToSchedule(req)
		conflicts, err := s.ScheduleService.CheckConflicts(schedule, nil)
		if err != nil {
			result.Failed++
			result.Errors = append(result.Errors, UploadRowError{Row: rowNumber, Error: "failed to check schedule conflict"})
			continue
		}
		if len(conflicts) > 0 {
			result.Failed++
			result.Errors = append(result.Errors, UploadRowError{Row: rowNumber, Error: strings.Join(conflicts, ", ")})
			continue
		}

		schedule.ID = uuid.New()
		if err := s.ScheduleService.DB.Create(&schedule).Error; err != nil {
			result.Failed++
			result.Errors = append(result.Errors, UploadRowError{Row: rowNumber, Error: "failed to insert schedule"})
			continue
		}
		result.Inserted++
	}

	return result, nil
}

func (s *ExcelService) ExportRecapJP(startDate string, endDate string) (*excelize.File, error) {
	rows, err := s.buildWeeklyTeacherRecap(startDate, endDate)
	if err != nil {
		return nil, err
	}

	f := excelize.NewFile()
	sheet := "Rekap JP"
	f.SetSheetName("Sheet1", sheet)

	_ = f.MergeCell(sheet, "A1", "J1")
	_ = f.SetCellValue(sheet, "A1", "Rekap Jam Pelajaran Pengajar")
	_ = f.MergeCell(sheet, "A2", "J2")
	_ = f.SetCellValue(sheet, "A2", "Periode: "+startDate+" s/d "+endDate)

	headerStyle, _ := f.NewStyle(&excelize.Style{
		Font:   &excelize.Font{Bold: true},
		Border: borderStyle(),
	})
	titleStyle, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{Bold: true, Size: 14},
		Alignment: &excelize.Alignment{
			Horizontal: "center",
		},
	})
	cellStyle, _ := f.NewStyle(&excelize.Style{Border: borderStyle()})

	_ = f.SetCellStyle(sheet, "A1", "J1", titleStyle)
	headers := []string{"No", "NIK", "Nama Pengajar", "Kelas yg Diajar", "Pekan 1", "Pekan 2", "Pekan 3", "Pekan 4", "Pekan 5", "Total JP"}
	for i, header := range headers {
		cell, _ := excelize.CoordinatesToCellName(i+1, 4)
		_ = f.SetCellValue(sheet, cell, header)
		_ = f.SetCellStyle(sheet, cell, cell, headerStyle)
	}

	for i, row := range rows {
		r := i + 5
		values := []any{
			i + 1,
			row.TeacherNIK,
			row.TeacherName,
			strings.Join(sortedClassNames(row.Classes), ", "),
			row.Weeks[0],
			row.Weeks[1],
			row.Weeks[2],
			row.Weeks[3],
			row.Weeks[4],
			row.TotalJP,
		}
		for col, value := range values {
			cell, _ := excelize.CoordinatesToCellName(col+1, r)
			_ = f.SetCellValue(sheet, cell, value)
			_ = f.SetCellStyle(sheet, cell, cell, cellStyle)
		}
	}

	_ = f.SetColWidth(sheet, "A", "A", 8)
	_ = f.SetColWidth(sheet, "B", "B", 18)
	_ = f.SetColWidth(sheet, "C", "C", 34)
	_ = f.SetColWidth(sheet, "D", "D", 32)
	_ = f.SetColWidth(sheet, "E", "J", 12)

	return f, nil
}

func (s *ExcelService) buildWeeklyTeacherRecap(startDate string, endDate string) ([]exportTeacherRecap, error) {
	var schedules []models.Schedule
	err := s.ScheduleService.DB.
		Where("date >= ? AND date <= ?", startDate, endDate).
		Order("teacher_name ASC, teacher_nik ASC, date ASC").
		Find(&schedules).Error
	if err != nil {
		return nil, err
	}

	recapByTeacher := make(map[string]*exportTeacherRecap)
	for _, schedule := range schedules {
		key := schedule.TeacherNIK + "|" + schedule.TeacherName
		recap, exists := recapByTeacher[key]
		if !exists {
			recap = &exportTeacherRecap{
				TeacherNIK:  schedule.TeacherNIK,
				TeacherName: schedule.TeacherName,
				Classes:     map[string]string{},
			}
			recapByTeacher[key] = recap
		}

		classLabel := strings.TrimSpace(schedule.ClassName)
		if classLabel == "" {
			classLabel = strings.TrimSpace(schedule.ClassCode)
		}
		if classLabel != "" {
			recap.Classes[schedule.ClassCode+"|"+schedule.ClassName] = classLabel
		}

		weekIndex, ok := weekIndexFromDate(schedule.Date)
		if ok {
			recap.Weeks[weekIndex]++
			recap.TotalJP++
		}
	}

	rows := make([]exportTeacherRecap, 0, len(recapByTeacher))
	for _, recap := range recapByTeacher {
		rows = append(rows, *recap)
	}
	sort.SliceStable(rows, func(i, j int) bool {
		if rows[i].TotalJP == rows[j].TotalJP {
			return rows[i].TeacherName < rows[j].TeacherName
		}
		return rows[i].TotalJP > rows[j].TotalJP
	})

	return rows, nil
}

func weekIndexFromDate(value string) (int, bool) {
	if len(value) > len("2006-01-02") {
		value = value[:len("2006-01-02")]
	}
	parsed, err := time.Parse("2006-01-02", value)
	if err != nil {
		return 0, false
	}

	switch day := parsed.Day(); {
	case day <= 7:
		return 0, true
	case day <= 14:
		return 1, true
	case day <= 21:
		return 2, true
	case day <= 28:
		return 3, true
	default:
		return 4, true
	}
}

func sortedClassNames(classes map[string]string) []string {
	result := make([]string, 0, len(classes))
	for _, className := range classes {
		result = append(result, className)
	}
	sort.Strings(result)
	return result
}

func borderStyle() []excelize.Border {
	return []excelize.Border{
		{Type: "left", Color: "000000", Style: 1},
		{Type: "top", Color: "000000", Style: 1},
		{Type: "bottom", Color: "000000", Style: 1},
		{Type: "right", Color: "000000", Style: 1},
	}
}

func mapHeaders(headers []string) map[string]int {
	result := map[string]int{}
	for i, header := range headers {
		key := strings.ToLower(strings.TrimSpace(header))
		if key != "" {
			result[key] = i
		}
	}
	return result
}

func rowToScheduleRequest(f *excelize.File, sheet string, rowNumber int, headerMap map[string]int) (dto.ScheduleRequest, error) {
	get := func(key string) string {
		col := headerMap[key] + 1
		cell, _ := excelize.CoordinatesToCellName(col, rowNumber)
		value, _ := f.GetCellValue(sheet, cell)
		return strings.TrimSpace(value)
	}

	jamKe, err := strconv.Atoi(get("jam_ke"))
	if err != nil {
		return dto.ScheduleRequest{}, fmt.Errorf("jam_ke must be a number")
	}

	date, err := normalizeExcelDate(f, sheet, rowNumber, headerMap["date"]+1, get("date"))
	if err != nil {
		return dto.ScheduleRequest{}, err
	}

	timeStart, err := normalizeExcelTime(f, sheet, rowNumber, headerMap["time_start"]+1, get("time_start"))
	if err != nil {
		return dto.ScheduleRequest{}, fmt.Errorf("invalid time_start format")
	}
	timeEnd, err := normalizeExcelTime(f, sheet, rowNumber, headerMap["time_end"]+1, get("time_end"))
	if err != nil {
		return dto.ScheduleRequest{}, fmt.Errorf("invalid time_end format")
	}

	return dto.ScheduleRequest{
		ClassCode:   get("class_code"),
		ClassName:   get("class_name"),
		SubjectCode: get("subject_code"),
		TeacherNIK:  get("teacher_nik"),
		TeacherName: get("teacher_name"),
		Date:        date,
		JamKe:       jamKe,
		TimeStart:   timeStart,
		TimeEnd:     timeEnd,
	}, nil
}

func normalizeExcelDate(f *excelize.File, sheet string, row int, col int, value string) (string, error) {
	if parsed, err := time.Parse("2006-01-02", value); err == nil {
		return parsed.Format("2006-01-02"), nil
	}

	cell, _ := excelize.CoordinatesToCellName(col, row)
	raw, _ := f.GetCellValue(sheet, cell, excelize.Options{RawCellValue: true})
	serial, err := strconv.ParseFloat(raw, 64)
	if err != nil {
		return "", fmt.Errorf("invalid date format")
	}
	parsed, err := excelize.ExcelDateToTime(serial, false)
	if err != nil {
		return "", fmt.Errorf("invalid date format")
	}
	return parsed.Format("2006-01-02"), nil
}

func normalizeExcelTime(f *excelize.File, sheet string, row int, col int, value string) (string, error) {
	if parsed, err := time.Parse("15:04:05", value); err == nil {
		return parsed.Format("15:04:05"), nil
	}
	if parsed, err := time.Parse("15:04", value); err == nil {
		return parsed.Format("15:04:05"), nil
	}

	cell, _ := excelize.CoordinatesToCellName(col, row)
	raw, _ := f.GetCellValue(sheet, cell, excelize.Options{RawCellValue: true})
	serial, err := strconv.ParseFloat(raw, 64)
	if err != nil {
		return "", err
	}
	seconds := int(serial * 24 * 60 * 60)
	return fmt.Sprintf("%02d:%02d:%02d", seconds/3600, (seconds%3600)/60, seconds%60), nil
}

func isEmptyRow(row []string) bool {
	for _, value := range row {
		if strings.TrimSpace(value) != "" {
			return false
		}
	}
	return true
}
