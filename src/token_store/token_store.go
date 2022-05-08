package token_store

// NOTE done

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"sync"
	"time"
)

const TOKEN_LEN int = 42

type TokenStore struct {
	name     string
	tokens   map[string]time.Time
	lifetime time.Duration
	lock     sync.Mutex
}

func New(name string, lifetime int) *TokenStore {

	return &TokenStore{
		name:     name,
		tokens:   make(map[string]time.Time),
		lifetime: time.Second * time.Duration(lifetime),
		lock:     sync.Mutex{},
	}
}

/// Creates a new token with a random id
/// Returns a Token that contains the id and expiration timestamp
func (s *TokenStore) NewToken() (*Token, error) {
	s.cleanExpired()
	id, err := s.randomID()
	if err != nil {
		return nil, err
	}
	exp := s.makeEpiryTimestamp()

	s.lock.Lock()
	defer s.lock.Unlock()
	s.tokens[id] = exp
	return &Token{name: s.name, id: id, expires: &exp}, nil
}

func (s *TokenStore) cleanExpired() {
	s.lock.Lock()
	defer s.lock.Unlock()
	now := time.Now()
	for _, k := range Keys(s.tokens) {
		if s.tokens[k].Before(now) {
			delete(s.tokens, k)
		}
	}
}

func Keys[M ~map[K]V, K comparable, V any](m M) []K {
	i := 0
	r := make([]K, len(m))
	for k := range m {
		r[i] = k
	}
	return r
}

func (s *TokenStore) makeEpiryTimestamp() time.Time {
	return time.Now().Add(s.lifetime)
}

func (s *TokenStore) randomID() (string, error) {
	rngContainer := make([]byte, TOKEN_LEN)
	for {
		_, err := io.ReadFull(rand.Reader, rngContainer)
		if err != nil {
			return "", err
		}
		id := hex.EncodeToString(rngContainer)

		_, exists := s.tokens[id]
		if !exists {
			return id, nil
		}
	}
}

/// Checks if token id exists and is not expired.
/// Returns bool indicating if id is a valid token and was able to be updated
func (s *TokenStore) IsValid(id string) bool {
	exp, contains := s.tokens[id]
	if !contains || exp.Before(time.Now()) {
		delete(s.tokens, id)
		return false
	}
	s.tokens[id] = s.makeEpiryTimestamp()
	return contains
}

/// Extends token exipration from now using lifetime.
/// Returns error if token does not exist or has already expired
func (s *TokenStore) RefreshExp(token *Token) error {
	s.cleanExpired()

	exp, contains := s.tokens[token.id]
	if !contains {
		return fmt.Errorf("invalid token")
	}

	s.lock.Lock()
	defer s.lock.Unlock()

	if exp.Before(time.Now()) {
		delete(s.tokens, token.id)
		return fmt.Errorf("invalid token")
	}

	exp = s.makeEpiryTimestamp()
	s.tokens[token.id] = exp
	token.expires = &exp

	return nil
}
