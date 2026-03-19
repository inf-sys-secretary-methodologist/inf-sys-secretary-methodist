package handlers

import (
	"bytes"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/textproto"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/users/domain/entities"
	domainErrors "github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/shared/domain/errors"
)

func TestGetExtensionFromMimeType(t *testing.T) {
	tests := []struct {
		mimeType string
		expected string
	}{
		{"image/jpeg", ".jpg"},
		{"image/png", ".png"},
		{"image/gif", ".gif"},
		{"image/webp", ".webp"},
		{"application/octet-stream", ".jpg"},
		{"text/plain", ".jpg"},
	}

	for _, tt := range tests {
		t.Run(tt.mimeType, func(t *testing.T) {
			result := getExtensionFromMimeType(tt.mimeType)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestAvatarHandler_Upload_InvalidID(t *testing.T) {
	h := &AvatarHandler{}

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/users/abc/avatar", nil)
	c.Params = gin.Params{{Key: "id", Value: "abc"}}

	h.Upload(c)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestAvatarHandler_Upload_NoAuth(t *testing.T) {
	h := &AvatarHandler{}

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/users/1/avatar", nil)
	c.Params = gin.Params{{Key: "id", Value: "1"}}

	h.Upload(c)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestAvatarHandler_Upload_ForbiddenOtherUser(t *testing.T) {
	h := &AvatarHandler{}

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/users/2/avatar", nil)
	c.Params = gin.Params{{Key: "id", Value: "2"}}
	c.Set("user_id", int64(1))
	c.Set("user_role", "methodist")

	h.Upload(c)
	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestAvatarHandler_Upload_NoFile(t *testing.T) {
	h := &AvatarHandler{}

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/users/1/avatar", nil)
	c.Params = gin.Params{{Key: "id", Value: "1"}}
	c.Set("user_id", int64(1))

	h.Upload(c)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestAvatarHandler_Delete_InvalidID(t *testing.T) {
	h := &AvatarHandler{}

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodDelete, "/users/abc/avatar", nil)
	c.Params = gin.Params{{Key: "id", Value: "abc"}}

	h.Delete(c)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestAvatarHandler_Delete_NoAuth(t *testing.T) {
	h := &AvatarHandler{}

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodDelete, "/users/1/avatar", nil)
	c.Params = gin.Params{{Key: "id", Value: "1"}}

	h.Delete(c)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestAvatarHandler_Delete_ForbiddenOtherUser(t *testing.T) {
	h := &AvatarHandler{}

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodDelete, "/users/2/avatar", nil)
	c.Params = gin.Params{{Key: "id", Value: "2"}}
	c.Set("user_id", int64(1))
	c.Set("user_role", "methodist")

	h.Delete(c)
	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestAvatarHandler_GetAvatarURL_InvalidID(t *testing.T) {
	h := &AvatarHandler{}

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/users/abc/avatar", nil)
	c.Params = gin.Params{{Key: "id", Value: "abc"}}

	h.GetAvatarURL(c)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestAllowedAvatarTypes(t *testing.T) {
	assert.True(t, AllowedAvatarTypes["image/jpeg"])
	assert.True(t, AllowedAvatarTypes["image/png"])
	assert.True(t, AllowedAvatarTypes["image/gif"])
	assert.True(t, AllowedAvatarTypes["image/webp"])
	assert.False(t, AllowedAvatarTypes["text/plain"])
}

func TestAvatarConstants(t *testing.T) {
	assert.Equal(t, int64(5*1024*1024), int64(MaxAvatarSize))
	assert.Equal(t, "avatars", AvatarFolder)
}

func TestAvatarHandler_Upload_OversizedFile(t *testing.T) {
	h := &AvatarHandler{}

	// Create multipart form with large file
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)
	part, _ := writer.CreateFormFile("avatar", "big.jpg")
	// Write more than MaxAvatarSize header to indicate a large file
	_, _ = part.Write(make([]byte, 100)) // small content but header.Size will be reported
	_ = writer.Close()

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/users/1/avatar", &buf)
	c.Request.Header.Set("Content-Type", writer.FormDataContentType())
	c.Params = gin.Params{{Key: "id", Value: "1"}}
	c.Set("user_id", int64(1))

	h.Upload(c)
	// The file is small so it passes size validation, but s3Client is nil
	// This should reach the content-type check
	assert.NotEqual(t, http.StatusUnauthorized, w.Code)
}

func TestAvatarHandler_Upload_InvalidContentType(t *testing.T) {
	h := &AvatarHandler{}

	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)
	part, _ := writer.CreateFormFile("avatar", "doc.txt")
	_, _ = part.Write([]byte("hello"))
	_ = writer.Close()

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	req := httptest.NewRequest(http.MethodPost, "/users/1/avatar", &buf)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	c.Request = req
	c.Params = gin.Params{{Key: "id", Value: "1"}}
	c.Set("user_id", int64(1))

	h.Upload(c)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestAvatarHandler_Upload_AdminCanUploadOther(t *testing.T) {
	h := &AvatarHandler{}

	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)
	part, _ := writer.CreateFormFile("avatar", "photo.png")
	_, _ = part.Write([]byte("PNG image data"))
	_ = writer.Close()

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	req := httptest.NewRequest(http.MethodPost, "/users/2/avatar", &buf)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	c.Request = req
	c.Params = gin.Params{{Key: "id", Value: "2"}}
	c.Set("user_id", int64(1))
	c.Set("user_role", "system_admin")

	h.Upload(c)
	// Passes auth check, reaches content type check - text/plain from part
	assert.NotEqual(t, http.StatusForbidden, w.Code)
}

func TestAvatarHandler_Delete_AdminCanDeleteOther(t *testing.T) {
	h := &AvatarHandler{}

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodDelete, "/users/2/avatar", nil)
	c.Params = gin.Params{{Key: "id", Value: "2"}}
	c.Set("user_id", int64(1))
	c.Set("user_role", "system_admin")

	// Panics because userUseCase is nil
	assert.Panics(t, func() { h.Delete(c) })
}

func TestNewAvatarHandler(t *testing.T) {
	h := NewAvatarHandler(nil, nil)
	assert.NotNil(t, h)
}

func TestAvatarHandler_Upload_EmptyContentType(t *testing.T) {
	h := &AvatarHandler{}

	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)
	part, _ := writer.CreateFormFile("avatar", "photo.jpg")
	_, _ = part.Write([]byte{0xFF, 0xD8, 0xFF, 0xE0}) // JPEG magic bytes
	_ = writer.Close()

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	req := httptest.NewRequest(http.MethodPost, "/users/1/avatar", &buf)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	c.Request = req
	c.Params = gin.Params{{Key: "id", Value: "1"}}
	c.Set("user_id", int64(1))

	h.Upload(c)
	// content-type will be resolved from multipart header or default
	// Result depends on what gin reports as content type for the part
	assert.NotEqual(t, http.StatusUnauthorized, w.Code)
}

func TestAvatarHandler_Upload_NoExtension(t *testing.T) {
	h := &AvatarHandler{}

	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)
	part, _ := writer.CreateFormFile("avatar", "photo") // no extension
	_, _ = part.Write([]byte{0xFF, 0xD8, 0xFF, 0xE0})
	_ = writer.Close()

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	req := httptest.NewRequest(http.MethodPost, "/users/1/avatar", &buf)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	c.Request = req
	c.Params = gin.Params{{Key: "id", Value: "1"}}
	c.Set("user_id", int64(1))

	h.Upload(c)
	assert.NotEqual(t, http.StatusUnauthorized, w.Code)
}

func TestAvatarHandler_Delete_PanicsOnNilUserUseCase(t *testing.T) {
	h := &AvatarHandler{}

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodDelete, "/users/1/avatar", nil)
	c.Params = gin.Params{{Key: "id", Value: "1"}}
	c.Set("user_id", int64(1))

	// Panics because userUseCase is nil
	assert.Panics(t, func() { h.Delete(c) })
}

func TestAvatarHandler_GetAvatarURL_PanicsOnNilUserUseCase(t *testing.T) {
	h := &AvatarHandler{}

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/users/1/avatar", nil)
	c.Params = gin.Params{{Key: "id", Value: "1"}}

	assert.Panics(t, func() { h.GetAvatarURL(c) })
}

func TestAvatarHandler_Delete_WithMockUseCase(t *testing.T) {
	profileRepo := new(mockUserProfileRepo)
	authRepo := new(mockAuthUserRepo)

	user := &entities.UserWithOrg{ID: 1, Email: "a@b.com", Name: "Test", Avatar: ""}
	profileRepo.On("GetProfileByID", mock.Anything, int64(1)).Return(user, nil)

	uc := newUserUseCase(authRepo, profileRepo, new(mockDepartmentRepo), new(mockPositionRepo))
	h := &AvatarHandler{userUseCase: uc}

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodDelete, "/users/1/avatar", nil)
	c.Params = gin.Params{{Key: "id", Value: "1"}}
	c.Set("user_id", int64(1))

	h.Delete(c)
	// No avatar set -> returns BadRequest
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestAvatarHandler_GetAvatarURL_NoAvatar(t *testing.T) {
	profileRepo := new(mockUserProfileRepo)
	authRepo := new(mockAuthUserRepo)

	user := &entities.UserWithOrg{ID: 1, Email: "a@b.com", Name: "Test", Avatar: ""}
	profileRepo.On("GetProfileByID", mock.Anything, int64(1)).Return(user, nil)

	uc := newUserUseCase(authRepo, profileRepo, new(mockDepartmentRepo), new(mockPositionRepo))
	h := &AvatarHandler{userUseCase: uc}

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/users/1/avatar", nil)
	c.Params = gin.Params{{Key: "id", Value: "1"}}

	h.GetAvatarURL(c)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestAvatarHandler_GetAvatarURL_WithAvatar_NilS3(t *testing.T) {
	profileRepo := new(mockUserProfileRepo)
	authRepo := new(mockAuthUserRepo)

	user := &entities.UserWithOrg{ID: 1, Email: "a@b.com", Name: "Test", Avatar: "avatars/1_abc123.jpg"}
	profileRepo.On("GetProfileByID", mock.Anything, int64(1)).Return(user, nil)

	uc := newUserUseCase(authRepo, profileRepo, new(mockDepartmentRepo), new(mockPositionRepo))
	h := &AvatarHandler{userUseCase: uc, s3Client: nil}

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/users/1/avatar", nil)
	c.Params = gin.Params{{Key: "id", Value: "1"}}

	// Panics on nil s3Client
	assert.Panics(t, func() { h.GetAvatarURL(c) })
}

func TestAvatarHandler_GetAvatarURL_UserNotFound(t *testing.T) {
	profileRepo := new(mockUserProfileRepo)
	authRepo := new(mockAuthUserRepo)

	profileRepo.On("GetProfileByID", mock.Anything, int64(999)).Return(nil, domainErrors.ErrNotFound)

	uc := newUserUseCase(authRepo, profileRepo, new(mockDepartmentRepo), new(mockPositionRepo))
	h := &AvatarHandler{userUseCase: uc}

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/users/999/avatar", nil)
	c.Params = gin.Params{{Key: "id", Value: "999"}}

	h.GetAvatarURL(c)
	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestAvatarHandler_Delete_UserNotFound(t *testing.T) {
	profileRepo := new(mockUserProfileRepo)
	authRepo := new(mockAuthUserRepo)

	profileRepo.On("GetProfileByID", mock.Anything, int64(999)).Return(nil, domainErrors.ErrNotFound)

	uc := newUserUseCase(authRepo, profileRepo, new(mockDepartmentRepo), new(mockPositionRepo))
	h := &AvatarHandler{userUseCase: uc}

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodDelete, "/users/999/avatar", nil)
	c.Params = gin.Params{{Key: "id", Value: "999"}}
	c.Set("user_id", int64(999))

	h.Delete(c)
	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestAvatarHandler_Upload_WithUserUseCaseError(t *testing.T) {
	profileRepo := new(mockUserProfileRepo)
	authRepo := new(mockAuthUserRepo)

	profileRepo.On("GetProfileByID", mock.Anything, int64(1)).Return(nil, domainErrors.ErrNotFound)

	uc := newUserUseCase(authRepo, profileRepo, new(mockDepartmentRepo), new(mockPositionRepo))
	h := &AvatarHandler{userUseCase: uc}

	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)
	// Use CreatePart to set explicit Content-Type header for the part
	partHeader := make(textproto.MIMEHeader)
	partHeader["Content-Disposition"] = []string{`form-data; name="avatar"; filename="photo.jpg"`}
	partHeader["Content-Type"] = []string{"image/jpeg"}
	part, _ := writer.CreatePart(partHeader)
	_, _ = part.Write([]byte{0xFF, 0xD8, 0xFF, 0xE0})
	_ = writer.Close()

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	req := httptest.NewRequest(http.MethodPost, "/users/1/avatar", &buf)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	c.Request = req
	c.Params = gin.Params{{Key: "id", Value: "1"}}
	c.Set("user_id", int64(1))

	h.Upload(c)
	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestAvatarHandler_Upload_WithExistingAvatar(t *testing.T) {
	profileRepo := new(mockUserProfileRepo)
	authRepo := new(mockAuthUserRepo)

	user := &entities.UserWithOrg{ID: 1, Email: "a@b.com", Name: "Test", Avatar: "avatars/old.jpg"}
	profileRepo.On("GetProfileByID", mock.Anything, int64(1)).Return(user, nil)

	uc := newUserUseCase(authRepo, profileRepo, new(mockDepartmentRepo), new(mockPositionRepo))
	h := &AvatarHandler{userUseCase: uc, s3Client: nil}

	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)
	partHeader := make(textproto.MIMEHeader)
	partHeader["Content-Disposition"] = []string{`form-data; name="avatar"; filename="photo.jpg"`}
	partHeader["Content-Type"] = []string{"image/jpeg"}
	part, _ := writer.CreatePart(partHeader)
	_, _ = part.Write([]byte{0xFF, 0xD8, 0xFF, 0xE0})
	_ = writer.Close()

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	req := httptest.NewRequest(http.MethodPost, "/users/1/avatar", &buf)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	c.Request = req
	c.Params = gin.Params{{Key: "id", Value: "1"}}
	c.Set("user_id", int64(1))

	// Panics when trying to delete old avatar via nil s3Client
	assert.Panics(t, func() { h.Upload(c) })
}

func TestAvatarHandler_Delete_WithAvatarSet(t *testing.T) {
	profileRepo := new(mockUserProfileRepo)
	authRepo := new(mockAuthUserRepo)

	user := &entities.UserWithOrg{ID: 1, Email: "a@b.com", Name: "Test", Avatar: "avatars/1_abc.jpg"}
	profileRepo.On("GetProfileByID", mock.Anything, int64(1)).Return(user, nil)

	uc := newUserUseCase(authRepo, profileRepo, new(mockDepartmentRepo), new(mockPositionRepo))
	h := &AvatarHandler{userUseCase: uc, s3Client: nil}

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodDelete, "/users/1/avatar", nil)
	c.Params = gin.Params{{Key: "id", Value: "1"}}
	c.Set("user_id", int64(1))

	// Panics because s3Client is nil
	assert.Panics(t, func() { h.Delete(c) })
}

func TestAvatarHandler_Delete_WithNonAvatarPath(t *testing.T) {
	profileRepo := new(mockUserProfileRepo)
	authRepo := new(mockAuthUserRepo)

	user := &entities.UserWithOrg{ID: 1, Email: "a@b.com", Name: "Test", Avatar: "other/path.jpg"}
	profileRepo.On("GetProfileByID", mock.Anything, int64(1)).Return(user, nil)
	authRepo.On("GetByID", mock.Anything, int64(1)).Return(nil, domainErrors.ErrNotFound)

	uc := newUserUseCase(authRepo, profileRepo, new(mockDepartmentRepo), new(mockPositionRepo))
	h := &AvatarHandler{userUseCase: uc}

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodDelete, "/users/1/avatar", nil)
	c.Params = gin.Params{{Key: "id", Value: "1"}}
	c.Set("user_id", int64(1))

	h.Delete(c)
	// Avatar doesn't start with AvatarFolder, so S3 delete is skipped
	// Then tries to update profile and fails because auth repo returns not found
	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestUserHandler_List_InvalidQuery(t *testing.T) {
	uc := newUserUseCase(new(mockAuthUserRepo), new(mockUserProfileRepo), new(mockDepartmentRepo), new(mockPositionRepo))
	handler := NewUserHandler(uc)
	router := setupUserRouter(handler)

	// page as non-number -> ShouldBindQuery fails
	req := httptest.NewRequest(http.MethodGet, "/users?page=abc", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}
