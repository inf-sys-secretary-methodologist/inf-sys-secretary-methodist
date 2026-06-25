// Command mock-1c is a standalone fake 1С:Предприятие OData v2 endpoint used
// to validate the integration module end-to-end without a real 1С server.
//
// A genuine public 1С OData test server does not exist, and real 1С demo bases
// (e.g. 1С:Фреш «Управление нашей фирмой») expose generic catalogs that do not
// match the Catalog_Сотрудники / Catalog_Студенты shape this system imports.
// This mock serves exactly the JSON our odata.Client consumes — it reuses the
// real entities.ODataEmployee / entities.ODataStudent types as the single
// source of truth, so the wire contract cannot drift.
//
// Usage:
//
//	go run ./cmd/mock-1c                 # listens on :9191
//	MOCK_1C_PORT=9300 go run ./cmd/mock-1c
//
// Point the backend at it:
//
//	INTEGRATION_1C_ENABLED=true
//	INTEGRATION_1C_BASE_URL=http://mock-1c:9191   # in compose
//	INTEGRATION_1C_USERNAME=demo
//	INTEGRATION_1C_PASSWORD=demo
//
// Then trigger a sync: POST /api/integration/sync/start {"entity_type":"employee"}.
package main

import (
	"encoding/base64"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/integration/domain/entities"
)

// odataList mirrors the OData v2 collection envelope the client decodes
// (see odata.Response[T]: "odata.metadata" + "value").
type odataList[T any] struct {
	Metadata string `json:"odata.metadata"`
	Value    []T    `json:"value"`
}

// sampleEmployees returns a small, stable employee set. Emails match the demo
// local users so the admin can exercise the link flow after a sync.
func sampleEmployees() []entities.ODataEmployee {
	return []entities.ODataEmployee{
		{
			RefKey: "11111111-1111-1111-1111-111111111111", Code: "СОТР-0001",
			Description: "Иванова Мария Петровна", FirstName: "Мария", LastName: "Иванова", MiddleName: "Петровна",
			Email: "methodist@inf-sys.local", Phone: "+7 900 000-00-01",
			Position: "Методист", Department: "Учебно-методический отдел", EmploymentDate: "2019-09-01T00:00:00",
		},
		{
			RefKey: "22222222-2222-2222-2222-222222222222", Code: "СОТР-0002",
			Description: "Петров Сергей Иванович", FirstName: "Сергей", LastName: "Петров", MiddleName: "Иванович",
			Email: "teacher@inf-sys.local", Phone: "+7 900 000-00-02",
			Position: "Преподаватель", Department: "Кафедра информатики", EmploymentDate: "2017-09-01T00:00:00",
		},
		{
			RefKey: "33333333-3333-3333-3333-333333333333", Code: "СОТР-0003",
			Description: "Сидорова Анна Викторовна", FirstName: "Анна", LastName: "Сидорова", MiddleName: "Викторовна",
			Email: "secretary@inf-sys.local", Phone: "+7 900 000-00-03",
			Position: "Учёный секретарь", Department: "Учебная часть", EmploymentDate: "2020-02-10T00:00:00",
			DeletionMark: false,
		},
	}
}

// sampleStudents returns a small, stable student set across two groups so a
// later academic-debt demo has realistic group/semester data to work with.
func sampleStudents() []entities.ODataStudent {
	return []entities.ODataStudent{
		{
			RefKey: "aaaaaaaa-0001-0000-0000-000000000001", Code: "СТУД-0001",
			Description: "Кузнецов Дмитрий Алексеевич", FirstName: "Дмитрий", LastName: "Кузнецов", MiddleName: "Алексеевич",
			Email: "student@inf-sys.local", Phone: "+7 901 000-00-01",
			GroupName: "БИ-21", Faculty: "Факультет ИТ", Specialty: "Бизнес-информатика", Course: 3,
			StudyForm: "Очная", EnrollmentDate: "2021-09-01T00:00:00", Status: "Учится",
		},
		{
			RefKey: "aaaaaaaa-0002-0000-0000-000000000002", Code: "СТУД-0002",
			Description: "Смирнова Елена Олеговна", FirstName: "Елена", LastName: "Смирнова", MiddleName: "Олеговна",
			Email: "e.smirnova@students.local", Phone: "+7 901 000-00-02",
			GroupName: "БИ-21", Faculty: "Факультет ИТ", Specialty: "Бизнес-информатика", Course: 3,
			StudyForm: "Очная", EnrollmentDate: "2021-09-01T00:00:00", Status: "Учится",
		},
		{
			RefKey: "aaaaaaaa-0003-0000-0000-000000000003", Code: "СТУД-0003",
			Description: "Волков Артём Игоревич", FirstName: "Артём", LastName: "Волков", MiddleName: "Игоревич",
			Email: "a.volkov@students.local", Phone: "+7 901 000-00-03",
			GroupName: "ПИ-22", Faculty: "Факультет ИТ", Specialty: "Прикладная информатика", Course: 2,
			StudyForm: "Очная", EnrollmentDate: "2022-09-01T00:00:00", Status: "Учится",
		},
		{
			RefKey: "aaaaaaaa-0004-0000-0000-000000000004", Code: "СТУД-0004",
			Description: "Морозова Ольга Дмитриевна", FirstName: "Ольга", LastName: "Морозова", MiddleName: "Дмитриевна",
			Email: "o.morozova@students.local", Phone: "+7 901 000-00-04",
			GroupName: "ПИ-22", Faculty: "Факультет ИТ", Specialty: "Прикладная информатика", Course: 2,
			StudyForm: "Заочная", EnrollmentDate: "2022-09-01T00:00:00", Status: "Академический отпуск",
		},
		{
			RefKey: "aaaaaaaa-0005-0000-0000-000000000005", Code: "СТУД-0005",
			Description: "Новиков Павел Романович", FirstName: "Павел", LastName: "Новиков", MiddleName: "Романович",
			Email: "p.novikov@students.local", Phone: "+7 901 000-00-05",
			GroupName: "БИ-21", Faculty: "Факультет ИТ", Specialty: "Бизнес-информатика", Course: 3,
			StudyForm: "Очная", EnrollmentDate: "2021-09-01T00:00:00", Status: "Отчислен", DeletionMark: true,
		},
	}
}

// sampleDebts returns a small, stable academic-debt set referencing the
// sample students by group/name so a 1С-import demo lands realistic rows.
// Control forms use the Russian 1С labels the student_debts adapter maps to
// its wire codes; one row per status-worthy scenario across both groups.
func sampleDebts() []entities.ODataStudentDebt {
	return []entities.ODataStudentDebt{
		{
			RefKey:      "dddddddd-0001-0000-0000-000000000001",
			StudentRef:  "aaaaaaaa-0001-0000-0000-000000000001",
			StudentName: "Кузнецов Дмитрий Алексеевич", GroupName: "БИ-21",
			Discipline: "Базы данных", Semester: 3, ControlForm: "Экзамен",
		},
		{
			RefKey:      "dddddddd-0002-0000-0000-000000000002",
			StudentRef:  "aaaaaaaa-0002-0000-0000-000000000002",
			StudentName: "Смирнова Елена Олеговна", GroupName: "БИ-21",
			Discipline: "Математический анализ", Semester: 3, ControlForm: "Зачёт",
		},
		{
			RefKey:      "dddddddd-0003-0000-0000-000000000003",
			StudentRef:  "aaaaaaaa-0003-0000-0000-000000000003",
			StudentName: "Волков Артём Игоревич", GroupName: "ПИ-22",
			Discipline: "Программирование", Semester: 2, ControlForm: "Дифференцированный зачёт",
		},
		{
			RefKey:      "dddddddd-0004-0000-0000-000000000004",
			StudentRef:  "aaaaaaaa-0003-0000-0000-000000000003",
			StudentName: "Волков Артём Игоревич", GroupName: "ПИ-22",
			Discipline: "Архитектура ЭВМ", Semester: 2, ControlForm: "Курсовой проект",
		},
	}
}

func main() {
	port := os.Getenv("MOCK_1C_PORT")
	if port == "" {
		port = "9191"
	}

	mux := http.NewServeMux()
	// Single catch-all: 1С catalog paths contain Cyrillic, which net/http
	// percent-encodes on the client and decodes into r.URL.Path on the server.
	// Routing by substring avoids any ServeMux pattern-encoding subtleties.
	mux.HandleFunc("/", handle)

	addr := ":" + clean(port)
	// #nosec G706 -- dev-only mock; port comes from operator env and is sanitized via clean()
	log.Printf("mock-1c OData server listening on %s", addr)
	log.Printf("  GET /Catalog_Сотрудники              -> %d employees", len(sampleEmployees()))
	log.Printf("  GET /Catalog_Студенты                -> %d students", len(sampleStudents()))
	log.Printf("  GET /Catalog_АкадемическиеЗадолженности -> %d debts", len(sampleDebts()))

	// Explicit ReadHeaderTimeout (vs http.ListenAndServe) bounds slow-header
	// clients — satisfies gosec G114 without an inline suppression.
	srv := &http.Server{
		Addr:              addr,
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}
	if err := srv.ListenAndServe(); err != nil {
		log.Fatalf("mock-1c failed: %v", err)
	}
}

func handle(w http.ResponseWriter, r *http.Request) {
	logRequest(r)

	path := r.URL.Path
	switch {
	case strings.Contains(path, "$metadata"):
		writeMetadata(w)
	case strings.Contains(path, "Сотрудники"):
		writeEmployees(w, r)
	case strings.Contains(path, "Задолженности"):
		writeStudentDebts(w, r)
	case strings.Contains(path, "Студенты"):
		writeStudents(w, r)
	case path == "/" || path == "":
		// Ping: the client GETs the service root and only checks reachability.
		writeJSON(w, odataList[entities.ODataEmployee]{
			Metadata: "http://mock-1c/$metadata", Value: nil,
		})
	default:
		http.NotFound(w, r)
	}
}

// pagedOut returns true when a $skip beyond the first page is requested; the
// client paginates by 100, so any skip>0 means "no more rows" for our tiny set.
func pagedOut(r *http.Request) bool {
	skip := r.URL.Query().Get("$skip")
	return skip != "" && skip != "0"
}

func writeEmployees(w http.ResponseWriter, r *http.Request) {
	out := odataList[entities.ODataEmployee]{Metadata: "http://mock-1c/$metadata#Catalog_Сотрудники"}
	if !pagedOut(r) {
		out.Value = sampleEmployees()
	}
	writeJSON(w, out)
}

func writeStudents(w http.ResponseWriter, r *http.Request) {
	out := odataList[entities.ODataStudent]{Metadata: "http://mock-1c/$metadata#Catalog_Студенты"}
	if !pagedOut(r) {
		out.Value = sampleStudents()
	}
	writeJSON(w, out)
}

func writeStudentDebts(w http.ResponseWriter, r *http.Request) {
	out := odataList[entities.ODataStudentDebt]{Metadata: "http://mock-1c/$metadata#Catalog_АкадемическиеЗадолженности"}
	if !pagedOut(r) {
		out.Value = sampleDebts()
	}
	writeJSON(w, out)
}

func writeMetadata(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/xml; charset=utf-8")
	_, _ = w.Write([]byte(`<?xml version="1.0" encoding="utf-8"?>` +
		`<edmx:Edmx Version="1.0" xmlns:edmx="http://schemas.microsoft.com/ado/2007/06/edmx">` +
		`<edmx:DataServices><Schema Namespace="StandardODATA"/></edmx:DataServices></edmx:Edmx>`))
}

func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	if err := json.NewEncoder(w).Encode(v); err != nil {
		log.Printf("encode error: %v", err)
	}
}

func logRequest(r *http.Request) {
	user := "-"
	if h := r.Header.Get("Authorization"); strings.HasPrefix(h, "Basic ") {
		if raw, err := base64.StdEncoding.DecodeString(strings.TrimPrefix(h, "Basic ")); err == nil {
			if u, _, ok := strings.Cut(string(raw), ":"); ok {
				user = u
			}
		}
	}
	q := r.URL.RawQuery
	if q != "" {
		q = "?" + q
	}
	// #nosec G706 -- dev-only mock; request fields sanitized via clean()
	log.Printf("%s %s%s (auth=%s)", clean(r.Method), clean(r.URL.Path), clean(q), clean(user))
}

// clean strips control characters (CR/LF/etc.) so untrusted request and
// environment values cannot forge log lines (log-injection defense).
func clean(s string) string {
	return strings.Map(func(r rune) rune {
		if r < 0x20 || r == 0x7f {
			return -1
		}
		return r
	}, s)
}
