//go:build linux

package config

var DefaultPaths struct {
	Config  string
	Passwd 		string
	Log         string
}

func inti(){
	DefaultPaths.Config = "/etc/better_auth/better_auth.conf"
	DefaultPaths.Passwd = "/etc/better_auth/better_auth.pw"
	DefaultPaths.Log = "/var/log/better_auth/"
}