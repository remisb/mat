package main

import (
	"context"
	"fmt"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	"github.com/remisb/mat/cmd/rest-api/internal/conf"
	"github.com/remisb/mat/cmd/rest-api/internal/restaurantapi"
	"github.com/remisb/mat/cmd/rest-api/internal/userapi"
	"github.com/remisb/mat/internal/db"
	"github.com/remisb/mat/internal/log"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"net/http"
	_ "net/http/pprof" // Register the pprof handlers
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"
)

func main() {
	config := conf.NewConfig()
	if err := initLogger(config); err != nil {
		fmt.Printf("error on initLogger err: %s", err)
		os.Exit(1)
	}

	if err := startAPIServerAndWait(*config); err != nil {
		log.Sugar.Errorf("error on starting api server, error :", err)
		os.Exit(1)
	}
}

func initLogger(config *conf.Config) error {
	var logPath string
	var err error

	logCfg := zap.NewDevelopmentConfig()

	if !filepath.IsAbs(config.Server.Log) {
		logPath, err = filepath.Abs(config.Server.Log)
		if err != nil {
			return errors.Wrap(err, "unable to get absolute logfile path")
		}
	}

	fmt.Printf("Log file path: %s", logPath)

	logCfg.OutputPaths = append(logCfg.OutputPaths, logPath)
	log.Logger, err = logCfg.Build()

	if err != nil {
		err := fmt.Errorf("error in log init err: %v", err)
		fmt.Print(err)
	}
	log.Sugar = log.Logger.Sugar()
	logCfg.Level.SetLevel(zapcore.DebugLevel)

	log.Sugar.Infof("Logging is set to console and file: %s", logPath)
	log.Sugar.Sync()

	return nil
}

func startDatabase(dbConf db.Config) (*sqlx.DB, error) {
	log.Sugar.Infof("main : Started : Initializing database support")

	dbx, err := db.Open(dbConf)
	if err != nil {
		return nil, errors.Wrap(err, "connecting to db")
	}

	return dbx, nil
}

func startAPIServerAndWait(config conf.Config) error {
	dbx, err := startDatabase(config.Db)
	if err != nil {
		return err
	}

	defer func() {
		log.Sugar.Infof("main : Database Stopping : %s", config.Db.Host)
		dbx.Close()
	}()

	startDebugService(config)

	// Make a channel to listen for an interrupt or terminate signal from the OS.
	// Use a buffered channel because the signal package requires it.
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)

	// Make a channel to listen for errors coming from the listener. Use a
	// buffered channel so the goroutine can exit if we don't collect this error.
	serverErrors := make(chan error, 1)

	apiServer := startAPIServer(config, dbx, shutdown, serverErrors)
	if err := waitShutdown(config.Server, apiServer, serverErrors, shutdown); err != nil {
		return err
	}
	return nil
}

func startDebugService(config conf.Config) {
	// /debug/pprof - Added to the default mux by importing the net/http/pprof package.
	// /debug/vars - Added to the default mux by importing the expvar package.

	log.Sugar.Infof("main : Started : Initializing debug support")

	go func() {
		log.Sugar.Infof("main : Degub Listening %s", config.Server.DebugHost)
		log.Sugar.Infof("main : Debug Listener closed : %v", http.ListenAndServe(config.Server.DebugHost, http.DefaultServeMux))
	}()
}

func startAPIServer(cfg conf.Config, dbx *sqlx.DB,
	shutdownChan chan os.Signal,
	serverErrors chan error) *http.Server {

	r := chi.NewRouter()
	// A good base middleware stack
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	// Set a timeout value on the request context (ctx), that will signal
	// through ctx.Done() that the request has timed out and further
	// processing should be stopped.
	r.Use(middleware.Timeout(60 * time.Second))

	userServer := userapi.NewServer("development", shutdownChan, dbx)
	restaurantServer := restaurantapi.NewServer("development", shutdownChan, dbx)
	r.Route("/api/v1/", func(r chi.Router) {
		r.Mount("/users", userServer.Router)
		r.Mount("/restaurant", restaurantServer.Router)
	})

	api := http.Server{
		Addr:         cfg.Server.Addr(),
		Handler:      r,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
	}

	// Start the service listening for requests.
	go func() {
		log.Sugar.Infof("main : API listening on %s", api.Addr)
		serverErrors <- api.ListenAndServe()
	}()
	return &api
}

func waitShutdown(serverConf conf.SrvConfig, apiServer *http.Server, serverErrors chan error, shutdown chan os.Signal) error {
	// =========================================================================
	// Shutdown

	// Blocking main and waiting for shutdown.
	select {
	case err := <-serverErrors:
		return errors.Wrap(err, "server error")

	case sig := <-shutdown:
		log.Sugar.Infof("main : %v : Start shutdown", sig)

		// Give outstanding requests a deadline for completion.
		ctx, cancel := context.WithTimeout(context.Background(), serverConf.ShutdownTimeout)
		defer cancel()

		// Asking listener to shutdown and load shed.
		err := apiServer.Shutdown(ctx)
		if err != nil {
			log.Sugar.Infof("main : Graceful shutdown did not complete in %v : %v", serverConf.ShutdownTimeout, err)
			err = apiServer.Close()
		}

		// Log the status of this shutdown.
		switch {
		case sig == syscall.SIGSTOP:
			return errors.New("integrity issue caused shutdown")
		case err != nil:
			return errors.Wrap(err, "could not stop server gracefully")
		}
	}
	return nil
}

