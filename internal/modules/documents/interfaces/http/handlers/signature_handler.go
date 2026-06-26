package http

import (
	"context"
	"encoding/base64"
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/documents/application/usecases"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/documents/domain/entities"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/documents/domain/repositories"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/shared/infrastructure/http/response"
)

// SignDocumentPort applies a signature to a document.
type SignDocumentPort interface {
	Execute(ctx context.Context, documentID, signerID int64, signerName, signerRole string) (*entities.DocumentSignature, error)
}

// ListSignaturesPort lists a document's signatures.
type ListSignaturesPort interface {
	Execute(ctx context.Context, documentID int64) ([]*entities.DocumentSignature, error)
}

// VerifySignaturePort verifies one stored signature.
type VerifySignaturePort interface {
	Execute(ctx context.Context, signatureID int64) (usecases.SignatureVerdict, error)
}

// SignerNameResolver resolves a signer's display name (denormalized into the
// signature for a durable legal trail). Implemented in main.go over the users
// repository.
type SignerNameResolver interface {
	FullName(ctx context.Context, userID int64) (string, error)
}

// SignatureHandler exposes the document e-signature endpoints (#140).
type SignatureHandler struct {
	sign   SignDocumentPort
	list   ListSignaturesPort
	verify VerifySignaturePort
	names  SignerNameResolver
}

// NewSignatureHandler wires the handler; panics on any nil port (fail-closed DI).
func NewSignatureHandler(sign SignDocumentPort, list ListSignaturesPort, verify VerifySignaturePort, names SignerNameResolver) *SignatureHandler {
	if sign == nil || list == nil || verify == nil || names == nil {
		panic("documents: NewSignatureHandler requires non-nil ports")
	}
	return &SignatureHandler{sign: sign, list: list, verify: verify, names: names}
}

// RegisterSignatureRoutes mounts the signature endpoints on rg.
func RegisterSignatureRoutes(rg *gin.RouterGroup, h *SignatureHandler) {
	rg.POST("/documents/:id/sign", h.Sign)
	rg.GET("/documents/:id/signatures", h.List)
	rg.GET("/documents/:id/signatures/:sigId/verify", h.Verify)
}

// Sign applies the actor's signature to the document.
func (h *SignatureHandler) Sign(c *gin.Context) {
	id, err := parseDocID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{errorKey: "invalid document id"})
		return
	}
	userID, role, ok := readActor(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{errorKey: "unauthorized"})
		return
	}
	name, err := h.names.FullName(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{errorKey: "could not resolve signer identity"})
		return
	}
	sig, err := h.sign.Execute(c.Request.Context(), id, userID, name, string(role))
	if err != nil {
		mapSignatureError(c, err)
		return
	}
	c.JSON(http.StatusCreated, response.Success(toSignatureDTO(sig)))
}

// List returns all signatures applied to a document.
func (h *SignatureHandler) List(c *gin.Context) {
	id, err := parseDocID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{errorKey: "invalid document id"})
		return
	}
	sigs, err := h.list.Execute(c.Request.Context(), id)
	if err != nil {
		mapSignatureError(c, err)
		return
	}
	dtos := make([]signatureDTO, 0, len(sigs))
	for _, s := range sigs {
		dtos = append(dtos, toSignatureDTO(s))
	}
	c.JSON(http.StatusOK, response.Success(dtos))
}

// Verify re-checks a stored signature against the document's current state.
func (h *SignatureHandler) Verify(c *gin.Context) {
	sigID, err := strconv.ParseInt(c.Param("sigId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{errorKey: "invalid signature id"})
		return
	}
	verdict, err := h.verify.Execute(c.Request.Context(), sigID)
	if err != nil {
		mapSignatureError(c, err)
		return
	}
	c.JSON(http.StatusOK, response.Success(toVerdictDTO(verdict)))
}

// mapSignatureError maps domain/repository errors to stable HTTP codes.
func mapSignatureError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, repositories.ErrSignatureNotFound), errors.Is(err, repositories.ErrDocumentNotFound):
		c.JSON(http.StatusNotFound, gin.H{errorKey: "not found"})
	case errors.Is(err, entities.ErrDocumentEditDenied):
		c.JSON(http.StatusForbidden, gin.H{errorKey: "not authorized to sign this document"})
	case errors.Is(err, entities.ErrInvalidDocumentSignature), errors.Is(err, entities.ErrInvalidSignatureAlgorithm):
		c.JSON(http.StatusUnprocessableEntity, gin.H{errorKey: "invalid signature"})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{errorKey: "internal error"})
	}
}

// --- DTOs -----------------------------------------------------------------

type signatureDTO struct {
	ID              int64     `json:"id"`
	DocumentID      int64     `json:"document_id"`
	DocumentVersion int       `json:"document_version"`
	SignerID        int64     `json:"signer_id"`
	SignerName      string    `json:"signer_name"`
	Algorithm       string    `json:"algorithm"`
	DigestHex       string    `json:"digest_hex"`
	SignatureBase64 string    `json:"signature_base64"`
	CertificatePEM  string    `json:"certificate_pem"`
	SignedAt        time.Time `json:"signed_at"`
}

func toSignatureDTO(s *entities.DocumentSignature) signatureDTO {
	return signatureDTO{
		ID:              s.ID,
		DocumentID:      s.DocumentID,
		DocumentVersion: s.DocumentVersion,
		SignerID:        s.SignerID,
		SignerName:      s.SignerName,
		Algorithm:       s.Algorithm.String(),
		DigestHex:       s.DigestHex,
		SignatureBase64: base64.StdEncoding.EncodeToString(s.SignatureDER),
		CertificatePEM:  s.CertificatePEM,
		SignedAt:        s.SignedAt,
	}
}

type verdictDTO struct {
	SignatureID    int64  `json:"signature_id"`
	Valid          bool   `json:"valid"`
	DigestMatch    bool   `json:"digest_match"`
	CryptoValid    bool   `json:"crypto_valid"`
	VersionChanged bool   `json:"version_changed"`
	Status         string `json:"status"`
}

func toVerdictDTO(v usecases.SignatureVerdict) verdictDTO {
	status := "crypto_invalid"
	if v.Valid {
		status = "valid"
	} else if !v.DigestMatch {
		status = "document_modified"
	}
	return verdictDTO{
		SignatureID:    v.SignatureID,
		Valid:          v.Valid,
		DigestMatch:    v.DigestMatch,
		CryptoValid:    v.CryptoValid,
		VersionChanged: v.VersionChanged,
		Status:         status,
	}
}
