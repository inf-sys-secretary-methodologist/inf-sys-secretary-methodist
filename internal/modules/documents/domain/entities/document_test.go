package entities

import (
	"errors"
	"testing"
	"time"
)

func TestNewDocument(t *testing.T) {
	title := "Test Document"
	documentTypeID := int64(1)
	authorID := int64(42)

	doc := NewDocument(title, documentTypeID, authorID)

	if doc.Title != title {
		t.Errorf("expected title %q, got %q", title, doc.Title)
	}
	if doc.DocumentTypeID != documentTypeID {
		t.Errorf("expected document type ID %d, got %d", documentTypeID, doc.DocumentTypeID)
	}
	if doc.AuthorID != authorID {
		t.Errorf("expected author ID %d, got %d", authorID, doc.AuthorID)
	}
	if doc.Status != DocumentStatusDraft {
		t.Errorf("expected status %q, got %q", DocumentStatusDraft, doc.Status)
	}
	if doc.Version != 1 {
		t.Errorf("expected version 1, got %d", doc.Version)
	}
	if doc.IsPublic {
		t.Error("expected IsPublic to be false")
	}
	if doc.Importance != ImportanceNormal {
		t.Errorf("expected importance %q, got %q", ImportanceNormal, doc.Importance)
	}
	if doc.CreatedAt.IsZero() {
		t.Error("expected CreatedAt to be set")
	}
}

func TestDocument_SetFile(t *testing.T) {
	doc := NewDocument("Test", 1, 1)
	fileName := "document.pdf"
	filePath := "/files/document.pdf"
	mimeType := "application/pdf"
	fileSize := int64(1024)

	doc.SetFile(fileName, filePath, mimeType, fileSize)

	if doc.FileName == nil || *doc.FileName != fileName {
		t.Errorf("expected file name %q, got %v", fileName, doc.FileName)
	}
	if doc.FilePath == nil || *doc.FilePath != filePath {
		t.Errorf("expected file path %q, got %v", filePath, doc.FilePath)
	}
	if doc.MimeType == nil || *doc.MimeType != mimeType {
		t.Errorf("expected mime type %q, got %v", mimeType, doc.MimeType)
	}
	if doc.FileSize == nil || *doc.FileSize != fileSize {
		t.Errorf("expected file size %d, got %v", fileSize, doc.FileSize)
	}
}

func TestDocument_ClearFile(t *testing.T) {
	doc := NewDocument("Test", 1, 1)
	doc.SetFile("doc.pdf", "/files/doc.pdf", "application/pdf", 1024)

	doc.ClearFile()

	if doc.FileName != nil {
		t.Error("expected file name to be nil")
	}
	if doc.FilePath != nil {
		t.Error("expected file path to be nil")
	}
	if doc.MimeType != nil {
		t.Error("expected mime type to be nil")
	}
	if doc.FileSize != nil {
		t.Error("expected file size to be nil")
	}
}

func TestDocument_Register(t *testing.T) {
	doc := NewDocument("Test", 1, 1)
	regNumber := "REG-2024-001"

	doc.Register(regNumber)

	if doc.RegistrationNumber == nil || *doc.RegistrationNumber != regNumber {
		t.Errorf("expected registration number %q, got %v", regNumber, doc.RegistrationNumber)
	}
	if doc.RegistrationDate == nil {
		t.Error("expected registration date to be set")
	}
	if doc.Status != DocumentStatusRegistered {
		t.Errorf("expected status %q, got %q", DocumentStatusRegistered, doc.Status)
	}
}

func TestDocument_IsDraft(t *testing.T) {
	doc := NewDocument("Test", 1, 1)

	if !doc.IsDraft() {
		t.Error("expected new document to be draft")
	}

	doc.Register("REG-001")

	if doc.IsDraft() {
		t.Error("expected registered document to not be draft")
	}
}

func TestDocument_IsDeleted(t *testing.T) {
	doc := NewDocument("Test", 1, 1)

	if doc.IsDeleted() {
		t.Error("expected new document to not be deleted")
	}

	doc.SoftDelete()

	if !doc.IsDeleted() {
		t.Error("expected soft-deleted document to be deleted")
	}
}

func TestDocument_SoftDelete(t *testing.T) {
	doc := NewDocument("Test", 1, 1)

	doc.SoftDelete()

	if doc.DeletedAt == nil {
		t.Error("expected DeletedAt to be set")
	}
}

func TestDocument_Restore(t *testing.T) {
	doc := NewDocument("Test", 1, 1)
	doc.SoftDelete()

	doc.Restore()

	if doc.DeletedAt != nil {
		t.Error("expected DeletedAt to be nil after restore")
	}
}

func TestDocument_HasFile(t *testing.T) {
	doc := NewDocument("Test", 1, 1)

	if doc.HasFile() {
		t.Error("expected new document to not have file")
	}

	doc.SetFile("doc.pdf", "/files/doc.pdf", "application/pdf", 1024)

	if !doc.HasFile() {
		t.Error("expected document with file to return true")
	}
}

func TestDocument_HasFile_EmptyPath(t *testing.T) {
	doc := NewDocument("Test", 1, 1)
	emptyPath := ""
	doc.FilePath = &emptyPath

	if doc.HasFile() {
		t.Error("expected document with empty path to not have file")
	}
}

func TestDocumentStatusConstants(t *testing.T) {
	tests := []struct {
		name     string
		status   DocumentStatus
		expected string
	}{
		{"draft", DocumentStatusDraft, "draft"},
		{"registered", DocumentStatusRegistered, "registered"},
		{"routing", DocumentStatusRouting, "routing"},
		{"approval", DocumentStatusApproval, "approval"},
		{"approved", DocumentStatusApproved, "approved"},
		{"rejected", DocumentStatusRejected, "rejected"},
		{"execution", DocumentStatusExecution, "execution"},
		{"executed", DocumentStatusExecuted, "executed"},
		{"archived", DocumentStatusArchived, "archived"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if string(tt.status) != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, tt.status)
			}
		})
	}
}

func TestDocumentImportanceConstants(t *testing.T) {
	tests := []struct {
		name       string
		importance DocumentImportance
		expected   string
	}{
		{"low", ImportanceLow, "low"},
		{"normal", ImportanceNormal, "normal"},
		{"high", ImportanceHigh, "high"},
		{"urgent", ImportanceUrgent, "urgent"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if string(tt.importance) != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, tt.importance)
			}
		})
	}
}

// --- Permission tests ---

func TestDocumentPermission_IsExpired(t *testing.T) {
	// Not expired: no expiry
	p := &DocumentPermission{Permission: PermissionRead}
	if p.IsExpired() {
		t.Error("permission without expiry should not be expired")
	}

	// Expired: past time
	past := time.Now().Add(-1 * time.Hour)
	p2 := &DocumentPermission{Permission: PermissionRead, ExpiresAt: &past}
	if !p2.IsExpired() {
		t.Error("permission with past expiry should be expired")
	}

	// Not expired: future time
	future := time.Now().Add(1 * time.Hour)
	p3 := &DocumentPermission{Permission: PermissionRead, ExpiresAt: &future}
	if p3.IsExpired() {
		t.Error("permission with future expiry should not be expired")
	}
}

func TestDocumentPermission_IsValid(t *testing.T) {
	userID := int64(1)
	role := RoleAdmin

	// Valid: has user, not expired
	p := &DocumentPermission{UserID: &userID, Permission: PermissionRead}
	if !p.IsValid() {
		t.Error("permission with user should be valid")
	}

	// Valid: has role, not expired
	p2 := &DocumentPermission{Role: &role, Permission: PermissionRead}
	if !p2.IsValid() {
		t.Error("permission with role should be valid")
	}

	// Invalid: expired
	past := time.Now().Add(-1 * time.Hour)
	p3 := &DocumentPermission{UserID: &userID, Permission: PermissionRead, ExpiresAt: &past}
	if p3.IsValid() {
		t.Error("expired permission should be invalid")
	}

	// Invalid: no user or role
	p4 := &DocumentPermission{Permission: PermissionRead}
	if p4.IsValid() {
		t.Error("permission without user or role should be invalid")
	}
}

func TestDocumentPermission_CanRead(t *testing.T) {
	tests := []struct {
		name       string
		permission PermissionLevel
		want       bool
	}{
		{"read can read", PermissionRead, true},
		{"write can read", PermissionWrite, true},
		{"delete can read", PermissionDelete, true},
		{"admin can read", PermissionAdmin, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &DocumentPermission{Permission: tt.permission}
			if got := p.CanRead(); got != tt.want {
				t.Errorf("CanRead() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDocumentPermission_CanWrite(t *testing.T) {
	tests := []struct {
		name       string
		permission PermissionLevel
		want       bool
	}{
		{"read cannot write", PermissionRead, false},
		{"write can write", PermissionWrite, true},
		{"delete can write", PermissionDelete, true},
		{"admin can write", PermissionAdmin, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &DocumentPermission{Permission: tt.permission}
			if got := p.CanWrite(); got != tt.want {
				t.Errorf("CanWrite() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDocumentPermission_CanDelete(t *testing.T) {
	tests := []struct {
		name       string
		permission PermissionLevel
		want       bool
	}{
		{"read cannot delete", PermissionRead, false},
		{"write cannot delete", PermissionWrite, false},
		{"delete can delete", PermissionDelete, true},
		{"admin can delete", PermissionAdmin, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &DocumentPermission{Permission: tt.permission}
			if got := p.CanDelete(); got != tt.want {
				t.Errorf("CanDelete() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDocumentPermission_IsAdmin(t *testing.T) {
	tests := []struct {
		name       string
		permission PermissionLevel
		want       bool
	}{
		{"read is not admin", PermissionRead, false},
		{"write is not admin", PermissionWrite, false},
		{"delete is not admin", PermissionDelete, false},
		{"admin is admin", PermissionAdmin, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &DocumentPermission{Permission: tt.permission}
			if got := p.IsAdmin(); got != tt.want {
				t.Errorf("IsAdmin() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPermissionLevelConstants(t *testing.T) {
	if string(PermissionRead) != "read" {
		t.Errorf("expected 'read', got %q", PermissionRead)
	}
	if string(PermissionWrite) != "write" {
		t.Errorf("expected 'write', got %q", PermissionWrite)
	}
	if string(PermissionDelete) != "delete" {
		t.Errorf("expected 'delete', got %q", PermissionDelete)
	}
	if string(PermissionAdmin) != "admin" {
		t.Errorf("expected 'admin', got %q", PermissionAdmin)
	}
}

func TestUserRoleConstants(t *testing.T) {
	if string(RoleAdmin) != "admin" {
		t.Errorf("expected 'admin', got %q", RoleAdmin)
	}
	if string(RoleSecretary) != "secretary" {
		t.Errorf("expected 'secretary', got %q", RoleSecretary)
	}
	if string(RoleMethodist) != "methodist" {
		t.Errorf("expected 'methodist', got %q", RoleMethodist)
	}
	if string(RoleTeacher) != "teacher" {
		t.Errorf("expected 'teacher', got %q", RoleTeacher)
	}
	if string(RoleStudent) != "student" {
		t.Errorf("expected 'student', got %q", RoleStudent)
	}
}

// --- PublicLink tests ---

func TestGenerateToken(t *testing.T) {
	token, err := GenerateToken()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(token) != 64 { // 32 bytes = 64 hex chars
		t.Errorf("expected token length 64, got %d", len(token))
	}

	// Check uniqueness
	token2, err := GenerateToken()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if token == token2 {
		t.Error("expected unique tokens")
	}
}

func TestPublicLink_IsExpired(t *testing.T) {
	// No expiry
	l := &PublicLink{}
	if l.IsExpired() {
		t.Error("link without expiry should not be expired")
	}

	// Past expiry
	past := time.Now().Add(-1 * time.Hour)
	l2 := &PublicLink{ExpiresAt: &past}
	if !l2.IsExpired() {
		t.Error("link with past expiry should be expired")
	}

	// Future expiry
	future := time.Now().Add(1 * time.Hour)
	l3 := &PublicLink{ExpiresAt: &future}
	if l3.IsExpired() {
		t.Error("link with future expiry should not be expired")
	}
}

func TestPublicLink_IsUsageLimitReached(t *testing.T) {
	// No limit
	l := &PublicLink{}
	if l.IsUsageLimitReached() {
		t.Error("link without limit should not be reached")
	}

	// Under limit
	maxUses := 10
	l2 := &PublicLink{MaxUses: &maxUses, UseCount: 5}
	if l2.IsUsageLimitReached() {
		t.Error("link under limit should not be reached")
	}

	// At limit
	l3 := &PublicLink{MaxUses: &maxUses, UseCount: 10}
	if !l3.IsUsageLimitReached() {
		t.Error("link at limit should be reached")
	}

	// Over limit
	l4 := &PublicLink{MaxUses: &maxUses, UseCount: 15}
	if !l4.IsUsageLimitReached() {
		t.Error("link over limit should be reached")
	}
}

func TestPublicLink_IsValid(t *testing.T) {
	// Valid: active, not expired, not at limit
	l := &PublicLink{IsActive: true}
	if !l.IsValid() {
		t.Error("active link should be valid")
	}

	// Invalid: not active
	l2 := &PublicLink{IsActive: false}
	if l2.IsValid() {
		t.Error("inactive link should not be valid")
	}

	// Invalid: expired
	past := time.Now().Add(-1 * time.Hour)
	l3 := &PublicLink{IsActive: true, ExpiresAt: &past}
	if l3.IsValid() {
		t.Error("expired link should not be valid")
	}

	// Invalid: usage limit reached
	maxUses := 1
	l4 := &PublicLink{IsActive: true, MaxUses: &maxUses, UseCount: 1}
	if l4.IsValid() {
		t.Error("link at usage limit should not be valid")
	}
}

func TestPublicLink_HasPassword(t *testing.T) {
	// No password
	l := &PublicLink{}
	if l.HasPassword() {
		t.Error("link without password should return false")
	}

	// Empty password
	empty := ""
	l2 := &PublicLink{PasswordHash: &empty}
	if l2.HasPassword() {
		t.Error("link with empty password hash should return false")
	}

	// Has password
	hash := "somehash"
	l3 := &PublicLink{PasswordHash: &hash}
	if !l3.HasPassword() {
		t.Error("link with password hash should return true")
	}
}

func TestPublicLink_CanDownload(t *testing.T) {
	l := &PublicLink{Permission: PublicLinkRead}
	if l.CanDownload() {
		t.Error("read-only link should not allow download")
	}

	l2 := &PublicLink{Permission: PublicLinkDownload}
	if !l2.CanDownload() {
		t.Error("download link should allow download")
	}
}

func TestPublicLink_IncrementUseCount(t *testing.T) {
	l := &PublicLink{UseCount: 0}
	l.IncrementUseCount()
	if l.UseCount != 1 {
		t.Errorf("expected use count 1, got %d", l.UseCount)
	}
	l.IncrementUseCount()
	if l.UseCount != 2 {
		t.Errorf("expected use count 2, got %d", l.UseCount)
	}
}

func TestPublicLink_Deactivate(t *testing.T) {
	l := &PublicLink{IsActive: true}
	l.Deactivate()
	if l.IsActive {
		t.Error("expected link to be deactivated")
	}
}

func TestPublicLink_Activate(t *testing.T) {
	l := &PublicLink{IsActive: false}
	l.Activate()
	if !l.IsActive {
		t.Error("expected link to be activated")
	}
}

func TestPublicLinkPermissionConstants(t *testing.T) {
	if string(PublicLinkRead) != "read" {
		t.Errorf("expected 'read', got %q", PublicLinkRead)
	}
	if string(PublicLinkDownload) != "download" {
		t.Errorf("expected 'download', got %q", PublicLinkDownload)
	}
}

// --- DocumentType / DocumentVersion tests ---

func TestNewDocumentVersion(t *testing.T) {
	doc := NewDocument("Test Doc", 1, 42)
	doc.ID = 100
	doc.Version = 3
	subject := "test subject"
	doc.Subject = &subject
	content := "some content"
	doc.Content = &content

	dv := NewDocumentVersion(doc, 99, "Updated title")

	if dv.DocumentID != doc.ID {
		t.Errorf("expected document ID %d, got %d", doc.ID, dv.DocumentID)
	}
	if dv.Version != doc.Version {
		t.Errorf("expected version %d, got %d", doc.Version, dv.Version)
	}
	if dv.Title == nil || *dv.Title != doc.Title {
		t.Errorf("expected title %q, got %v", doc.Title, dv.Title)
	}
	if dv.Subject == nil || *dv.Subject != *doc.Subject {
		t.Errorf("expected subject %q, got %v", *doc.Subject, dv.Subject)
	}
	if dv.Content == nil || *dv.Content != *doc.Content {
		t.Errorf("expected content %q, got %v", *doc.Content, dv.Content)
	}
	if dv.ChangedBy != 99 {
		t.Errorf("expected changed by 99, got %d", dv.ChangedBy)
	}
	if dv.ChangeDescription == nil || *dv.ChangeDescription != "Updated title" {
		t.Errorf("expected description 'Updated title', got %v", dv.ChangeDescription)
	}
	if dv.Status == nil || *dv.Status != string(DocumentStatusDraft) {
		t.Errorf("expected status 'draft', got %v", dv.Status)
	}
	if dv.CreatedAt.IsZero() {
		t.Error("expected CreatedAt to be set")
	}
}

func TestNewDocumentVersion_WithFile(t *testing.T) {
	doc := NewDocument("Test", 1, 1)
	doc.ID = 10
	doc.SetFile("file.pdf", "/path/file.pdf", "application/pdf", 2048)

	dv := NewDocumentVersion(doc, 1, "added file")

	if dv.FileName == nil || *dv.FileName != "file.pdf" {
		t.Errorf("expected file name 'file.pdf', got %v", dv.FileName)
	}
	if dv.FilePath == nil || *dv.FilePath != "/path/file.pdf" {
		t.Errorf("expected file path '/path/file.pdf', got %v", dv.FilePath)
	}
	if dv.FileSize == nil || *dv.FileSize != 2048 {
		t.Errorf("expected file size 2048, got %v", dv.FileSize)
	}
	if dv.MimeType == nil || *dv.MimeType != "application/pdf" {
		t.Errorf("expected mime type 'application/pdf', got %v", dv.MimeType)
	}
	if dv.StorageKey == nil || *dv.StorageKey != "/path/file.pdf" {
		t.Errorf("expected storage key '/path/file.pdf', got %v", dv.StorageKey)
	}
}

// TestDocument_CanBeEditedBy table-pins the edit-permission rule for
// every role in the system. Without this method the Update use case
// has to assemble the rule by hand at every call site, which is how
// the gap that this v0.108.2 release closes was introduced — teachers
// could update any document because no caller checked AuthorID.
//
// Rule (mirrors the audit report):
//   - methodist / academic_secretary / system_admin: edit any document
//   - teacher: only own documents (userID == AuthorID)
//   - student: never (defense-in-depth — handler also blocks via
//     RequireNonStudent middleware introduced in v0.105.3)
//   - any unknown role: deny
func TestDocument_CanBeEditedBy(t *testing.T) {
	const authorID int64 = 100
	const otherUserID int64 = 200
	doc := &Document{ID: 1, AuthorID: authorID}

	cases := []struct {
		name      string
		userID    int64
		role      UserRole
		wantAllow bool
	}{
		{"methodist edits another author's doc", otherUserID, RoleMethodist, true},
		{"academic_secretary edits another author's doc", otherUserID, "academic_secretary", true},
		{"system_admin edits another author's doc", otherUserID, "system_admin", true},
		{"teacher edits own doc", authorID, RoleTeacher, true},
		{"teacher edits another author's doc -> denied", otherUserID, RoleTeacher, false},
		{"student edits own doc -> denied (defense in depth)", authorID, RoleStudent, false},
		{"unknown role -> denied", otherUserID, "alien", false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := doc.CanBeEditedBy(tc.userID, tc.role)
			if tc.wantAllow && err != nil {
				t.Errorf("expected allow, got error: %v", err)
			}
			if !tc.wantAllow && err == nil {
				t.Errorf("expected denial, got nil error")
			}
		})
	}
}

// TestDocument_CanBeEditedBy_DenialIsErrEditDenied — denial must
// surface ErrDocumentEditDenied via errors.Is so handlers can map to
// a stable HTTP code (403) without string parsing.
func TestDocument_CanBeEditedBy_DenialIsErrEditDenied(t *testing.T) {
	doc := &Document{ID: 1, AuthorID: 1}
	err := doc.CanBeEditedBy(2, RoleTeacher)
	if !errors.Is(err, ErrDocumentEditDenied) {
		t.Errorf("expected errors.Is(err, ErrDocumentEditDenied) == true, got %v", err)
	}
}
