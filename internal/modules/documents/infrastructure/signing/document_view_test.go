package signing

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/documents/application/usecases"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/documents/domain/entities"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/shared/infrastructure/storage"
)

type fakeDocReader struct {
	doc *entities.Document
	err error
}

func (f *fakeDocReader) GetByID(_ context.Context, _ int64) (*entities.Document, error) {
	return f.doc, f.err
}

type fakeDownloader struct {
	body []byte
	err  error
}

func (f *fakeDownloader) Download(_ context.Context, _ string) (io.ReadCloser, *storage.FileInfo, error) {
	if f.err != nil {
		return nil, nil, f.err
	}
	return io.NopCloser(bytes.NewReader(f.body)), &storage.FileInfo{}, nil
}

func strptr(s string) *string { return &s }

func TestDocumentView_TextContent(t *testing.T) {
	content := "договор о практике"
	doc := &entities.Document{ID: 42, Version: 2, Content: strptr(content)}
	view := NewDocumentView(&fakeDocReader{doc: doc}, &fakeDownloader{})

	version, hashHex, err := view.GetForSigning(context.Background(), 42)
	require.NoError(t, err)
	assert.Equal(t, 2, version)
	want := sha256.Sum256([]byte(content))
	assert.Equal(t, hex.EncodeToString(want[:]), hashHex)
}

func TestDocumentView_FileBytes(t *testing.T) {
	fileBytes := []byte("%PDF-1.7 binary body")
	doc := &entities.Document{ID: 42, Version: 5, FileName: strptr("doc.pdf"), FilePath: strptr("documents/1/doc.pdf")}
	view := NewDocumentView(&fakeDocReader{doc: doc}, &fakeDownloader{body: fileBytes})

	version, hashHex, err := view.GetForSigning(context.Background(), 42)
	require.NoError(t, err)
	assert.Equal(t, 5, version)
	want := sha256.Sum256(fileBytes)
	assert.Equal(t, hex.EncodeToString(want[:]), hashHex)
}

func TestDocumentView_NotFoundPropagates(t *testing.T) {
	view := NewDocumentView(&fakeDocReader{err: usecases.ErrDocumentNotFound}, &fakeDownloader{})
	_, _, err := view.GetForSigning(context.Background(), 99)
	assert.ErrorIs(t, err, usecases.ErrDocumentNotFound)
}
