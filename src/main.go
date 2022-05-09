package main

import (
	"better_auth/config"
	"better_auth/logging"
	"better_auth/pw"
	"fmt"
	"net/http"
	"os"
	"syscall"

	"github.com/jbrodriguez/mlog"
	"golang.org/x/term"
)

func main() {
	var err error

	conf, err := config.Build()
	if err != nil {
		mlog.Error(err)
		os.Exit(1)
	}

	err = logging.Start(conf)
	if err != nil {
		mlog.Error(err)
		os.Exit(1)
	}

	switch {
	case conf.AddUser != nil:
		subCommandAddUser(conf)
	default:
		s, err := NewServer(conf)
		if err != nil {
			mlog.Error(err)
			os.Exit(1)
		}
		s.StartAndBlock()
	}
}

func subCommandAddUser(conf *config.Config) {
	fmt.Printf("Adding new user `%s`\n", conf.AddUser.Username)
	pw_man, err := pw.New(conf.PasswdFile)
	if err != nil {
		mlog.Error(err)
		return
	}

	if conf.AddUser.Password == "" {
		for {
			fmt.Printf("Enter Password for %s:", conf.AddUser.Username)
			bytepw, err := term.ReadPassword(int(syscall.Stdin))
			if err != nil {
				mlog.Error(err)
				return
			}
			err = pw.ValidatePassword(string(bytepw))
			if err == nil {
				conf.AddUser.Password = string(bytepw)
				break
			}
			fmt.Println(err)
		}
	}
	fmt.Println()

	err = pw_man.AddUser(conf.AddUser.Username, conf.AddUser.Password)
	if err != nil {
		mlog.Error(err)
		return
	}

	mlog.Info("User %s added to %s \n", conf.AddUser.Username, conf.PasswdFile)

	addr := fmt.Sprintf("http://%s:%d/reloadpasswd", conf.Address, conf.Port)
	fmt.Println("Attempting to update better_auth server...")
	resp, err := http.Get(addr)
	if err != nil {
		fmt.Println("Could not reach better_auth server")
		return
	}
	if resp.StatusCode != 200 {
		fmt.Printf("better_auth server responded with status %d\n", resp.StatusCode)
	}

	fmt.Printf("better_auth server updated with new user `%s`\n", conf.AddUser.Username)
}
