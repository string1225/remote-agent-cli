package server

import (
	"errors"
	"net/http"

	"github.com/string1225/remote-agent-cli/internal/store"
	"golang.org/x/crypto/bcrypt"
)

func (s *Server) requireAdmin(w http.ResponseWriter, r *http.Request) (store.User, bool) {
	session, ok := s.require(w, r)
	if !ok {
		return store.User{}, false
	}
	u, err := s.Store.UserByID(session.UserID)
	if err != nil || u.Disabled || !u.Admin {
		fail(w, 403, "administrator access required")
		return store.User{}, false
	}
	return u, true
}

func passwordHash(w http.ResponseWriter, password string) ([]byte, bool) {
	if len(password) < 6 || len(password) > 72 {
		fail(w, 400, "password must be 6–72 bytes")
		return nil, false
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(password), 12)
	if err != nil {
		fail(w, 500, "cannot hash password")
		return nil, false
	}
	return hash, true
}

func accountError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, store.ErrExists):
		fail(w, 409, "username already exists")
	case errors.Is(err, store.ErrLastAdmin):
		fail(w, 409, store.ErrLastAdmin.Error())
	case errors.Is(err, store.ErrInvalid):
		fail(w, 400, "username must contain 1–64 characters and no control characters")
	case errors.Is(err, store.ErrDenied):
		fail(w, 403, "account update denied")
	default:
		fail(w, 500, "cannot update account")
	}
}

func (s *Server) changeAccount(w http.ResponseWriter, r *http.Request) {
	session, ok := s.require(w, r)
	if !ok {
		return
	}
	var in struct {
		Username        *string `json:"username"`
		CurrentPassword string  `json:"currentPassword"`
		Password        *string `json:"password"`
	}
	if !decode(w, r, &in) {
		return
	}
	if in.Username == nil && in.Password == nil {
		fail(w, 400, "username or password required")
		return
	}
	u, err := s.Store.UserByID(session.UserID)
	if err != nil || in.CurrentPassword == "" || bcrypt.CompareHashAndPassword(u.Password, []byte(in.CurrentPassword)) != nil {
		fail(w, 403, "current password is incorrect")
		return
	}
	name := ""
	if in.Username != nil {
		var e error
		name, _, e = store.NormalizeUsername(*in.Username)
		if e != nil {
			accountError(w, e)
			return
		}
	}
	var hash []byte
	if in.Password != nil {
		var valid bool
		hash, valid = passwordHash(w, *in.Password)
		if !valid {
			return
		}
	}
	if err := s.Store.ChangeAccount(u.ID, name, hash, u.Password); err != nil {
		accountError(w, err)
		return
	}
	s.hub.closeUserBrowsers(u.ID)
	s.cookie(w, "", -1)
	reply(w, 200, map[string]bool{"ok": true})
}

type accountView struct {
	ID          string `json:"id"`
	Username    string `json:"username"`
	Admin       bool   `json:"admin"`
	Disabled    bool   `json:"disabled"`
	CreatedAt   int64  `json:"createdAt,omitempty"`
	DeviceCount int    `json:"deviceCount"`
}

func (s *Server) accountView(u store.User) accountView {
	agents, _ := s.Store.Agents(u.ID)
	return accountView{ID: u.ID, Username: u.Username, Admin: u.Admin, Disabled: u.Disabled, CreatedAt: u.CreatedAt, DeviceCount: len(agents)}
}

func (s *Server) adminUsers(w http.ResponseWriter, r *http.Request) {
	if _, ok := s.requireAdmin(w, r); !ok {
		return
	}
	users, err := s.Store.Users(r.URL.Query().Get("search"))
	if err != nil {
		fail(w, 500, "cannot list accounts")
		return
	}
	list := []accountView{}
	for _, u := range users {
		list = append(list, s.accountView(u))
	}
	reply(w, 200, list)
}

func (s *Server) adminCreateUser(w http.ResponseWriter, r *http.Request) {
	actor, ok := s.requireAdmin(w, r)
	if !ok {
		return
	}
	var in struct {
		Username string `json:"username"`
		Password string `json:"password"`
		Admin    bool   `json:"admin"`
	}
	if !decode(w, r, &in) {
		return
	}
	name, _, err := store.NormalizeUsername(in.Username)
	if err != nil {
		accountError(w, err)
		return
	}
	hash, ok := passwordHash(w, in.Password)
	if !ok {
		return
	}
	u := store.User{ID: store.Token(), Username: name, Password: hash, Admin: in.Admin}
	if err := s.Store.CreateUserAsAdmin(actor.ID, u); err != nil {
		accountError(w, err)
		return
	}
	u, err = s.Store.UserByID(u.ID)
	if err != nil {
		fail(w, 500, "cannot read created account")
		return
	}
	reply(w, 201, s.accountView(u))
}

func (s *Server) adminUpdateUser(w http.ResponseWriter, r *http.Request) {
	actor, ok := s.requireAdmin(w, r)
	if !ok {
		return
	}
	var in struct {
		Admin    *bool   `json:"admin"`
		Disabled *bool   `json:"disabled"`
		Password *string `json:"password"`
	}
	if !decode(w, r, &in) {
		return
	}
	if in.Admin == nil && in.Disabled == nil && in.Password == nil {
		fail(w, 400, "account update required")
		return
	}
	if _, err := s.Store.UserByID(r.PathValue("id")); err != nil {
		fail(w, 404, "account not found")
		return
	}
	var hash []byte
	if in.Password != nil {
		var valid bool
		hash, valid = passwordHash(w, *in.Password)
		if !valid {
			return
		}
	}
	u, err := s.Store.UpdateUser(actor.ID, r.PathValue("id"), in.Admin, in.Disabled, hash)
	if err != nil {
		accountError(w, err)
		return
	}
	if u.Disabled {
		s.cancelInference(u.ID)
	}
	s.hub.closeUserBrowsers(u.ID)
	if u.Disabled {
		agents, err := s.Store.Agents(u.ID)
		if err != nil {
			fail(w, 500, "cannot disconnect devices")
			return
		}
		for _, a := range agents {
			s.hub.closeAgent(a.ID)
		}
	}
	reply(w, 200, s.accountView(u))
}
