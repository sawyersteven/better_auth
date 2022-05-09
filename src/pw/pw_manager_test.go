package pw

import (
	"os"
	"path"
	"testing"

	"github.com/jbrodriguez/mlog"
)

func TestMain(m *testing.M) {
	mlog.Start(mlog.LevelError, "")
	m.Run()
}

/// Tests creating new manager
func TestNewPWManager(t *testing.T) {
	dir := t.TempDir()
	f := path.Join(dir, "better_auth.pw")
	os.WriteFile(f, []byte(""), 0644)

	_, err := New(f)
	if err != nil {
		t.Fatal(err)
	}
}

func TestCreateFile(t *testing.T) {
	dir := t.TempDir()
	f := path.Join(dir, "better_auth.pw")
	_, err := New(f)
	if err != nil {
		t.Fatal(err)
	}
}

/// Tests adding user to auth file and memory. Writes and reads temporary auth file
func TestAddUser(t *testing.T) {
	dir := t.TempDir()
	f := path.Join(dir, "better_auth.pw")
	os.WriteFile(f, []byte(""), 0644)

	user := "JohnWayne"
	pass := "19IwoJima49"

	c, err := New(f)
	if err != nil {
		t.Fatal(err)
	}

	err = c.AddUser(user, pass)
	if err != nil {
		t.Fatal(err)
	}

	match := c.Verify(user, pass)
	if !match {
		t.Fatal("User and Password do not match in pw map")
	}

	match = c.Verify("JohnWayne ", pass)
	if match {
		t.Fatal("Invalid user passed pw verification")
	}

	c, err = New(f)
	if err != nil {
		t.Fatal(err)
	}
	match = c.Verify(user, pass)
	if !match {
		t.Fatal("User and Password do not match pw file")
	}

	match = c.Verify("JohnWayne ", pass)
	if match {
		t.Fatal("Invalid user passed pw verification")
	}
}

/// Tests bad user names
func TestBadUsernames(t *testing.T) {
	dir := t.TempDir()
	f := path.Join(dir, "better_auth.pw")
	os.WriteFile(f, []byte(""), 0644)

	c, err := New(f)
	if err != nil {
		t.Fatal(err)
	}

	err = c.AddUser("JohnWayne", "19IwoJima49")
	if err != nil {
		t.Fatal(err)
	}

	for _, name := range []string{"an_invalid:user_name", "", "JohnWayne", "1234567890_1234567890_1234567890_1234567890_1234567890_1234567890_123456789"} {
		err = c.AddUser(name, "a_valid_password")
		if err == nil {
			t.Fatalf("Bad username `%s` passed validation", name)
		}
	}
}

/// Tests bad passwords
func TestBadPasswords(t *testing.T) {
	dir := t.TempDir()
	f := path.Join(dir, "better_auth.pw")
	os.WriteFile(f, []byte(""), 0644)

	c, err := New(f)
	if err != nil {
		t.Fatal(err)
	}

	for _, password := range []string{"short"} {
		err = c.AddUser("a_valid_username", password)
		if err == nil {
			t.Fatalf("Bad password `%s` passed validation", password)
		}
	}
}

func TestBadPWFile(t *testing.T) {
	dir := t.TempDir()
	f := path.Join(dir, "better_auth.pw")
	os.WriteFile(f, []byte("not_a:valid:entry"), 0644)

	_, err := New(f)
	if err == nil {
		t.Fatal("invalid entry [1] passed pw parser")
	}

	os.WriteFile(f, []byte("not_a_valid_entry"), 0644)

	_, err = New(f)
	if err == nil {
		t.Fatal("invalid entry [2] passed pw parser")
	}
}
