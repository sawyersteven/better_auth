/*
PW files are stored on disk with each user/password on its own line like:

clint_eastwood:some_bcrypted_pass
john_wayne:another_bcrypted_pass
*/

package pw

import (
	"better_auth/files"
	"bufio"
	"errors"
	"fmt"
	"os"
	"strings"
	"sync"

	"github.com/jbrodriguez/mlog"
	"golang.org/x/crypto/bcrypt"
)

type PWManager struct {
	users map[string][]byte // username: hashed pw
	file  string
	lock  sync.Mutex
}

/// Creates new PWManager from data in filePath. If filePath does not exist a
/// new empty better_auth.pw will be created.
func New(filePath string) (*PWManager, error) {
	pwMan := &PWManager{users: make(map[string][]byte), file: filePath, lock: sync.Mutex{}}

	if !files.FileExists(filePath) {
		mlog.Info("Creating new password file `%s`", filePath)
		f, err := os.Create(filePath)
		if err != nil {
			return nil, fmt.Errorf("unable to create password file `%s`: %s", filePath, err)
		}
		f.Close()
	}
	err := pwMan.parseAuthFile(filePath)
	return pwMan, err
}

func (a *PWManager) parseAuthFile(filePath string) error {
	mlog.Info("Reading password file from %s", filePath)

	file, err := os.OpenFile(filePath, os.O_RDONLY, 0000)
	if err != nil {
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	scanner.Split(bufio.ScanLines)

	a.lock.Lock()
	defer a.lock.Unlock()
	line := 0
	for scanner.Scan() {
		line++
		parts := strings.Split(scanner.Text(), ":")
		if len(parts) != 2 {
			err = fmt.Errorf("invalid entry on line %d of %s", line, filePath)
			return err
		}
		a.users[parts[0]] = []byte(parts[1])
	}
	return nil
}

func (a *PWManager) Reload() error {
	a.lock.Lock()
	for k := range a.users {
		delete(a.users, k)
	}
	a.lock.Unlock()
	return a.parseAuthFile(a.file)
}

/// Adds user to file and in-memory cache
func (a *PWManager) AddUser(username string, password string) error {
	mlog.Info("Adding user `%s` to password file `%s`", username, a.file)
	err := a.validateUsername(username)
	if err != nil {
		return err
	}

	err = ValidatePassword(password)
	if err != nil {
		return err
	}

	hashedPassword, err := hashPassword(password)
	if err != nil {
		return err
	}

	err = a.writeUserToFile(username, hashedPassword)
	if err != nil {
		return err
	}
	a.lock.Lock()
	defer a.lock.Unlock()
	a.users[username] = hashedPassword
	return nil
}

/// Adds user and password to pw file on disk
func (a *PWManager) writeUserToFile(username string, hashedPassword []byte) error {

	f, err := files.MkDirsAndOpen(a.file, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = f.Write([]byte(username + ":" + string(hashedPassword) + "\n"))
	if err != nil {
		return err
	}

	return nil
}

/// Verifies that the username exists and the password matches the loaded password file
func (a *PWManager) Verify(username string, password string) bool {
	hashedPass, userExists := a.users[username]
	if !userExists {
		return false
	}
	return bcrypt.CompareHashAndPassword(hashedPass, []byte(password)) == nil
}

func hashPassword(password string) ([]byte, error) {
	return bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
}

func (a *PWManager) validateUsername(username string) error {
	_, found := a.users[username]
	if found {
		return fmt.Errorf("username `%s` already exists in password file `%s`", username, a.file)
	}

	if len(username) == 0 {
		return errors.New("username may not be empty")
	}

	if len([]byte(username)) > 72 {
		return errors.New("username too long (max 72 bytes)")
	}

	if strings.Contains(username, ":") {
		return errors.New("illegal character `:` in username")
	}
	return nil
}

func ValidatePassword(password string) error {
	if len(password) < 8 {
		return errors.New("password too short (min 8 characters)")
	}
	return nil
}
