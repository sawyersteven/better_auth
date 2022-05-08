package config

import (
	"os"
	"path"
	"reflect"
	"testing"

	"github.com/jbrodriguez/mlog"
)

func TestMain(m *testing.M) {
	mlog.Start(mlog.LevelError, "")
	m.Run()
}

/// Tests creating a default config, writing to disk, reading from disk, and
///  comparing read values to default values
func TestNew(t *testing.T) {
	dir := t.TempDir()
	f := path.Join(dir, "better_auth.conf")

	os.Args = []string{os.Args[0], "--config", f}
	conf, err := Build()
	if err != nil {
		t.Fatal(err)
	}

	def := Default()
	def.ConfigFile = f
	vd := reflect.ValueOf(*def)

	vc := reflect.ValueOf(*conf)
	tc := vc.Type()
	for i := 0; i < tc.NumField(); i++ {
		n := tc.Field(i).Name
		if vc.FieldByName(n).Interface() != vd.FieldByName(n).Interface() {
			t.Logf("Config mismatch at field %s", n)
			t.FailNow()
		}
	}
}

/// Tests that a NewDefault config has all fields assigned
func TestNewDefault(t *testing.T) {

	c := Default()

	vc := reflect.ValueOf(*c)
	tc := vc.Type()

	for i := 0; i < tc.NumField(); i++ {
		n := tc.Field(i).Name
		if n[0] >= 97 {
			continue
		}

		fi := vc.Field(i).Interface()

		switch ft := fi.(type) {
		case int:
			if ft == 0 {
				t.Fatalf("Field %s should not be 0", vc.Field(i).Type().Name())
			}
		case string:
			if ft == "" {
				t.Fatalf("Field %s should not be empty", vc.Field(i).Type().Name())
			}
		case *adduserCmd:
			if ft != nil {
				t.Fatal("Subcommand fields should be nil")
			}
		default:
			t.Logf("Unexpected type %s. This may be an error or the test may need to be updated", vc.Field(i).Type())
			t.Fail()
		}
	}
}
