package entities

// SyncStatus represents the status of a synchronization operation
type SyncStatus string

// SyncStatus values.
const (
	SyncStatusPending    SyncStatus = "pending"
	SyncStatusInProgress SyncStatus = "in_progress"
	SyncStatusCompleted  SyncStatus = "completed"
	SyncStatusFailed     SyncStatus = "failed"
	SyncStatusCancelled  SyncStatus = "canceled"
)

// SyncDirection represents the direction of data synchronization
type SyncDirection string

// SyncDirection values.
const (
	SyncDirectionImport SyncDirection = "import" // From 1C to local
	SyncDirectionExport SyncDirection = "export" // From local to 1C
	SyncDirectionBoth   SyncDirection = "both"   // Bidirectional
)

// SyncEntityType represents the type of entity being synchronized
type SyncEntityType string

// SyncEntityType values.
const (
	SyncEntityEmployee SyncEntityType = "employee"
	SyncEntityStudent  SyncEntityType = "student"
	SyncEntityFinance  SyncEntityType = "finance"
)

// ConflictResolution represents how a conflict was resolved
type ConflictResolution string

// ConflictResolution values.
const (
	ConflictResolutionPending     ConflictResolution = "pending"
	ConflictResolutionUseLocal    ConflictResolution = "use_local"
	ConflictResolutionUseExternal ConflictResolution = "use_external"
	ConflictResolutionMerge       ConflictResolution = "merge"
	ConflictResolutionSkip        ConflictResolution = "skip"
)
