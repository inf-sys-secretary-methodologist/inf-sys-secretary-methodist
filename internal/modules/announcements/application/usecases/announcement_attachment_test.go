package usecases

import (
	"bytes"
	"context"
	"errors"
	"io"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/shared/infrastructure/storage"
)

// mockAttachmentStorage stub implements AttachmentStorage for tests.
type mockAttachmentStorage struct {
	uploaded    map[string][]byte
	uploadErr   error
	deletedKeys []string
	deleteErr   error
}

func newMockAttachmentStorage() *mockAttachmentStorage {
	return &mockAttachmentStorage{uploaded: make(map[string][]byte)}
}

func (m *mockAttachmentStorage) Upload(_ context.Context, key string, reader io.Reader, _ int64, _ string) (*storage.FileInfo, error) {
	if m.uploadErr != nil {
		return nil, m.uploadErr
	}
	data, err := io.ReadAll(reader)
	if err != nil {
		return nil, err
	}
	m.uploaded[key] = data
	return &storage.FileInfo{
		Key:        key,
		Size:       int64(len(data)),
		UploadedAt: time.Now(),
	}, nil
}

func (m *mockAttachmentStorage) Delete(_ context.Context, key string) error {
	if m.deleteErr != nil {
		return m.deleteErr
	}
	m.deletedKeys = append(m.deletedKeys, key)
	delete(m.uploaded, key)
	return nil
}

func (m *mockAttachmentStorage) GetPresignedURL(_ context.Context, key string, _ time.Duration) (string, error) {
	return "https://example.test/" + key, nil
}

func TestAddAttachment_Success(t *testing.T) {
	repo := NewMockAnnouncementRepository()
	storage := newMockAttachmentStorage()
	uc := NewAnnouncementUseCase(repo, createTestAuditLogger(), nil, nil)
	uc.SetAttachmentStorage(storage)

	// Seed an announcement so we can attach to it.
	ctx := context.Background()
	ann, err := uc.Create(ctx, 42, createDefaultRequest())
	require.NoError(t, err)

	body := "hello world"
	att, err := uc.AddAttachment(ctx, ann.ID, "doc.pdf", strings.NewReader(body), int64(len(body)), "application/pdf", 42)
	require.NoError(t, err)
	require.NotNil(t, att)

	assert.Equal(t, ann.ID, att.AnnouncementID)
	assert.Equal(t, "doc.pdf", att.FileName)
	assert.Equal(t, int64(len(body)), att.FileSize)
	assert.Equal(t, "application/pdf", att.MimeType)
	assert.Equal(t, int64(42), att.UploadedBy)
	assert.NotEmpty(t, att.FilePath, "file_path must be set")

	// Stored bytes match.
	stored, ok := storage.uploaded[att.FilePath]
	require.True(t, ok, "object must exist in storage at FilePath")
	assert.Equal(t, body, string(stored))
}

func TestAddAttachment_AnnouncementNotFound(t *testing.T) {
	repo := NewMockAnnouncementRepository()
	storage := newMockAttachmentStorage()
	uc := NewAnnouncementUseCase(repo, createTestAuditLogger(), nil, nil)
	uc.SetAttachmentStorage(storage)

	_, err := uc.AddAttachment(context.Background(), 9999, "x.pdf", bytes.NewReader([]byte("x")), 1, "application/pdf", 1)
	assert.ErrorIs(t, err, ErrAnnouncementNotFound)
	assert.Empty(t, storage.uploaded, "no upload should happen if announcement is not found")
}

func TestAddAttachment_StorageNotConfigured(t *testing.T) {
	repo := NewMockAnnouncementRepository()
	uc := NewAnnouncementUseCase(repo, createTestAuditLogger(), nil, nil)

	ann, err := uc.Create(context.Background(), 1, createDefaultRequest())
	require.NoError(t, err)

	_, err = uc.AddAttachment(context.Background(), ann.ID, "x.pdf", bytes.NewReader([]byte("x")), 1, "application/pdf", 1)
	assert.ErrorIs(t, err, ErrStorageNotConfigured)
}

func TestAddAttachment_StorageUploadFails(t *testing.T) {
	repo := NewMockAnnouncementRepository()
	stor := newMockAttachmentStorage()
	stor.uploadErr = errors.New("s3 down")
	uc := NewAnnouncementUseCase(repo, createTestAuditLogger(), nil, nil)
	uc.SetAttachmentStorage(stor)

	ann, err := uc.Create(context.Background(), 1, createDefaultRequest())
	require.NoError(t, err)

	_, err = uc.AddAttachment(context.Background(), ann.ID, "x.pdf", bytes.NewReader([]byte("x")), 1, "application/pdf", 1)
	require.Error(t, err)
	// repo should NOT have the attachment if storage failed
	atts, _ := repo.GetAttachments(context.Background(), ann.ID)
	assert.Empty(t, atts)
}

func TestRemoveAttachment_Success(t *testing.T) {
	repo := NewMockAnnouncementRepository()
	stor := newMockAttachmentStorage()
	uc := NewAnnouncementUseCase(repo, createTestAuditLogger(), nil, nil)
	uc.SetAttachmentStorage(stor)

	ctx := context.Background()
	ann, err := uc.Create(ctx, 1, createDefaultRequest())
	require.NoError(t, err)

	att, err := uc.AddAttachment(ctx, ann.ID, "x.pdf", strings.NewReader("body"), 4, "application/pdf", 1)
	require.NoError(t, err)
	require.NotZero(t, att.ID)

	err = uc.RemoveAttachment(ctx, att.ID, ann.ID, 1, "methodist")
	require.NoError(t, err)

	// Storage object deleted
	assert.Contains(t, stor.deletedKeys, att.FilePath)
	// Repo attachment removed
	atts, _ := repo.GetAttachments(ctx, ann.ID)
	assert.Empty(t, atts)
}

func TestRemoveAttachment_NotFound(t *testing.T) {
	repo := NewMockAnnouncementRepository()
	stor := newMockAttachmentStorage()
	uc := NewAnnouncementUseCase(repo, createTestAuditLogger(), nil, nil)
	uc.SetAttachmentStorage(stor)

	err := uc.RemoveAttachment(context.Background(), 9999, 1, 1, "methodist")
	assert.ErrorIs(t, err, ErrAttachmentNotFound)
}

// TestRemoveAttachment_AuthzGate pins v0.163.0 ADR-3 (#303 TIER 0):
// pre-fix RemoveAttachment took only (ctx, attachmentID) and never
// checked who was calling — anyone could delete any attachment
// system-wide. After fix the method verifies (a) the URL's
// announcement_id matches the stored attachment's announcement_id,
// (b) the actor is either the announcement author or a system_admin.
func TestRemoveAttachment_AuthzGate(t *testing.T) {
	type setup struct {
		authorID, attachmentID, actorID, urlAnnouncementID int64
		actorRole                                          string
		wantErr                                            error
	}

	tests := []struct {
		name string
		s    setup
	}{
		{
			name: "author removes their own attachment",
			s:    setup{authorID: 1, actorID: 1, actorRole: "methodist", wantErr: nil},
		},
		{
			name: "system_admin removes another author's attachment",
			s:    setup{authorID: 1, actorID: 99, actorRole: "system_admin", wantErr: nil},
		},
		{
			name: "stranger (non-author non-admin) blocked",
			s:    setup{authorID: 1, actorID: 2, actorRole: "methodist", wantErr: ErrAttachmentForbidden},
		},
		{
			name: "URL announcement_id mismatch blocked",
			s:    setup{authorID: 1, actorID: 1, actorRole: "methodist", urlAnnouncementID: 999, wantErr: ErrAttachmentForbidden},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			repo := NewMockAnnouncementRepository()
			stor := newMockAttachmentStorage()
			uc := NewAnnouncementUseCase(repo, createTestAuditLogger(), nil, nil)
			uc.SetAttachmentStorage(stor)

			ctx := context.Background()
			ann, err := uc.Create(ctx, tc.s.authorID, createDefaultRequest())
			require.NoError(t, err)

			att, err := uc.AddAttachment(ctx, ann.ID, "x.pdf", strings.NewReader("body"), 4, "application/pdf", tc.s.authorID)
			require.NoError(t, err)

			urlID := tc.s.urlAnnouncementID
			if urlID == 0 {
				urlID = ann.ID
			}

			err = uc.RemoveAttachment(ctx, att.ID, urlID, tc.s.actorID, tc.s.actorRole)
			if tc.s.wantErr == nil {
				assert.NoError(t, err)
			} else {
				assert.ErrorIs(t, err, tc.s.wantErr)
			}
		})
	}
}

// TestAddAttachment_RejectsOversizedAndBadMime pins v0.163.0 ADR-5
// (#303 TIER 1): pre-fix AddAttachment trusted the client-supplied
// Content-Type and had no size cap. After fix uploads above 10 MiB
// or with disallowed MIME types are rejected before reaching object
// storage.
func TestAddAttachment_RejectsOversizedAndBadMime(t *testing.T) {
	tests := []struct {
		name     string
		size     int64
		mimeType string
		wantErr  error
	}{
		{
			name:     "oversized payload",
			size:     attachmentMaxSize + 1,
			mimeType: "application/pdf",
			wantErr:  ErrAttachmentTooLarge,
		},
		{
			name:     "disallowed MIME (executable)",
			size:     1024,
			mimeType: "application/x-msdownload",
			wantErr:  ErrAttachmentMimeRejected,
		},
		{
			name:     "octet-stream loophole closed",
			size:     1024,
			mimeType: "application/octet-stream",
			wantErr:  ErrAttachmentMimeRejected,
		},
		{
			name:     "legitimate pdf within cap accepted",
			size:     2048,
			mimeType: "application/pdf",
			wantErr:  nil,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			repo := NewMockAnnouncementRepository()
			stor := newMockAttachmentStorage()
			uc := NewAnnouncementUseCase(repo, createTestAuditLogger(), nil, nil)
			uc.SetAttachmentStorage(stor)

			ctx := context.Background()
			ann, err := uc.Create(ctx, 1, createDefaultRequest())
			require.NoError(t, err)

			_, err = uc.AddAttachment(ctx, ann.ID, "x.bin", strings.NewReader("body"),
				tc.size, tc.mimeType, 1)
			if tc.wantErr == nil {
				assert.NoError(t, err)
			} else {
				assert.ErrorIs(t, err, tc.wantErr)
			}
		})
	}
}
