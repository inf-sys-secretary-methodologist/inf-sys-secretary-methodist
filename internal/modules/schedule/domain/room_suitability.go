package domain

import "strings"

// AllowedRoomTypesForLesson maps a lesson type (identified by its human short
// name, e.g. "Лек", "Лаб") to the classroom types suitable for it. It bridges
// the human-named lesson types entered by the methodist to the machine
// classroom-type codes used by the classroom catalog:
// 'lecture','practice','lab','computer','conference','sports'.
//
// An empty result means "no restriction" — any room type fits. This is the
// scheduling business rule the CSP solver checks via Variable.AllowedRoomTypes;
// unknown lesson types deliberately fall back to "any room" so the generator
// never blocks a lesson it cannot classify.
func AllowedRoomTypesForLesson(lessonShortName string) []string {
	n := strings.ToLower(strings.TrimSpace(lessonShortName))
	switch {
	case strings.HasPrefix(n, "лек"): // Лекция
		return []string{"lecture"}
	case strings.HasPrefix(n, "практ"), strings.HasPrefix(n, "сем"): // Практика, Семинар
		return []string{"practice"}
	case strings.HasPrefix(n, "лаб"): // Лабораторная работа
		return []string{"lab", "computer"}
	default: // Консультация, Экзамен, Зачет, неизвестное — любая аудитория
		return nil
	}
}
