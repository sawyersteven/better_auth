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
	r := m.Run()
	os.Exit(r)
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

var port int = 9000

func mockConfig(t *testing.T) *config.Config {
	port += 1
	return &config.Config{
		Address:    "localhost",
		Port:       port,
		PasswdFile: path.Join(t.TempDir(), "better_auth.pw"),
	}
}

func TestMissingHTML(t *testing.T) {
	cfg := mockConfig(t)
	srv, _ := NewServer(cfg)
	addr := fmt.Sprintf("http://%s:%d/", cfg.Address, cfg.Port)
	go srv.StartAndBlock()

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
	srv, _ := NewServer(cfg)
	addr := fmt.Sprintf("http://%s:%d/", cfg.Address, cfg.Port)
	go srv.StartAndBlock()

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

	srv, _ := NewServer(cfg)
	addr := fmt.Sprintf("http://%s:%d/", cfg.Address, cfg.Port)
	go srv.StartAndBlock()

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

func TestReloadPW(t *testing.T) {
	cfg := mockConfig(t)
	srv, _ := NewServer(cfg)
	addr := fmt.Sprintf("http://%s:%d/", cfg.Address, cfg.Port)
	go srv.StartAndBlock()

	client := makeClient()

	const USERNAME string = "Asimov"
	const PASSWORD string = "found@tion"

	// get csrf
	_, err := client.Get(addr + "login")
	if err != nil {
		t.Fatal(err)
	}

	// attempt to log in
	resp, err := client.PostForm(addr+"login", url.Values{
		"username": {USERNAME},
		"password": {PASSWORD},
	})

	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode == 200 {
		t.Fatal("login succeeded with test user. remove Asimov and run test again")
	}

	pw_man, err := pw.New(cfg.PasswdFile)
	if err != nil {
		mlog.Error(err)
		return
	}

	err = pw_man.AddUser(USERNAME, PASSWORD)
	if err != nil {
		t.Fatal(err)
	}

	resp, err = http.Get(addr + "reloadpasswd")
	if err != nil {
		t.Fatal(err)
		return
	}

	if resp.StatusCode != 200 {
		t.Fatalf("unexpected status code %d", resp.StatusCode)
	}

	// attempt to log in
	resp, err = client.PostForm(addr+"login", url.Values{
		"username": {USERNAME},
		"password": {PASSWORD},
	})

	if err != nil {
		t.Fatal(err)
	}

	if resp.StatusCode != 200 {
		t.Fatalf("login failed: %d", resp.StatusCode)
	}
}

func TestAuthRequest(t *testing.T) {
	const TESTUSER string = "Archer"
	const TESTPASS string = "codename_duchess"
	cfg := mockConfig(t)

	pw_man, err := pw.New(cfg.PasswdFile)
	if err != nil {
		mlog.Error(err)
		return
	}

	err = pw_man.AddUser(TESTUSER, TESTPASS)
	if err != nil {
		mlog.Error(err)
		return
	}

	srv, _ := NewServer(cfg)
	addr := fmt.Sprintf("http://%s:%d/", cfg.Address, cfg.Port)
	go srv.StartAndBlock()
	client := makeClient()
	resp, err := client.Get(addr + "authrequest")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != 401 {
		t.Fatalf("unexpected status code %d", resp.StatusCode)
	}

	// log in...
	_, err = client.Get(addr + "login")
	if err != nil {
		t.Fatal(err)
	}

	resp, err = client.PostForm(addr+"login", url.Values{
		"username": {TESTUSER},
		"password": {TESTPASS},
	})
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != 200 {
		t.Fatalf("unexpected status code %d", resp.StatusCode)
	}

	// At this point the client refuses to accept the SetCookie header from the
	// login request, so the session_id never gets sent to the auth request and
	// it will always return a 401. Just test this manually in a web browser to
	// make sure it works if you've changed anything

}

func getCookie(name string, resp *http.Response) *http.Cookie {
	for _, c := range resp.Cookies() {
		if c.Name == name {
			return c
		}
	}
	return nil
}
