package main

import (
	"api/config"
	"api/internal/log"
	"context"
	"flag"
	"os"

	_ "api/migrations"

	_ "github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
	"go.uber.org/zap"
)

var (
	flags = flag.NewFlagSet("goose", flag.ExitOnError)
	dir   = flags.String("dir", "./migrations", "directory with migration files")
)

func main() {
	cfg := config.MustNewConfig()
	log := log.MustNewLogger(cfg)

	flags.Parse(os.Args[1:])
	args := flags.Args()

	if len(args) < 3 {
		flags.Usage()
		return
	}

	_, command := args[1], args[2]

	goose.SetVerbose(true)
	goose.SetLogger(&gooseLogger{log: log.Logger})

	db, err := goose.OpenDBWithDriver("postgres", cfg.PgDBURL)
	if err != nil {
		log.Fatal("goose: failed to open DB", zap.Error(err))
	}

	defer func() {
		if err := db.Close(); err != nil {
			log.Fatal("goose: failed to close DB", zap.Error(err))
		}
	}()

	arguments := []string{}
	if len(args) > 3 {
		arguments = append(arguments, args[3:]...)
	}

	log.Debug("goose", zap.String("command", command), zap.String("dir", *dir), zap.Any("args", arguments))
	ctx := context.Background()
	if err := goose.RunContext(ctx, command, db, *dir, arguments...); err != nil {
		log.Fatal("goose ", zap.String("command", command), zap.Error(err))
	}
}

type gooseLogger struct {
	log *zap.Logger
}

func (gl *gooseLogger) Printf(format string, v ...interface{}) {
	gl.log.Info(format, zap.Any("args", v))
}

func (gl *gooseLogger) Fatalf(format string, v ...interface{}) {
	gl.log.Error(format, zap.Any("args", v))
}
