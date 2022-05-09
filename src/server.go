package main

import (
	"better_auth/config"
	"better_auth/pw"
	"better_auth/token_store"
	"fmt"
	"net/http"
	"os"

	"github.com/jbrodriguez/mlog"
)

const CSRF_TOKEN string = "csrf_token"
const SESSION_TOKEN string = "better_auth_session_token"

type Server struct {
	//addr         string
	pwManager    *pw.PWManager
	csrfStore    *token_store.TokenStore
	sessionStore *token_store.TokenStore
	addr         string
}

func NewServer(cfg *config.Config) (*Server, error) {
	pwm, err := pw.New(cfg.PasswdFile)
	if err != nil {
		return nil, err
	}
	return &Server{
		pwManager:    pwm,
		csrfStore:    token_store.New(CSRF_TOKEN, 15*60),
		sessionStore: token_store.New(SESSION_TOKEN, cfg.SessionTimeout),
		addr:         fmt.Sprintf("%s:%d", cfg.Address, cfg.Port),
	}, nil
}

func (s *Server) StartAndBlock() {
	m := http.NewServeMux()
	m.HandleFunc("/reloadpasswd", s.reloadPasswd)
	m.HandleFunc("/authrequest", s.authrequest)
	m.HandleFunc("/login", s.login)
	mlog.Info("Serving at %s\n", s.addr)
	err := http.ListenAndServe(s.addr, m)

	if err != nil && err.Error() != "http: Server closed" {
		mlog.Error(err)
		os.Exit(1)
	}
}

/// GET returns login page html
/// POST verfies user/password.
///  If name/password aren't valid returns 511. This isn't 100% proper, but makes
///    more sense than putting an error in the body and parsing it at the client
///  If successful starts new session and assigns a cookie to the client.
///  If an error occurred generating the ID a 500 is returned
func (s *Server) login(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		csrfCookie, err := r.Cookie(CSRF_TOKEN)
		if err != nil {
			token, err := s.csrfStore.NewToken()
			if err != nil {
				mlog.Error(err)
				w.WriteHeader(500)
				return
			}
			http.SetCookie(w, token.ToCookie())
		}

		if err != nil {

		} else {
			s.csrfStore.IsValid(csrfCookie.Value)
		}
		http.ServeFile(w, r, "./static/login.html")
		return
	case http.MethodPost:
		csrfCookie, err := r.Cookie(CSRF_TOKEN)
		if err != nil || !s.csrfStore.IsValid(csrfCookie.Value) {
			w.WriteHeader(511)
			return
		}

		usr := r.FormValue("username")
		pwd := r.FormValue("password")
		mlog.Info("Login attempt for user %s from %s", usr, r.RemoteAddr)

		if s.pwManager.Verify(usr, pwd) {
			token, err := s.sessionStore.NewToken()
			if err != nil {
				mlog.Error(err)
				w.WriteHeader(500)
				return
			}

			mlog.Info("Login attempt successful for user %s from %s", usr, r.RemoteAddr)
			http.SetCookie(w, token.ToCookie())
			w.WriteHeader(200)
			return
		}
		mlog.Info("Login attempt failed for user %s from %s", usr, r.RemoteAddr)
		w.WriteHeader(401)
		return
	}
}

/// Handles auth subrequest from nginx
func (s *Server) authrequest(w http.ResponseWriter, r *http.Request) {
	id, _ := r.Cookie(SESSION_TOKEN)
	if id != nil && s.sessionStore.IsValid(id.Value) {
		return
	}
	w.WriteHeader(401)
}

func (s *Server) reloadPasswd(w http.ResponseWriter, r *http.Request) {
	if s.pwManager.Reload() != nil {
		w.WriteHeader(500)
	}
}
