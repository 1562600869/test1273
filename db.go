package main

import (
	"database/sql"
	"fmt"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

var levels = []string{"入门", "一级", "二级", "三级", "四级", "五级"}
var statuses = []string{"在读", "暂停", "结业"}

func levelIndex(l string) (int, bool) {
	for i, v := range levels {
		if v == l {
			return i, true
		}
	}
	return -1, false
}

type Student struct {
	ID        int64  `json:"id"`
	Nickname  string `json:"nickname"`
	Phone     string `json:"phone"`
	Level     string `json:"level"`
	Status    string `json:"status"`
	CreatedAt string `json:"created_at"`
}

type CoursePackage struct {
	ID             int64  `json:"id"`
	StudentID      int64  `json:"student_id"`
	PurchaseDate   string `json:"purchase_date"`
	TotalHours     int    `json:"total_hours"`
	RemainingHours int    `json:"remaining_hours"`
	ExpiryDate     string `json:"expiry_date"`
}

type AttendanceRecord struct {
	ID          int64  `json:"id"`
	StudentID   int64  `json:"student_id"`
	PackageID   int64  `json:"package_id"`
	ConsumedAt  string `json:"consumed_at"`
	PackageInfo string `json:"package_info,omitempty"`
}

type GradeApplication struct {
	ID          int64  `json:"id"`
	StudentID   int64  `json:"student_id"`
	TargetLevel string `json:"target_level"`
	Passed      bool   `json:"passed"`
	AppliedAt   string `json:"applied_at"`
	Nickname    string `json:"nickname,omitempty"`
	Level       string `json:"level,omitempty"`
}

type MonthlyStat struct {
	Level string `json:"level"`
	Count int    `json:"count"`
}

func InitDB(dbPath string) (*sql.DB, error) {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("open db: %w", err)
	}
	if err := createTables(db); err != nil {
		db.Close()
		return nil, fmt.Errorf("create tables: %w", err)
	}
	return db, nil
}

func createTables(db *sql.DB) error {
	schema := `
	CREATE TABLE IF NOT EXISTS students (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		nickname TEXT NOT NULL,
		phone TEXT NOT NULL DEFAULT '',
		level TEXT NOT NULL DEFAULT '入门',
		status TEXT NOT NULL DEFAULT '在读',
		created_at TEXT NOT NULL DEFAULT (datetime('now','localtime'))
	);
	CREATE TABLE IF NOT EXISTS course_packages (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		student_id INTEGER NOT NULL,
		purchase_date TEXT NOT NULL,
		total_hours INTEGER NOT NULL CHECK(total_hours > 0),
		remaining_hours INTEGER NOT NULL DEFAULT 0,
		expiry_date TEXT NOT NULL,
		FOREIGN KEY(student_id) REFERENCES students(id)
	);
	CREATE TABLE IF NOT EXISTS attendance_records (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		student_id INTEGER NOT NULL,
		package_id INTEGER NOT NULL,
		consumed_at TEXT NOT NULL DEFAULT (datetime('now','localtime')),
		FOREIGN KEY(student_id) REFERENCES students(id),
		FOREIGN KEY(package_id) REFERENCES course_packages(id)
	);
	CREATE TABLE IF NOT EXISTS grade_applications (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		student_id INTEGER NOT NULL,
		target_level TEXT NOT NULL,
		passed INTEGER NOT NULL DEFAULT 0,
		applied_at TEXT NOT NULL DEFAULT (datetime('now','localtime')),
		FOREIGN KEY(student_id) REFERENCES students(id)
	);
	`
	_, err := db.Exec(schema)
	return err
}

func ListStudents(db *sql.DB) ([]Student, error) {
	rows, err := db.Query("SELECT id, nickname, phone, level, status, created_at FROM students ORDER BY id DESC")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []Student
	for rows.Next() {
		var s Student
		if err := rows.Scan(&s.ID, &s.Nickname, &s.Phone, &s.Level, &s.Status, &s.CreatedAt); err != nil {
			return nil, err
		}
		list = append(list, s)
	}
	return list, nil
}

func CreateStudent(db *sql.DB, nickname, phone, level, status string) (*Student, error) {
	if level == "" {
		level = "入门"
	}
	if status == "" {
		status = "在读"
	}
	res, err := db.Exec("INSERT INTO students (nickname, phone, level, status) VALUES (?, ?, ?, ?)", nickname, phone, level, status)
	if err != nil {
		return nil, err
	}
	id, _ := res.LastInsertId()
	return &Student{ID: id, Nickname: nickname, Phone: phone, Level: level, Status: status}, nil
}

func UpdateStudent(db *sql.DB, id int64, nickname, phone, level, status string) (*Student, error) {
	_, err := db.Exec("UPDATE students SET nickname=?, phone=?, level=?, status=? WHERE id=?", nickname, phone, level, status, id)
	if err != nil {
		return nil, err
	}
	return &Student{ID: id, Nickname: nickname, Phone: phone, Level: level, Status: status}, nil
}

func GetStudent(db *sql.DB, id int64) (*Student, error) {
	var s Student
	err := db.QueryRow("SELECT id, nickname, phone, level, status, created_at FROM students WHERE id=?", id).Scan(&s.ID, &s.Nickname, &s.Phone, &s.Level, &s.Status, &s.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &s, nil
}

func PurchasePackage(db *sql.DB, studentID int64, purchaseDate string, totalHours int, expiryDate string) (*CoursePackage, error) {
	res, err := db.Exec("INSERT INTO course_packages (student_id, purchase_date, total_hours, remaining_hours, expiry_date) VALUES (?, ?, ?, ?, ?)",
		studentID, purchaseDate, totalHours, totalHours, expiryDate)
	if err != nil {
		return nil, err
	}
	id, _ := res.LastInsertId()
	return &CoursePackage{ID: id, StudentID: studentID, PurchaseDate: purchaseDate, TotalHours: totalHours, RemainingHours: totalHours, ExpiryDate: expiryDate}, nil
}

func ListPackages(db *sql.DB, studentID int64) ([]CoursePackage, error) {
	rows, err := db.Query("SELECT id, student_id, purchase_date, total_hours, remaining_hours, expiry_date FROM course_packages WHERE student_id=? ORDER BY expiry_date ASC", studentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []CoursePackage
	for rows.Next() {
		var p CoursePackage
		if err := rows.Scan(&p.ID, &p.StudentID, &p.PurchaseDate, &p.TotalHours, &p.RemainingHours, &p.ExpiryDate); err != nil {
			return nil, err
		}
		list = append(list, p)
	}
	return list, nil
}

func ConsumeClass(db *sql.DB, studentID int64) (*AttendanceRecord, error) {
	tx, err := db.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	now := time.Now().Format("2006-01-02")

	var pkgID int64
	var remaining int
	err = tx.QueryRow(
		"SELECT id, remaining_hours FROM course_packages WHERE student_id=? AND expiry_date>=? AND remaining_hours>0 ORDER BY remaining_hours ASC, expiry_date ASC LIMIT 1",
		studentID, now).Scan(&pkgID, &remaining)
	if err == sql.ErrNoRows {
		expiredCount, _ := countRows(tx, "SELECT COUNT(*) FROM course_packages WHERE student_id=? AND expiry_date<? AND remaining_hours>0", studentID, now)
		usedUpCount, _ := countRows(tx, "SELECT COUNT(*) FROM course_packages WHERE student_id=? AND expiry_date>=? AND remaining_hours=0", studentID, now)
		reason := "无法消课："
		if expiredCount > 0 {
			reason += fmt.Sprintf("有%d个课时包已过期；", expiredCount)
		}
		if usedUpCount > 0 {
			reason += fmt.Sprintf("有%d个课时包已用完；", usedUpCount)
		}
		if expiredCount == 0 && usedUpCount == 0 {
			reason += "该学员没有课时包"
		}
		return nil, fmt.Errorf("%s", reason)
	}
	if err != nil {
		return nil, err
	}

	_, err = tx.Exec("UPDATE course_packages SET remaining_hours=remaining_hours-1 WHERE id=?", pkgID)
	if err != nil {
		return nil, err
	}

	var recordID int64
	res, err := tx.Exec("INSERT INTO attendance_records (student_id, package_id) VALUES (?, ?)", studentID, pkgID)
	if err != nil {
		return nil, err
	}
	recordID, _ = res.LastInsertId()

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return &AttendanceRecord{ID: recordID, StudentID: studentID, PackageID: pkgID}, nil
}

func countRows(tx *sql.Tx, query string, args ...interface{}) (int, error) {
	var count int
	err := tx.QueryRow(query, args...).Scan(&count)
	return count, err
}

func ApplyGrade(db *sql.DB, studentID int64, targetLevel string, passed bool) (*GradeApplication, error) {
	tx, err := db.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	var currentLevel string
	err = tx.QueryRow("SELECT level FROM students WHERE id=?", studentID).Scan(&currentLevel)
	if err != nil {
		return nil, fmt.Errorf("学员不存在")
	}

	curIdx, ok := levelIndex(currentLevel)
	if !ok {
		return nil, fmt.Errorf("学员当前级别无效: %s", currentLevel)
	}
	tgtIdx, ok := levelIndex(targetLevel)
	if !ok {
		return nil, fmt.Errorf("目标级别无效: %s", targetLevel)
	}
	if tgtIdx != curIdx+1 {
		return nil, fmt.Errorf("目标级别只能比当前级别高一级，当前: %s，目标: %s", currentLevel, targetLevel)
	}

	passedInt := 0
	if passed {
		passedInt = 1
	}
	res, err := tx.Exec("INSERT INTO grade_applications (student_id, target_level, passed) VALUES (?, ?, ?)", studentID, targetLevel, passedInt)
	if err != nil {
		return nil, err
	}
	appID, _ := res.LastInsertId()

	if passed {
		_, err = tx.Exec("UPDATE students SET level=? WHERE id=?", targetLevel, studentID)
		if err != nil {
			return nil, err
		}
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return &GradeApplication{ID: appID, StudentID: studentID, TargetLevel: targetLevel, Passed: passed}, nil
}

func GetRemainingHours(db *sql.DB, studentID int64) (int, error) {
	now := time.Now().Format("2006-01-02")
	var total int
	err := db.QueryRow(
		"SELECT COALESCE(SUM(remaining_hours),0) FROM course_packages WHERE student_id=? AND expiry_date>=? AND remaining_hours>0",
		studentID, now).Scan(&total)
	return total, err
}

func GetValidPackages(db *sql.DB, studentID int64) ([]CoursePackage, error) {
	now := time.Now().Format("2006-01-02")
	rows, err := db.Query(
		"SELECT id, student_id, purchase_date, total_hours, remaining_hours, expiry_date FROM course_packages WHERE student_id=? AND expiry_date>=? AND remaining_hours>0 ORDER BY remaining_hours ASC, expiry_date ASC",
		studentID, now)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []CoursePackage
	for rows.Next() {
		var p CoursePackage
		if err := rows.Scan(&p.ID, &p.StudentID, &p.PurchaseDate, &p.TotalHours, &p.RemainingHours, &p.ExpiryDate); err != nil {
			return nil, err
		}
		list = append(list, p)
	}
	return list, nil
}

func GetMonthlyStats(db *sql.DB) ([]MonthlyStat, error) {
	now := time.Now()
	monthStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location()).Format("2006-01-02")
	rows, err := db.Query(
		`SELECT s.level, COUNT(a.id) as cnt
		FROM students s
		JOIN attendance_records a ON s.id = a.student_id
		WHERE a.consumed_at >= ?
		GROUP BY s.level
		ORDER BY s.level`, monthStart)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []MonthlyStat
	for rows.Next() {
		var st MonthlyStat
		if err := rows.Scan(&st.Level, &st.Count); err != nil {
			return nil, err
		}
		list = append(list, st)
	}
	return list, nil
}

func ListAttendance(db *sql.DB, studentID int64) ([]AttendanceRecord, error) {
	rows, err := db.Query(
		`SELECT a.id, a.student_id, a.package_id, a.consumed_at,
		'购买日:'||cp.purchase_date||' 过期:'||cp.expiry_date||' 剩余:'||cp.remaining_hours as package_info
		FROM attendance_records a
		JOIN course_packages cp ON a.package_id = cp.id
		WHERE a.student_id=?
		ORDER BY a.consumed_at DESC`, studentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []AttendanceRecord
	for rows.Next() {
		var r AttendanceRecord
		if err := rows.Scan(&r.ID, &r.StudentID, &r.PackageID, &r.ConsumedAt, &r.PackageInfo); err != nil {
			return nil, err
		}
		list = append(list, r)
	}
	return list, nil
}

func ListGradeApplications(db *sql.DB) ([]GradeApplication, error) {
	rows, err := db.Query(
		`SELECT g.id, g.student_id, g.target_level, g.passed, g.applied_at, s.nickname, s.level
		FROM grade_applications g
		JOIN students s ON g.student_id = s.id
		ORDER BY g.applied_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []GradeApplication
	for rows.Next() {
		var g GradeApplication
		var passedInt int
		if err := rows.Scan(&g.ID, &g.StudentID, &g.TargetLevel, &passedInt, &g.AppliedAt, &g.Nickname, &g.Level); err != nil {
			return nil, err
		}
		g.Passed = passedInt == 1
		list = append(list, g)
	}
	return list, nil
}
