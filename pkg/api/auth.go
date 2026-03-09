package api

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"os"
	"strings"
	"time"
)

type signInRequest struct {
	Password string `json:"password"`
}

type tokenPayload struct {
	PasswordHash string `json:"pwd"`
}

func passwordHash(pwd string) string {
	sum := sha256.Sum256([]byte(pwd))
	return base64.StdEncoding.EncodeToString(sum[:])
}

func encodeJWT(pwd string) (string, error) {
	header := map[string]string{
		"alg": "HS256",
		"typ": "JWT",
	}
	headerJSON, err := json.Marshal(header)
	if err != nil {
		return "", err
	}

	payload := tokenPayload{
		PasswordHash: passwordHash(pwd),
	}
	payloadJSON, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}

	enc := base64.RawURLEncoding
	h := enc.EncodeToString(headerJSON)
	p := enc.EncodeToString(payloadJSON)
	unsigned := h + "." + p

	key := []byte(pwd)
	mac := hmac.New(sha256.New, key)
	mac.Write([]byte(unsigned))
	signature := enc.EncodeToString(mac.Sum(nil))

	return unsigned + "." + signature, nil
}

func verifyJWT(token, pwd string) bool {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return false
	}

	enc := base64.RawURLEncoding
	unsigned := parts[0] + "." + parts[1]

	key := []byte(pwd)
	mac := hmac.New(sha256.New, key)
	mac.Write([]byte(unsigned))
	expectedSig := enc.EncodeToString(mac.Sum(nil))
	if !hmac.Equal([]byte(expectedSig), []byte(parts[2])) {
		return false
	}

	payloadBytes, err := enc.DecodeString(parts[1])
	if err != nil {
		return false
	}

	var payload tokenPayload
	if err := json.Unmarshal(payloadBytes, &payload); err != nil {
		return false
	}

	return payload.PasswordHash == passwordHash(pwd)
}

// signInHandler обрабатывает POST /api/signin.
func signInHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	envPwd := os.Getenv("TODO_PASSWORD")
	if strings.TrimSpace(envPwd) == "" {
		writeJSONError(w, "Authentication disabled")
		return
	}

	var req signInRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, err.Error())
		return
	}

	if req.Password != envPwd {
		w.WriteHeader(http.StatusUnauthorized)
		writeJSONError(w, "Invalid password")
		return
	}

	token, err := encodeJWT(envPwd)
	if err != nil {
		writeJSONError(w, err.Error())
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "token",
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		Expires:  time.Now().Add(24 * time.Hour),
	})

	writeJSON(w, map[string]string{
		"token": token,
	})
}

// authMiddleware защищает API, если задан TODO_PASSWORD.
func authMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		envPwd := os.Getenv("TODO_PASSWORD")
		if strings.TrimSpace(envPwd) == "" {
			next(w, r)
			return
		}

		cookie, err := r.Cookie("token")
		if err != nil || strings.TrimSpace(cookie.Value) == "" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		if !verifyJWT(cookie.Value, envPwd) {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		next(w, r)
	}
}

