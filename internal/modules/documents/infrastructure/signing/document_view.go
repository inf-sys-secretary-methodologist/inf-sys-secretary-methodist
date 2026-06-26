package signing

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/documents/domain/entities"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/shared/infrastructure/storage"
)

// documentReader is the narrow read port DocumentView needs over the document
// repository. Returns repositories.ErrDocumentNotFound when absent.
type documentReader interface {
	GetByID(ctx context.Context, id int64) (*entities.Document, error)
}

// objectDownloader is the narrow port over object storage (S3/MinIO).
type objectDownloader interface {
	Download(ctx context.Context, key string) (io.ReadCloser, *storage.FileInfo, error)
}

// DocumentView adapts the document repository + object storage to the
// usecases.DocumentSigningView port: it exposes a document's current version
// and a SHA-256 hex of its canonical body (uploaded file bytes if present,
// otherwise the text content).
type DocumentView struct {
	docs documentReader
	objs objectDownloader
}

// NewDocumentView wires the adapter.
func NewDocumentView(docs documentReader, objs objectDownloader) *DocumentView {
	if docs == nil || objs == nil {
		panic("documents/signing: NewDocumentView requires non-nil reader and downloader")
	}
	return &DocumentView{docs: docs, objs: objs}
}

// GetForSigning returns the document's version and the SHA-256 hex of its body.
func (v *DocumentView) GetForSigning(ctx context.Context, documentID int64) (int, string, error) {
	doc, err := v.docs.GetByID(ctx, documentID)
	if err != nil {
		return 0, "", err
	}

	body, err := v.bodyBytes(ctx, doc)
	if err != nil {
		return 0, "", err
	}
	sum := sha256.Sum256(body)
	return doc.Version, hex.EncodeToString(sum[:]), nil
}

// bodyBytes returns the canonical body of a document: the uploaded file bytes
// when present, otherwise the text content (possibly empty).
func (v *DocumentView) bodyBytes(ctx context.Context, doc *entities.Document) ([]byte, error) {
	if doc.HasFile() && doc.FilePath != nil {
		rc, _, err := v.objs.Download(ctx, *doc.FilePath)
		if err != nil {
			return nil, fmt.Errorf("documents/signing: download body: %w", err)
		}
		defer func() { _ = rc.Close() }()
		body, err := io.ReadAll(rc)
		if err != nil {
			return nil, fmt.Errorf("documents/signing: read body: %w", err)
		}
		return body, nil
	}
	if doc.Content != nil {
		return []byte(*doc.Content), nil
	}
	return nil, nil
}
