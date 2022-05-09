//go:build linux

package config

var DefaultPaths struct {
	UserConfig  string
	Auth 		string
	Log         string
}{
	Config:"/etc/better_auth/better_auth.conf",
	Passwd:"/etc/better_auth/better_auth.pw",
	Log:"/var/log/better_auth/"
}