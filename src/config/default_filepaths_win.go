//go:build windows

package config

import (
	"os/user"
	"path"
)

var DefaultPaths struct {
	Config string
	Passwd string
	Log    string
}

func init() {
	u, _ := user.Current()
	dir := path.Join(u.HomeDir, "better_auth")
	DefaultPaths.Config = path.Join(dir, "better_auth.conf")
	DefaultPaths.Passwd = path.Join(dir, "better_auth.users")
	DefaultPaths.Log = path.Join(dir, "logs", "better_auth.log")
}
