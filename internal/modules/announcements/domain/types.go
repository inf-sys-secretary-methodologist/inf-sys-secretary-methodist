// Package domain contains announcement domain types and enums.
package domain

// AnnouncementStatus represents the status of an announcement.
type AnnouncementStatus string

// AnnouncementStatus values.
const (
	AnnouncementStatusDraft     AnnouncementStatus = "draft"
	AnnouncementStatusPublished AnnouncementStatus = "published"
	AnnouncementStatusArchived  AnnouncementStatus = "archived"
)

// IsValid checks if the announcement status is valid.
func (s AnnouncementStatus) IsValid() bool {
	switch s {
	case AnnouncementStatusDraft, AnnouncementStatusPublished, AnnouncementStatusArchived:
		return true
	}
	return false
}

// AnnouncementPriority represents the priority level of an announcement.
type AnnouncementPriority string

// AnnouncementPriority values.
const (
	AnnouncementPriorityLow    AnnouncementPriority = "low"
	AnnouncementPriorityNormal AnnouncementPriority = "normal"
	AnnouncementPriorityHigh   AnnouncementPriority = "high"
	AnnouncementPriorityUrgent AnnouncementPriority = "urgent"
)

// IsValid checks if the announcement priority is valid.
func (p AnnouncementPriority) IsValid() bool {
	switch p {
	case AnnouncementPriorityLow, AnnouncementPriorityNormal, AnnouncementPriorityHigh, AnnouncementPriorityUrgent:
		return true
	}
	return false
}

// TargetAudience represents who can see the announcement.
type TargetAudience string

// TargetAudience values.
const (
	TargetAudienceAll      TargetAudience = "all"
	TargetAudienceStudents TargetAudience = "students"
	TargetAudienceTeachers TargetAudience = "teachers"
	TargetAudienceStaff    TargetAudience = "staff"
	TargetAudienceAdmins   TargetAudience = "admins"
)

// IsValid checks if the target audience is valid.
func (t TargetAudience) IsValid() bool {
	switch t {
	case TargetAudienceAll, TargetAudienceStudents, TargetAudienceTeachers, TargetAudienceStaff, TargetAudienceAdmins:
		return true
	}
	return false
}

// CanAccessAudience reports whether a caller of the given role is
// allowed to receive announcements addressed к the given audience.
//
// v0.163.0 ADR-2 (#303 TIER 0): pre-fix the handler derived audience
// from the client (?audience=admins) and trusted it. A student could
// request `?audience=admins` to read admin-broadcasts. This function
// is the canonical access matrix consulted at the handler boundary
// before any repo query runs.
//
//   - student   → all, students
//   - teacher   → all, teachers
//   - methodist / academic_secretary → all, staff
//   - system_admin → all five audiences
func CanAccessAudience(role string, audience TargetAudience) bool {
	switch role {
	case "system_admin":
		return audience.IsValid()
	case "methodist", "academic_secretary":
		return audience == TargetAudienceAll || audience == TargetAudienceStaff
	case "teacher":
		return audience == TargetAudienceAll || audience == TargetAudienceTeachers
	case "student":
		return audience == TargetAudienceAll || audience == TargetAudienceStudents
	default:
		return audience == TargetAudienceAll
	}
}
