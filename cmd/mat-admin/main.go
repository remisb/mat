package main

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"github.com/pkg/errors"
	"github.com/remisb/mat/cmd/mat-admin/internal/conf"
	"github.com/remisb/mat/internal/auth"
	"github.com/remisb/mat/internal/db"
	"github.com/remisb/mat/internal/log"
	"github.com/remisb/mat/internal/schema"
	"github.com/remisb/mat/internal/user"
	"os"
	"time"
)

var (
	// Populated during a build
	version = ""
	commit  = ""
	date    = ""
)

func main() {
	cfg := conf.NewConfig()

	if err := run(cfg); err != nil {
		message := fmt.Sprintf("error: %s", err.Error())
		if log.Sugar == nil {
			fmt.Printf(message)
			os.Exit(1)
		}
		log.Sugar.Error(message)
		os.Exit(1)
	}
}

func run(cfg *conf.Config) error {
	var err error

	cmd := cfg.Args.Num(0)
	switch cmd {
	case "migrate":
		err = migrate(cfg.Db)
	case "seed":
		err = seed(cfg.Db)
	case "useradd":
		err = userAdd(cfg.Db, cfg.Args.Num(1), cfg.Args.Num(2))
	case "keygen":
		err = keygen(cfg.Args.Num(1))
	default:
		err = errors.New("Must specify a command")
	}

	return err
}

func migrate(cfg db.Config) error {
	db, err := db.Open(cfg)
	if err != nil {
		return err
	}
	defer db.Close()

	if err := schema.Migrate(db); err != nil {
		return err
	}

	fmt.Println("Migrations complete")
	return nil
}

func seed(cfg db.Config) error {
	dbc, err := db.Open(cfg)
	if err != nil {
		return err
	}
	defer dbc.Close()

	if err := schema.Seed(dbc); err != nil {
		return err
	}

	fmt.Println("Seed data complete")
	return nil
}

func userAdd(cfg db.Config, email, password string) error {
	dbc, err := db.Open(cfg)
	if err != nil {
		return err
	}
	defer dbc.Close()

	if email == "" || password == "" {
		return errors.New("useradd command must be called with two additional arguments for email and password")
	}

	fmt.Printf("Admin user will be created with email %q and password %q\n", email, password)
	confirm := askForConfirmation("Continue?")
	if !confirm {
		fmt.Println("Canceling")
		return nil
	}

	ctx := context.Background()

	u, err := user.Create(ctx, dbc, "", email, password, []string{auth.RoleUser, auth.RoleAdmin}, time.Now())
	if err != nil {
		return err
	}

	fmt.Println("User created with id:", u.ID)
	return nil
}

// keygen creates an x509 private key for signing auth tokens.
func keygen(path string) error {
	if path == "" {
		return errors.New("keygen missing argument for key path")
	}

	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return errors.Wrap(err, "creating private file")
	}
	file, err := os.Create(path)
	if err != nil {
		return errors.Wrap(err, "creating private key file")
	}
	defer file.Close()

	block := pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(key),
	}

	if err := pem.Encode(file, &block); err != nil {
		return errors.Wrap(err, "encoding to private key file")
	}

	if err := file.Close(); err != nil {
		return errors.Wrap(err, "closing private key file")
	}

	return nil
}
