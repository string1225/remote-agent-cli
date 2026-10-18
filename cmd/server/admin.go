package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"strings"

	"github.com/string1225/remote-agent-cli/internal/store"
	"golang.org/x/crypto/bcrypt"
)

func runAdmin(args []string, input io.Reader, output io.Writer) error {
	if len(args) < 1 || (args[0] != "create" && args[0] != "grant") {
		return errors.New("usage: remote-server admin create --username NAME --password-stdin | admin grant --username NAME (stop the server first)")
	}
	flags := flag.NewFlagSet("admin "+args[0], flag.ContinueOnError)
	flags.SetOutput(io.Discard)
	username := flags.String("username", "", "account username")
	passwordStdin := flags.Bool("password-stdin", false, "read password from stdin")
	if err := flags.Parse(args[1:]); err != nil {
		return errors.New("invalid admin flags; use --username NAME and, for create, --password-stdin")
	}
	if flags.NArg() != 0 {
		return errors.New("unexpected admin arguments")
	}
	name, _, err := store.NormalizeUsername(*username)
	if err != nil {
		return errors.New("username must contain 1–64 characters and no control characters")
	}
	var hash []byte
	if args[0] == "create" {
		if !*passwordStdin {
			return errors.New("admin create requires --password-stdin")
		}
		raw, err := io.ReadAll(io.LimitReader(input, 4097))
		if err != nil {
			return errors.New("cannot read password from stdin")
		}
		password := strings.TrimSuffix(strings.TrimSuffix(string(raw), "\n"), "\r")
		if len(password) < 6 || len(password) > 72 {
			return errors.New("password must be 6–72 bytes")
		}
		hash, err = bcrypt.GenerateFromPassword([]byte(password), 12)
		if err != nil {
			return errors.New("cannot hash password")
		}
	} else if *passwordStdin {
		return errors.New("admin grant does not accept --password-stdin")
	}
	db, err := store.Open(env("DATABASE_PATH", "data/remote-agent.db"))
	if err != nil {
		return fmt.Errorf("cannot open account database; stop remote-server before using admin: %w", err)
	}
	defer db.Close()
	if args[0] == "create" {
		u := store.User{ID: store.Token(), Username: name, Password: hash, Admin: true}
		if err := db.CreateUser(u); err != nil {
			return fmt.Errorf("cannot create administrator: %w", err)
		}
		_, err = fmt.Fprintf(output, "Created administrator %s (%s).\n", name, u.ID)
		return err
	}
	u, err := db.User(name)
	if err != nil {
		return errors.New("account not found")
	}
	if u.Disabled {
		return errors.New("cannot grant administrator access to a disabled account")
	}
	yes := true
	if _, err := db.UpdateUser("", u.ID, &yes, nil, nil); err != nil {
		return fmt.Errorf("cannot grant administrator access: %w", err)
	}
	_, err = fmt.Fprintf(output, "Granted administrator access to %s (%s).\n", u.Username, u.ID)
	return err
}
