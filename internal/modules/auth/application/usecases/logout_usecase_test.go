package usecases

import (
	"context"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// mockRevokedTokenRepo is a hand-rolled mock for RevokedTokenRepository.
// Defined here (not in shared mocks) to keep this RED test isolated.
type mockRevokedTokenRepo struct {
	mock.Mock
}

func (m *mockRevokedTokenRepo) Revoke(ctx context.Context, jti string, ttl time.Duration) error {
	args := m.Called(ctx, jti, ttl)
	return args.Error(0)
}

func (m *mockRevokedTokenRepo) IsRevoked(ctx context.Context, jti string) (bool, error) {
	args := m.Called(ctx, jti)
	return args.Bool(0), args.Error(1)
}

// signTokenWithJTI builds a minimal valid JWT signed with secret. Used to
// give the logout use case something it can parse and look at exp / jti.
func signTokenWithJTI(t *testing.T, secret []byte, jti string, expiresAt time.Time) string {
	t.Helper()
	tok := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"jti":     jti,
		"user_id": float64(7),
		"role":    "teacher",
		"exp":     expiresAt.Unix(),
		"iat":     time.Now().Add(-time.Minute).Unix(),
	})
	signed, err := tok.SignedString(secret)
	assert.NoError(t, err)
	return signed
}

// TestLogoutUseCase_RevokesToken_HappyPath verifies that Logout extracts the
// JTI from the access token and stores it in the revoked-token repository
// with TTL bounded by the token's remaining lifetime — so the entry expires
// right when the token would have anyway, no Redis garbage left over.
func TestLogoutUseCase_RevokesToken_HappyPath(t *testing.T) {
	secret := []byte("test-secret")
	expiresAt := time.Now().Add(10 * time.Minute)
	token := signTokenWithJTI(t, secret, "jti-abc-123", expiresAt)

	repo := new(mockRevokedTokenRepo)
	// Expect: Revoke called with exact jti and a TTL within ±2s of 10min.
	repo.On("Revoke", mock.Anything, "jti-abc-123",
		mock.MatchedBy(func(d time.Duration) bool {
			diff := d - 10*time.Minute
			if diff < 0 {
				diff = -diff
			}
			return diff <= 2*time.Second
		}),
	).Return(nil)

	uc := NewLogoutUseCase(repo, secret)
	err := uc.Logout(context.Background(), token)

	assert.NoError(t, err)
	repo.AssertExpectations(t)
}

// TestLogoutUseCase_RejectsInvalidToken — malformed or wrong-secret tokens
// must not be silently accepted; logout must surface an error.
func TestLogoutUseCase_RejectsInvalidToken(t *testing.T) {
	secret := []byte("test-secret")
	repo := new(mockRevokedTokenRepo)

	uc := NewLogoutUseCase(repo, secret)
	err := uc.Logout(context.Background(), "not.a.real.token")

	assert.Error(t, err)
	repo.AssertNotCalled(t, "Revoke", mock.Anything, mock.Anything, mock.Anything)
}

// TestLogoutUseCase_RejectsExpiredToken — already-expired tokens have no
// remaining lifetime; logging out is a no-op error rather than writing
// a zero-TTL entry.
func TestLogoutUseCase_RejectsExpiredToken(t *testing.T) {
	secret := []byte("test-secret")
	token := signTokenWithJTI(t, secret, "jti-old", time.Now().Add(-time.Hour))

	repo := new(mockRevokedTokenRepo)

	uc := NewLogoutUseCase(repo, secret)
	err := uc.Logout(context.Background(), token)

	assert.Error(t, err)
	repo.AssertNotCalled(t, "Revoke", mock.Anything, mock.Anything, mock.Anything)
}

// TestLogoutUseCase_RejectsTokenWithoutJTI — JTI is the only stable handle
// for revocation. Tokens without it cannot be safely revoked; reject loudly.
func TestLogoutUseCase_RejectsTokenWithoutJTI(t *testing.T) {
	secret := []byte("test-secret")
	tok := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": float64(7),
		"exp":     time.Now().Add(10 * time.Minute).Unix(),
	})
	signed, _ := tok.SignedString(secret)

	repo := new(mockRevokedTokenRepo)

	uc := NewLogoutUseCase(repo, secret)
	err := uc.Logout(context.Background(), signed)

	assert.Error(t, err)
	repo.AssertNotCalled(t, "Revoke", mock.Anything, mock.Anything, mock.Anything)
}
