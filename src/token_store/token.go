package token_store

import (
	"net/http"
	"time"
)

type Token struct {
	name    string
	id      string
	expires *time.Time
}

func (t *Token) ID() string {
	return t.id
}

func (t *Token) Expires() time.Time {
	return *t.expires
}

func (t *Token) ToCookie() *http.Cookie {
	return &http.Cookie{
		Name:     t.name,
		Expires:  *t.expires,
		Value:    t.id,
		SameSite: http.SameSiteStrictMode,
		HttpOnly: true,
		Path:     "/",
	}
}
