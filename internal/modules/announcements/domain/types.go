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
