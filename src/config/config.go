// NOTE done
package config

import (
	"better_auth/files"
	"encoding/json"
	"fmt"
	"os"

	"github.com/alexflint/go-arg"
	"github.com/jbrodriguez/mlog"
)

/// Config represents operating config for entire application
/// Combines default options, file options, and cli arguments
type Config struct {
	AddUser        *adduserCmd `arg:"subcommand:adduser" json:"-"`
	Address        string      `arg:"-a,--address" help:"server address"`
	Port           int         `arg:"-p,--port" help:"server port"`
	SessionTimeout int         `arg:"-"`
	PasswdFile     string      `arg:"--pw" help:"path to better_auth.pw file"`
	LogDir         string      `arg:"--logdir" help:"path to log directory"`
	LogSize        int         `arg:"-"`
	LogBackups     int         `arg:"-"`
	ConfigFile     string      `arg:"--config" help:"path to better_auth.conf file" json:"-"`
}

func Default() *Config {
	return &Config{
		Address:        "localhost",
		Port:           8675,
		SessionTimeout: 3600,
		PasswdFile:     DefaultPaths.Passwd,
		LogDir:         DefaultPaths.Log,
		LogSize:        1,
		LogBackups:     5,

		ConfigFile: DefaultPaths.Config,
	}
}

type adduserCmd struct {
	Username string `arg:"positional,required" help:"New user name"`
	Password string `arg:"positional" help:"New user's password"`
}

func Build() (*Config, error) {
	var err error

	confPath := getConfigPath()
	conf := Default()
	err = parseJsonOver(conf, confPath)
	if err != nil {
		return nil, err
	}
	parseArgsOver(conf)
	return conf, nil
}

func getConfigPath() string {
	def := Default()
	parseArgsOver(def)
	return def.ConfigFile
}

/// Applies arguments from cli over default config
func parseArgsOver(conf *Config) {
	arg.MustParse(conf)
}

func parseJsonOver(conf *Config, filePath string) error {
	if !files.FileExists(filePath) {
		writeNewDefault(filePath)
	}

	mlog.Info("Loading config file `%s`\n", filePath)
	file, err := os.OpenFile(filePath, os.O_RDONLY, 0644)
	if err != nil {
		return fmt.Errorf("unable to read from config file `%s`: %s", filePath, err)
	}
	defer file.Close()

	err = json.NewDecoder(file).Decode(conf)
	if err != nil {
		return err
	}

	return nil
}

func writeNewDefault(filePath string) error {
	mlog.Info("Config file not found at %s, writing a new default config\n", filePath)

	jsonDump, err := json.MarshalIndent(Default(), "", "\t")
	if err != nil {
		return err
	}

	f, err := files.MkDirsAndOpen(filePath, os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return fmt.Errorf("Config file `%s` could not be created: %s", filePath, err)

	}
	_, err = f.Write(jsonDump)
	if err != nil {
		return fmt.Errorf("Config file `%s` could not be written to: %s", filePath, err)
	}
	f.Close()

	return nil
}
