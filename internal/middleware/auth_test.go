package middleware

import (
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

func TestMain(m *testing.M) {
	gin.SetMode(gin.TestMode)
	os.Setenv("JWT_SECRET", "test-secret")
	os.Exit(m.Run())
}

func TestGenerateJWT_RoundTrip(t *testing.T) {
	userID := uuid.New()
	email := "test@example.com"

	tokenStr, err := GenerateJWT(userID, email)
	if err != nil {
		t.Fatalf("GenerateJWT returned error: %v", err)
	}
	if tokenStr == "" {
		t.Fatal("GenerateJWT returned an empty token")
	}

	token, err := jwt.ParseWithClaims(tokenStr, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte("test-secret"), nil
	})
	if err != nil || !token.Valid {
		t.Fatalf("failed to parse generated token: %v", err)
	}

	claims := token.Claims.(*JWTClaims)
	if claims.UserID != userID {
		t.Errorf("expected UserID %s, got %s", userID, claims.UserID)
	}
	if claims.Email != email {
		t.Errorf("expected Email %s, got %s", email, claims.Email)
	}
	if claims.ExpiresAt == nil || !claims.ExpiresAt.Time.After(time.Now()) {
		t.Error("expected ExpiresAt to be in the future")
	}
}

func newTestRouter() *gin.Engine {
	r := gin.New()
	r.GET("/protected", AuthRequired(), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"userID": c.GetString("userID"),
			"email":  c.GetString("email"),
		})
	})
	return r
}

func TestAuthRequired_MissingHeader(t *testing.T) {
	r := newTestRouter()

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}
}

func TestAuthRequired_MalformedHeader(t *testing.T) {
	r := newTestRouter()

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "NotBearer sometoken")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}
}

func TestAuthRequired_InvalidToken(t *testing.T) {
	r := newTestRouter()

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer garbage.token.here")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}
}

func TestAuthRequired_WrongSecret(t *testing.T) {
	claims := &JWTClaims{
		UserID: uuid.New(),
		Email:  "test@example.com",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenStr, _ := token.SignedString([]byte("wrong-secret"))

	r := newTestRouter()
	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+tokenStr)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401 for token signed with wrong secret, got %d", w.Code)
	}
}

func TestAuthRequired_ExpiredToken(t *testing.T) {
	claims := &JWTClaims{
		UserID: uuid.New(),
		Email:  "expired@example.com",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(-1 * time.Hour)),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenStr, _ := token.SignedString([]byte("test-secret"))

	r := newTestRouter()
	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+tokenStr)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401 for expired token, got %d", w.Code)
	}
}

func TestAuthRequired_ValidToken(t *testing.T) {
	userID := uuid.New()
	tokenStr, err := GenerateJWT(userID, "valid@example.com")
	if err != nil {
		t.Fatalf("GenerateJWT failed: %v", err)
	}

	r := newTestRouter()
	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+tokenStr)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d, body: %s", w.Code, w.Body.String())
	}
	if !strings.Contains(w.Body.String(), userID.String()) {
		t.Errorf("expected response to contain userID %s, got %s", userID, w.Body.String())
	}
}
