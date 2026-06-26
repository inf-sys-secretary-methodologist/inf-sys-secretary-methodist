// Package signing provides the cryptographic engine and at-rest key custody for
// the document electronic-signature feature (#140). Each user gets a per-user
// ECDSA P-256 keypair wrapped in a self-signed X.509 certificate; the private
// key is encrypted with AES-256-GCM before it touches the database.
package signing

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"database/sql"
	"encoding/pem"
	"errors"
	"fmt"
	"math/big"
	"time"

	"github.com/lib/pq"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/documents/domain/entities"
	appcrypto "github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/shared/infrastructure/crypto"
)

// certValidity is how long an issued signer certificate stays valid.
const certValidity = 10 * 365 * 24 * time.Hour

// Service signs document digests with per-user ECDSA keys and verifies
// signatures. It is the infrastructure adapter behind the usecase-layer
// SignatureEngine port (declared in PR4).
type Service struct {
	db  *sql.DB
	kek []byte
	now func() time.Time
}

// NewService constructs the signing service. kek must be a 32-byte AES-256 key.
func NewService(db *sql.DB, kek []byte, now func() time.Time) *Service {
	if db == nil || len(kek) != 32 || now == nil {
		panic("documents/signing: NewService requires db, 32-byte kek and clock")
	}
	return &Service{db: db, kek: kek, now: now}
}

// SignDigest signs a 32-byte digest with the signer's ECDSA private key,
// lazily creating the keypair + self-signed certificate on first use. It
// returns the ASN.1 DER signature and the signer's certificate (PEM).
func (s *Service) SignDigest(ctx context.Context, userID int64, subjectCommonName string, digest []byte) ([]byte, string, error) {
	priv, certPEM, err := s.loadOrCreateKey(ctx, userID, subjectCommonName)
	if err != nil {
		return nil, "", err
	}
	der, err := ecdsa.SignASN1(rand.Reader, priv, digest)
	if err != nil {
		return nil, "", fmt.Errorf("documents/signing: sign: %w", err)
	}
	return der, certPEM, nil
}

// Verify reports whether der is a valid ECDSA signature over digest under the
// public key embedded in certPEM.
func (s *Service) Verify(certPEM string, digest []byte, der []byte) (bool, error) {
	block, _ := pem.Decode([]byte(certPEM))
	if block == nil || block.Type != "CERTIFICATE" {
		return false, fmt.Errorf("documents/signing: verify: not a PEM certificate")
	}
	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return false, fmt.Errorf("documents/signing: verify: parse certificate: %w", err)
	}
	pub, ok := cert.PublicKey.(*ecdsa.PublicKey)
	if !ok {
		return false, fmt.Errorf("documents/signing: verify: certificate public key is not ECDSA")
	}
	return ecdsa.VerifyASN1(pub, digest, der), nil
}

// loadOrCreateKey returns the user's private key + certificate PEM, generating
// and persisting them on first use. On a concurrent first-use race (unique
// violation), it re-reads the row the other writer committed.
func (s *Service) loadOrCreateKey(ctx context.Context, userID int64, subjectCommonName string) (*ecdsa.PrivateKey, string, error) {
	priv, certPEM, found, err := s.readKey(ctx, userID)
	if err != nil {
		return nil, "", err
	}
	if found {
		return priv, certPEM, nil
	}

	priv, certPEM, encPriv, err := s.generateKey(subjectCommonName)
	if err != nil {
		return nil, "", err
	}
	const insert = `INSERT INTO signing_keys (user_id, algorithm, certificate_pem, encrypted_private_key)
		VALUES ($1, $2, $3, $4)`
	_, err = s.db.ExecContext(ctx, insert, userID, string(entities.SignatureAlgorithmECDSAP256SHA256), certPEM, encPriv)
	if err != nil {
		if isUniqueViolation(err) {
			// Another writer created the key first; use theirs.
			priv, certPEM, found, rerr := s.readKey(ctx, userID)
			if rerr != nil {
				return nil, "", rerr
			}
			if found {
				return priv, certPEM, nil
			}
		}
		return nil, "", fmt.Errorf("documents/signing: persist key: %w", err)
	}
	return priv, certPEM, nil
}

// readKey loads and decrypts an existing signing key. found is false when no
// row exists for the user.
func (s *Service) readKey(ctx context.Context, userID int64) (*ecdsa.PrivateKey, string, bool, error) {
	const q = `SELECT certificate_pem, encrypted_private_key FROM signing_keys WHERE user_id = $1`
	var certPEM, encPriv string
	err := s.db.QueryRowContext(ctx, q, userID).Scan(&certPEM, &encPriv)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, "", false, nil
	}
	if err != nil {
		return nil, "", false, fmt.Errorf("documents/signing: read key: %w", err)
	}
	privPEM, err := appcrypto.DecryptString(encPriv, s.kek)
	if err != nil {
		return nil, "", false, fmt.Errorf("documents/signing: decrypt key: %w", err)
	}
	priv, err := parseECPrivateKey(privPEM)
	if err != nil {
		return nil, "", false, err
	}
	return priv, certPEM, true, nil
}

// generateKey mints a fresh ECDSA P-256 keypair + self-signed certificate and
// returns the private key, the certificate PEM, and the AES-encrypted private
// key PEM ready for storage.
func (s *Service) generateKey(subjectCommonName string) (*ecdsa.PrivateKey, string, string, error) {
	priv, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, "", "", fmt.Errorf("documents/signing: generate key: %w", err)
	}

	serial, err := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 128))
	if err != nil {
		return nil, "", "", fmt.Errorf("documents/signing: serial: %w", err)
	}
	now := s.now()
	tmpl := &x509.Certificate{
		SerialNumber:          serial,
		Subject:               pkix.Name{CommonName: subjectCommonName},
		NotBefore:             now,
		NotAfter:              now.Add(certValidity),
		KeyUsage:              x509.KeyUsageDigitalSignature,
		BasicConstraintsValid: true,
	}
	certDER, err := x509.CreateCertificate(rand.Reader, tmpl, tmpl, &priv.PublicKey, priv)
	if err != nil {
		return nil, "", "", fmt.Errorf("documents/signing: create certificate: %w", err)
	}
	certPEM := string(pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certDER}))

	privDER, err := x509.MarshalECPrivateKey(priv)
	if err != nil {
		return nil, "", "", fmt.Errorf("documents/signing: marshal key: %w", err)
	}
	privPEM := string(pem.EncodeToMemory(&pem.Block{Type: "EC PRIVATE KEY", Bytes: privDER}))
	encPriv, err := appcrypto.EncryptString(privPEM, s.kek)
	if err != nil {
		return nil, "", "", fmt.Errorf("documents/signing: encrypt key: %w", err)
	}
	return priv, certPEM, encPriv, nil
}

// parseECPrivateKey decodes a PEM-encoded EC private key.
func parseECPrivateKey(privPEM string) (*ecdsa.PrivateKey, error) {
	block, _ := pem.Decode([]byte(privPEM))
	if block == nil {
		return nil, fmt.Errorf("documents/signing: private key is not PEM")
	}
	priv, err := x509.ParseECPrivateKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("documents/signing: parse private key: %w", err)
	}
	return priv, nil
}

// isUniqueViolation reports whether err is a PostgreSQL unique-constraint error.
func isUniqueViolation(err error) bool {
	var pqErr *pq.Error
	return errors.As(err, &pqErr) && pqErr.Code == "23505"
}
