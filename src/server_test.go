package main

import (
	"better_auth/config"
	"better_auth/pw"
	"fmt"
	"log"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"path"
	"testing"

	"github.com/jbrodriguez/mlog"
)

func TestMain(m *testing.M) {
	mlog.Start(mlog.LevelError, "")
	m.Run()
}

func redir(req *http.Request, via []*http.Request) error {
	return http.ErrUseLastResponse
}

func makeClient() *http.Client {
	jar, err := cookiejar.New(nil)
	if err != nil {
		log.Fatalf("Got error while creating cookie jar %s", err.Error())
	}
	return &http.Client{
		CheckRedirect: redir,
		Jar:           jar,
	}
}

func mockConfig(t *testing.T) *config.Config {
	return &config.Config{
		Address:    "localhost",
		Port:       8675,
		PasswdFile: path.Join(t.TempDir(), "better_auth.pw"),
	}
}

func startServer(cfg *config.Config) (string, error) {
	srv, err := NewServer(cfg)
	if err != nil {
		return "", err
	}
	go srv.StartAndBlock()
	return fmt.Sprintf("http://%s:%d/", cfg.Address, cfg.Port), nil
}

func TestMissingHTML(t *testing.T) {
	cfg := mockConfig(t)
	addr, _ := startServer(cfg)

	err := os.Rename("./static/login.html", "./static/login.html.backup")
	if err != nil {
		t.Fatalf("Failed to temporarily move login.html: %s", err)
	}

	defer os.Rename("./static/login.html.backup", "./static/login.html")

	resp, err := http.Get(addr)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != 404 {
		t.Fatalf("Unexpected status code %d", resp.StatusCode)
	}
}

func TestGetLogin(t *testing.T) {
	cfg := mockConfig(t)
	addr, _ := startServer(cfg)

	httpclient := makeClient()

	resp, err := httpclient.Get(addr + "login")
	if err != nil {
		t.Fatal(err)
	}

	if resp.StatusCode != 200 {
		t.Fatalf("Unexpected status code %d", resp.StatusCode)
	}

	csrf := getCookie(CSRF_TOKEN, resp)
	if csrf == nil {
		t.Fatal("csrf cookie not in response")
	}

	resp, err = httpclient.Get(addr + "login")
	if err != nil {
		t.Fatal(err)
	}

	if resp.StatusCode != 200 {
		t.Fatalf("Unexpected status code %d", resp.StatusCode)
	}

	csrf = getCookie(CSRF_TOKEN, resp)
	if csrf != nil {
		t.Fatal("csrf cookie set twice")
	}
}

func TestPostLogin(t *testing.T) {
	const TESTUSER string = "qwerty"
	const TESTPASS string = "1q2w3e4r5t6y"

	cfg := mockConfig(t)
	pwMan, _ := pw.New(cfg.PasswdFile)
	pwMan.AddUser(TESTUSER, TESTPASS)

	addr, _ := startServer(cfg)
	client := makeClient()

	// post good login with no csrf
	resp, err := client.PostForm(addr+"login", url.Values{
		"username": {TESTUSER},
		"password": {TESTPASS},
	})
	_ = err

	if resp.StatusCode != 511 {
		t.Fatalf("unexpected status code %d for no-csrf post", resp.StatusCode)
	}

	// get csrf
	_, err = client.Get(addr + "login")
	if err != nil {
		t.Fatal(err)
	}

	// post bad username and password
	resp, err = client.PostForm(addr+"login", url.Values{
		"username": {"a bad user"},
		"password": {"a bad pass"},
	})

	if err != nil {
		t.Fatal(err)
	}

	if resp.StatusCode != 401 {
		t.Fatalf("unexpected status code %d for bad user/password post", resp.StatusCode)
	}

	// post bad username
	resp, _ = client.PostForm(addr+"login", url.Values{
		"username": {"a bad user"},
		"password": {TESTPASS},
	})

	if resp.StatusCode != 401 {
		t.Fatalf("unexpected status code %d for bad username post", resp.StatusCode)
	}

	// post bad password
	resp, _ = client.PostForm(addr+"login", url.Values{
		"username": {TESTUSER},
		"password": {"a bad pass"},
	})

	if resp.StatusCode != 401 {
		t.Fatalf("unexpected status code %d for bad password post", resp.StatusCode)
	}

	// post good login
	resp, _ = client.PostForm(addr+"login", url.Values{
		"username": {TESTUSER},
		"password": {TESTPASS},
	})

	if resp.StatusCode != 200 {
		t.Fatalf("unexpected status code %d for good login post", resp.StatusCode)
	}

}

func getCookie(name string, resp *http.Response) *http.Cookie {
	for _, c := range resp.Cookies() {
		if c.Name == name {
			return c
		}
	}
	return nil
}
