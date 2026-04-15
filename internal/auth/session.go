package auth

import (
	"context"
	"database/sql"
	"encoding/base64"
	"errors"
	"net/http"
	"time"

	"cups-web/internal/store"

	"github.com/gorilla/securecookie"
)

var s *securecookie.SecureCookie

const sessionCookieName = "session"
const csrfCookieName = "csrf_token"

const (
	settingHashKey  = "session_hash_key"
	settingBlockKey = "session_block_key"
)

func SetupSecureCookie(db *sql.DB) error {
	ctx := context.Background()

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	hashKeyStr, err := store.GetSettingString(ctx, tx, settingHashKey, "")
	if err != nil {
		return err
	}
	blockKeyStr, err := store.GetSettingString(ctx, tx, settingBlockKey, "")
	if err != nil {
		return err
	}

	var hashKey, blockKey []byte

	if hashKeyStr == "" {
		hashKey = securecookie.GenerateRandomKey(32)
		hashKeyStr = base64.StdEncoding.EncodeToString(hashKey)
		if err := store.SetSettingString(ctx, tx, settingHashKey, hashKeyStr); err != nil {
			return err
		}
	} else {
		hashKey, _ = base64.StdEncoding.DecodeString(hashKeyStr)
		if len(hashKey) == 0 {
			hashKey = []byte(hashKeyStr)
		}
	}

	if blockKeyStr == "" {
		blockKey = securecookie.GenerateRandomKey(32)
		blockKeyStr = base64.StdEncoding.EncodeToString(blockKey)
		if err := store.SetSettingString(ctx, tx, settingBlockKey, blockKeyStr); err != nil {
			return err
		}
	} else {
		blockKey, _ = base64.StdEncoding.DecodeString(blockKeyStr)
		if len(blockKey) == 0 {
			blockKey = []byte(blockKeyStr)
		}
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	s = securecookie.New(hashKey, blockKey)
	return nil
}

type Session struct {
	UserID   int64     `json:"userId"`
	Username string    `json:"username"`
	Role     string    `json:"role"`
	Expires  time.Time `json:"expires"`
}

func SetSession(w http.ResponseWriter, sess Session) error {
	if s == nil {
		return errors.New("securecookie not initialized")
	}
	encoded, err := s.Encode(sessionCookieName, sess)
	if err != nil {
		return err
	}
	cookie := &http.Cookie{
		Name:     sessionCookieName,
		Value:    encoded,
		Path:     "/",
		HttpOnly: true,
		Secure:   false,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   86400,
	}
	http.SetCookie(w, cookie)
	return nil
}

func ClearSession(w http.ResponseWriter) {
	cookie := &http.Cookie{
		Name:     sessionCookieName,
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		Secure:   false,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   -1,
	}
	http.SetCookie(w, cookie)
	// clear csrf cookie too
	csrf := &http.Cookie{
		Name:   csrfCookieName,
		Value:  "",
		Path:   "/",
		MaxAge: -1,
	}
	http.SetCookie(w, csrf)
}

func GetSession(r *http.Request) (Session, error) {
	var sess Session
	if s == nil {
		return sess, errors.New("securecookie not initialized")
	}
	c, err := r.Cookie(sessionCookieName)
	if err != nil {
		return sess, err
	}
	err = s.Decode(sessionCookieName, c.Value, &sess)
	if err != nil {
		return sess, err
	}
	return sess, nil
}
