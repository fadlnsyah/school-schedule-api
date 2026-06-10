package controllers

import (
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"school-schedule-api/dto"
	"school-schedule-api/models"
	"school-schedule-api/services"
	"school-schedule-api/utils"

	"github.com/gin-gonic/gin"
)

type ScheduleController struct {
	Service      *services.ScheduleService
	ExcelService *services.ExcelService
}

func NewScheduleController(service *services.ScheduleService) *ScheduleController {
	return &ScheduleController{
		Service:      service,
		ExcelService: services.NewExcelService(service),
	}
}

// Health godoc
// @Summary Health check
// @Tags health
// @Produce json
// @Success 200 {object} map[string]string
// @Router /health [get]
func Health(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":  "ok",
		"message": "School Schedule API is running",
	})
}

func (ctrl *ScheduleController) Create(c *gin.Context) {
	var req dto.ScheduleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, http.StatusBadRequest, "Invalid JSON request")
		return
	}

	if details := services.ValidateScheduleRequest(req); len(details) > 0 {
		utils.ValidationError(c, details)
		return
	}

	schedule, conflicts, err := ctrl.Service.Create(req)
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, "Failed to create schedule")
		return
	}
	if len(conflicts) > 0 {
		utils.ConflictError(c, conflicts)
		return
	}

	utils.Success(c, http.StatusCreated, "Schedule created successfully", schedule)
}

func (ctrl *ScheduleController) FindAll(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	filters := services.ScheduleFilters{
		ClassCode:  c.Query("class_code"),
		TeacherNIK: c.Query("teacher_nik"),
		Date:       c.Query("date"),
		StartDate:  c.Query("start_date"),
		EndDate:    c.Query("end_date"),
		Page:       page,
		Limit:      limit,
	}

	if details := validateOptionalDateFilters(filters); len(details) > 0 {
		utils.ValidationError(c, details)
		return
	}

	schedules, total, err := ctrl.Service.FindAll(filters)
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, "Failed to retrieve schedules")
		return
	}

	page, limit = services.NormalizePagination(page, limit)
	meta := dto.PaginationMeta{Page: page, Limit: limit, Total: total}
	utils.SuccessWithMeta(c, http.StatusOK, "Schedules retrieved successfully", schedules, meta)
}

func (ctrl *ScheduleController) FindByID(c *gin.Context) {
	schedule, err := ctrl.Service.FindByID(c.Param("id"))
	if err != nil {
		if services.IsNotFound(err) {
			utils.Error(c, http.StatusNotFound, "Schedule not found")
			return
		}
		utils.Error(c, http.StatusInternalServerError, "Failed to retrieve schedule")
		return
	}

	utils.Success(c, http.StatusOK, "Schedule retrieved successfully", schedule)
}

func (ctrl *ScheduleController) Update(c *gin.Context) {
	var req dto.ScheduleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, http.StatusBadRequest, "Invalid JSON request")
		return
	}

	if details := services.ValidateScheduleRequest(req); len(details) > 0 {
		utils.ValidationError(c, details)
		return
	}

	schedule, conflicts, err := ctrl.Service.Update(c.Param("id"), req)
	if err != nil {
		if services.IsNotFound(err) {
			utils.Error(c, http.StatusNotFound, "Schedule not found")
			return
		}
		utils.Error(c, http.StatusInternalServerError, "Failed to update schedule")
		return
	}
	if len(conflicts) > 0 {
		utils.ConflictError(c, conflicts)
		return
	}

	utils.Success(c, http.StatusOK, "Schedule updated successfully", schedule)
}

func (ctrl *ScheduleController) Delete(c *gin.Context) {
	if err := ctrl.Service.Delete(c.Param("id")); err != nil {
		if services.IsNotFound(err) {
			utils.Error(c, http.StatusNotFound, "Schedule not found")
			return
		}
		utils.Error(c, http.StatusInternalServerError, "Failed to delete schedule")
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Schedule deleted successfully"})
}

func (ctrl *ScheduleController) Student(c *gin.Context) {
	classCode := c.Query("class_code")
	date := c.Query("date")
	var details []string
	if classCode == "" {
		details = append(details, "class_code is required")
	}
	if date == "" {
		details = append(details, "date is required")
	} else if _, err := time.Parse("2006-01-02", date); err != nil {
		details = append(details, "date must be YYYY-MM-DD")
	}
	if len(details) > 0 {
		utils.ValidationError(c, details)
		return
	}

	schedules, err := ctrl.Service.StudentSchedule(classCode, date)
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, "Failed to retrieve student schedule")
		return
	}

	c.JSON(http.StatusOK, studentResponse(classCode, date, schedules))
}

func (ctrl *ScheduleController) Teacher(c *gin.Context) {
	teacherNIK := c.Query("teacher_nik")
	startDate := c.Query("start_date")
	endDate := c.Query("end_date")
	details := services.ValidateDateRange(startDate, endDate)
	if teacherNIK == "" {
		details = append([]string{"teacher_nik is required"}, details...)
	}
	if len(details) > 0 {
		utils.ValidationError(c, details)
		return
	}

	schedules, err := ctrl.Service.TeacherSchedule(teacherNIK, startDate, endDate)
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, "Failed to retrieve teacher schedule")
		return
	}

	c.JSON(http.StatusOK, teacherResponse(teacherNIK, startDate, endDate, schedules))
}

func (ctrl *ScheduleController) RecapJP(c *gin.Context) {
	startDate := c.Query("start_date")
	endDate := c.Query("end_date")
	if details := services.ValidateDateRange(startDate, endDate); len(details) > 0 {
		utils.ValidationError(c, details)
		return
	}

	data, err := ctrl.Service.RecapJP(startDate, endDate)
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, "Failed to retrieve recap JP")
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"start_date": startDate,
		"end_date":   endDate,
		"data":       data,
	})
}

func (ctrl *ScheduleController) Upload(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		utils.Error(c, http.StatusBadRequest, "file is required")
		return
	}
	if filepath.Ext(file.Filename) != ".xlsx" {
		utils.Error(c, http.StatusBadRequest, "file must be .xlsx")
		return
	}

	if err := os.MkdirAll("tmp", 0755); err != nil {
		utils.Error(c, http.StatusInternalServerError, "Failed to prepare upload directory")
		return
	}
	savedPath := filepath.Join("tmp", strconv.FormatInt(time.Now().UnixNano(), 10)+"-"+filepath.Base(file.Filename))
	if err := c.SaveUploadedFile(file, savedPath); err != nil {
		utils.Error(c, http.StatusInternalServerError, "Failed to save uploaded file")
		return
	}
	defer func() { _ = os.Remove(savedPath) }()

	result, err := ctrl.ExcelService.ImportSchedules(file, savedPath)
	if err != nil {
		utils.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":  "Excel uploaded successfully",
		"inserted": result.Inserted,
		"failed":   result.Failed,
		"errors":   result.Errors,
	})
}

func (ctrl *ScheduleController) Export(c *gin.Context) {
	startDate := c.Query("start_date")
	endDate := c.Query("end_date")
	if details := services.ValidateDateRange(startDate, endDate); len(details) > 0 {
		utils.ValidationError(c, details)
		return
	}

	file, err := ctrl.ExcelService.ExportRecapJP(startDate, endDate)
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, "Failed to export recap JP")
		return
	}
	defer func() { _ = file.Close() }()

	filename := "rekap-jp-" + startDate + "-" + endDate + ".xlsx"
	c.Header("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	c.Header("Content-Disposition", `attachment; filename="`+filename+`"`)
	c.Header("File-Name", filename)
	if err := file.Write(c.Writer); err != nil {
		utils.Error(c, http.StatusInternalServerError, "Failed to write Excel file")
		return
	}
}

func validateOptionalDateFilters(filters services.ScheduleFilters) []string {
	var details []string
	if filters.Date != "" {
		if _, err := time.Parse("2006-01-02", filters.Date); err != nil {
			details = append(details, "date must be YYYY-MM-DD")
		}
	}
	if filters.StartDate != "" {
		if _, err := time.Parse("2006-01-02", filters.StartDate); err != nil {
			details = append(details, "start_date must be YYYY-MM-DD")
		}
	}
	if filters.EndDate != "" {
		if _, err := time.Parse("2006-01-02", filters.EndDate); err != nil {
			details = append(details, "end_date must be YYYY-MM-DD")
		}
	}
	if filters.StartDate != "" && filters.EndDate != "" {
		if rangeDetails := services.ValidateDateRange(filters.StartDate, filters.EndDate); len(rangeDetails) > 0 {
			details = append(details, rangeDetails...)
		}
	}
	return details
}

func studentResponse(classCode string, date string, schedules []models.Schedule) gin.H {
	className := ""
	items := make([]gin.H, 0, len(schedules))
	for _, schedule := range schedules {
		if className == "" {
			className = schedule.ClassName
		}
		items = append(items, gin.H{
			"jam_ke":       schedule.JamKe,
			"subject_code": schedule.SubjectCode,
			"teacher_name": schedule.TeacherName,
			"time_start":   schedule.TimeStart,
			"time_end":     schedule.TimeEnd,
		})
	}

	return gin.H{"class_code": classCode, "class_name": className, "date": date, "schedules": items}
}

func teacherResponse(teacherNIK string, startDate string, endDate string, schedules []models.Schedule) gin.H {
	teacherName := ""
	items := make([]gin.H, 0, len(schedules))
	for _, schedule := range schedules {
		if teacherName == "" {
			teacherName = schedule.TeacherName
		}
		items = append(items, gin.H{
			"date":         schedule.Date,
			"class_code":   schedule.ClassCode,
			"class_name":   schedule.ClassName,
			"subject_code": schedule.SubjectCode,
			"jam_ke":       schedule.JamKe,
			"time_start":   schedule.TimeStart,
			"time_end":     schedule.TimeEnd,
		})
	}

	return gin.H{
		"teacher_nik":  teacherNIK,
		"teacher_name": teacherName,
		"start_date":   startDate,
		"end_date":     endDate,
		"total_jp":     len(schedules),
		"schedules":    items,
	}
}
