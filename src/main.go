package main

import (
	"better_auth/config"
	"better_auth/logging"
	"better_auth/pw"
	"fmt"
	"os"

	"github.com/jbrodriguez/mlog"
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
	mlog.Info(fmt.Sprintf("Adding new user '%s'", conf.AddUser.Username))
	pw_man, err := pw.New(conf.PasswdFile)
	if err != nil {
		mlog.Error(err)
	}

	err = pw_man.AddUser(conf.AddUser.Username, conf.AddUser.Password)
	if err != nil {
		mlog.Error(err)
	}

	mlog.Info("User %s added to %s \n", conf.AddUser.Username, conf.PasswdFile)

}
