package token_store

import (
	"testing"
	"time"
)

func TestNew(t *testing.T) {
	s := New("Test", 1)
	if s.tokens == nil {
		t.Fatal("Tokens map may not be nil")
	}
	if s.name != "Test" {
		t.Fatalf("Incorrect token store name `%s`", s.name)
	}
	if s.lifetime != 1*time.Second {
		t.Fatalf("Incorrect lifetime `%s`", s.lifetime)
	}
}

func TestUniqueID(t *testing.T) {
	s := New("Test", 1)

	ids := make(map[string]struct{})

	for i := 0; i < 4096; i++ {
		token, err := s.NewToken()
		if err != nil {
			t.Fatal(err)
		}
		_, exists := ids[token.id]
		if exists {
			t.Fatalf("Duplicate token id generated `%s`", token.id)
		}
		ids[token.id] = struct{}{}
	}
}

func TestStartNewSession(t *testing.T) {
	s := New("Test", 1)

	token, err := s.NewToken()
	if err != nil {
		t.Fatal(err)
	}

	if token.name != "Test" {
		t.Fatalf("Incorrect token name `%s`", token.name)
	}

	if token.id == "" {
		t.Fatal("Empty token id")
	}

	if token.expires.Before(time.Now()) {
		t.Fatal("Bad token expiration")
	}
}

func TestIsValid(t *testing.T) {
	s := New("Test", 1)

	token, _ := s.NewToken()

	if !s.IsValid(token.id) {
		t.Fatal("Invalid token ID")
	}

	time.Sleep(time.Second * 2)
	if s.IsValid(token.id) {
		t.Fatal("Expired token ID is valid")
	}
}

func TestRefresh(t *testing.T) {
	s := New("Test", 1)

	token, _ := s.NewToken()
	time.Sleep(time.Millisecond * 500)
	ok := s.IsValid(token.id)
	if !ok {
		t.Fatal("Token expired too quickly")
	}

	err := s.RefreshExp(token)
	if err != nil {
		t.Fatal(err)
	}

	time.Sleep(time.Millisecond * 750)
	ok = s.IsValid(token.id)
	if !ok {
		t.Fatal("Token expired too quickly")
	}

	time.Sleep(time.Millisecond * 750)
	ok = s.IsValid(token.id)
	if !ok {
		t.Fatal("Token didn't expire")
	}

}
