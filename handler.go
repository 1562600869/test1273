package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
)

type Server struct {
	DB *sql.DB
}

type APIResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func (s *Server) ok(w http.ResponseWriter, data interface{}) {
	writeJSON(w, http.StatusOK, APIResponse{Success: true, Data: data})
}

func (s *Server) fail(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, APIResponse{Success: false, Error: msg})
}

func (s *Server) HandleStudents(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		list, err := ListStudents(s.DB)
		if err != nil {
			s.fail(w, 500, err.Error())
			return
		}
		if list == nil {
			list = []Student{}
		}
		s.ok(w, list)
	case http.MethodPost:
		var req struct {
			Nickname string `json:"nickname"`
			Phone    string `json:"phone"`
			Level    string `json:"level"`
			Status   string `json:"status"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			s.fail(w, 400, "请求格式错误")
			return
		}
		if req.Nickname == "" {
			s.fail(w, 400, "昵称不能为空")
			return
		}
		stu, err := CreateStudent(s.DB, req.Nickname, req.Phone, req.Level, req.Status)
		if err != nil {
			s.fail(w, 500, err.Error())
			return
		}
		s.ok(w, stu)
	default:
		s.fail(w, 405, "方法不允许")
	}
}

func (s *Server) HandleStudent(w http.ResponseWriter, r *http.Request, id int64) {
	switch r.Method {
	case http.MethodGet:
		stu, err := GetStudent(s.DB, id)
		if err != nil {
			s.fail(w, 500, err.Error())
			return
		}
		if stu == nil {
			s.fail(w, 404, "学员不存在")
			return
		}
		s.ok(w, stu)
	case http.MethodPut:
		var req struct {
			Nickname string `json:"nickname"`
			Phone    string `json:"phone"`
			Level    string `json:"level"`
			Status   string `json:"status"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			s.fail(w, 400, "请求格式错误")
			return
		}
		stu, err := UpdateStudent(s.DB, id, req.Nickname, req.Phone, req.Level, req.Status)
		if err != nil {
			s.fail(w, 500, err.Error())
			return
		}
		s.ok(w, stu)
	case http.MethodDelete:
		_, err := s.DB.Exec("DELETE FROM students WHERE id=?", id)
		if err != nil {
			s.fail(w, 500, err.Error())
			return
		}
		s.ok(w, nil)
	default:
		s.fail(w, 405, "方法不允许")
	}
}

func (s *Server) HandlePackages(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		s.fail(w, 405, "方法不允许")
		return
	}
	var req struct {
		StudentID    int64  `json:"student_id"`
		PurchaseDate string `json:"purchase_date"`
		TotalHours   int    `json:"total_hours"`
		ExpiryDate   string `json:"expiry_date"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.fail(w, 400, "请求格式错误")
		return
	}
	if req.StudentID == 0 {
		s.fail(w, 400, "学员ID不能为空")
		return
	}
	if req.TotalHours <= 0 {
		s.fail(w, 400, "课时数必须为正整数")
		return
	}
	if req.PurchaseDate == "" || req.ExpiryDate == "" {
		s.fail(w, 400, "购买日期和有效期不能为空")
		return
	}
	stu, err := GetStudent(s.DB, req.StudentID)
	if err != nil || stu == nil {
		s.fail(w, 404, "学员不存在")
		return
	}
	pkg, err := PurchasePackage(s.DB, req.StudentID, req.PurchaseDate, req.TotalHours, req.ExpiryDate)
	if err != nil {
		s.fail(w, 500, err.Error())
		return
	}
	s.ok(w, pkg)
}

func (s *Server) HandleStudentPackages(w http.ResponseWriter, r *http.Request, studentID int64) {
	if r.Method != http.MethodGet {
		s.fail(w, 405, "方法不允许")
		return
	}
	list, err := ListPackages(s.DB, studentID)
	if err != nil {
		s.fail(w, 500, err.Error())
		return
	}
	if list == nil {
		list = []CoursePackage{}
	}
	s.ok(w, list)
}

func (s *Server) HandleConsumeClass(w http.ResponseWriter, r *http.Request, studentID int64) {
	if r.Method != http.MethodPost {
		s.fail(w, 405, "方法不允许")
		return
	}
	rec, err := ConsumeClass(s.DB, studentID)
	if err != nil {
		s.fail(w, 400, err.Error())
		return
	}
	s.ok(w, rec)
}

func (s *Server) HandleApplyGrade(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		s.fail(w, 405, "方法不允许")
		return
	}
	var req struct {
		StudentID   int64  `json:"student_id"`
		TargetLevel string `json:"target_level"`
		Passed      bool   `json:"passed"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.fail(w, 400, "请求格式错误")
		return
	}
	if req.StudentID == 0 {
		s.fail(w, 400, "学员ID不能为空")
		return
	}
	if req.TargetLevel == "" {
		s.fail(w, 400, "目标级别不能为空")
		return
	}
	app, err := ApplyGrade(s.DB, req.StudentID, req.TargetLevel, req.Passed)
	if err != nil {
		s.fail(w, 400, err.Error())
		return
	}
	s.ok(w, app)
}

func (s *Server) HandleRemaining(w http.ResponseWriter, r *http.Request, studentID int64) {
	if r.Method != http.MethodGet {
		s.fail(w, 405, "方法不允许")
		return
	}
	total, err := GetRemainingHours(s.DB, studentID)
	if err != nil {
		s.fail(w, 500, err.Error())
		return
	}
	pkgs, err := GetValidPackages(s.DB, studentID)
	if err != nil {
		s.fail(w, 500, err.Error())
		return
	}
	if pkgs == nil {
		pkgs = []CoursePackage{}
	}
	s.ok(w, map[string]interface{}{"total_remaining": total, "packages": pkgs})
}

func (s *Server) HandleMonthlyStats(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		s.fail(w, 405, "方法不允许")
		return
	}
	stats, err := GetMonthlyStats(s.DB)
	if err != nil {
		s.fail(w, 500, err.Error())
		return
	}
	if stats == nil {
		stats = []MonthlyStat{}
	}
	s.ok(w, stats)
}

func (s *Server) HandleAttendance(w http.ResponseWriter, r *http.Request, studentID int64) {
	if r.Method != http.MethodGet {
		s.fail(w, 405, "方法不允许")
		return
	}
	list, err := ListAttendance(s.DB, studentID)
	if err != nil {
		s.fail(w, 500, err.Error())
		return
	}
	if list == nil {
		list = []AttendanceRecord{}
	}
	s.ok(w, list)
}

func (s *Server) HandleGradeApplications(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		s.fail(w, 405, "方法不允许")
		return
	}
	list, err := ListGradeApplications(s.DB)
	if err != nil {
		s.fail(w, 500, err.Error())
		return
	}
	if list == nil {
		list = []GradeApplication{}
	}
	s.ok(w, list)
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path

	if strings.HasPrefix(path, "/api/") {
		s.handleAPI(w, r, path)
		return
	}

	w.Header().Set("Cache-Control", "no-cache")
	http.ServeFile(w, r, "index.html")
}

func (s *Server) handleAPI(w http.ResponseWriter, r *http.Request, path string) {
	switch {
	case path == "/api/students" && r.Method == http.MethodGet:
		s.HandleStudents(w, r)
	case path == "/api/students" && r.Method == http.MethodPost:
		s.HandleStudents(w, r)

	case strings.HasPrefix(path, "/api/students/"):
		parts := strings.Split(strings.TrimPrefix(path, "/api/students/"), "/")
		id, err := strconv.ParseInt(parts[0], 10, 64)
		if err != nil {
			s.fail(w, 400, "无效的学员ID")
			return
		}
		if len(parts) == 1 {
			s.HandleStudent(w, r, id)
			return
		}
		switch parts[1] {
		case "packages":
			if len(parts) == 2 {
				if r.Method == http.MethodGet {
					s.HandleStudentPackages(w, r, id)
				} else {
					s.HandlePackages(w, r)
				}
			}
		case "consume":
			s.HandleConsumeClass(w, r, id)
		case "remaining":
			s.HandleRemaining(w, r, id)
		case "attendance":
			s.HandleAttendance(w, r, id)
		default:
			s.fail(w, 404, fmt.Sprintf("未知路径: %s", path))
		}

	case path == "/api/packages" && r.Method == http.MethodPost:
		s.HandlePackages(w, r)
	case path == "/api/grades":
		s.HandleApplyGrade(w, r)
	case path == "/api/grades" && r.Method == http.MethodGet:
		s.HandleGradeApplications(w, r)
	case path == "/api/stats/monthly":
		s.HandleMonthlyStats(w, r)
	default:
		s.fail(w, 404, fmt.Sprintf("未知路径: %s", path))
	}
}
