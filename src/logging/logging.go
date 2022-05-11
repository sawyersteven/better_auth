package logging

import (
	"better_auth/config"
	"fmt"
	"os"

	"github.com/jbrodriguez/mlog"
)

func init() {
	mlog.Start(mlog.LevelInfo, "")
}

func Start(conf *config.Config) error {
	err := os.MkdirAll(conf.LogDir, 0644)
	if err != nil {
		return err
	}

	mlog.StartEx(mlog.LevelInfo, "better_auth.log", conf.LogSize*1024*1024, conf.LogBackups)
	fmt.Printf("Logging started at %s \n", conf.LogDir)
	return nil
}
