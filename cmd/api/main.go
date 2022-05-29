package main

import (
	"context"
	"database/sql"
	"flag"
	"github.com/joho/godotenv"
	"github.com/xuzheng465/greenlight/internal/data"
	"github.com/xuzheng465/greenlight/internal/jsonlog"
	"github.com/xuzheng465/greenlight/internal/mailer"
	"log"
	"os"
	"strconv"
	"sync"
	"time"

	_ "github.com/lib/pq"
)

const version = "1.0.0"

// Define a config struct to hold all the configuration settings for our application.
// For now, the only configuration settings will be the network port that we want the
// server to listen on, and the name of the current operating environment for the application
// We will read in these configuration settings from command-line flags when the application starts.
type config struct {
	port int
	env  string
	db   struct {
		dsn          string
		maxOpenConns int
		maxIdleConns int
		maxIdleTime  string
	}
	limiter struct {
		rps     float64
		burst   int
		enabled bool
	}
	smtp struct {
		host     string
		port     int
		username string
		password string
		sender   string
	}
}

// Define an application struct to hold the dependencies for our HTTP handlers, helpers,
// and middleware. At the moment this only contains a copy of the config struct and a logger
type application struct {
	config config
	logger *jsonlog.Logger
	models data.Models
	mailer mailer.Mailer
	wg     sync.WaitGroup
}

func main() {
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatalf("Error loading .env file")
	}

	var cfg config

	flag.IntVar(&cfg.port, "port", 4000, "API server port")
	flag.StringVar(&cfg.env, "env", "development", "ENV (development|staging|production)")
	flag.IntVar(&cfg.db.maxOpenConns, "db-max-open-conns", 100, "DB max open connections")
	flag.IntVar(&cfg.db.maxIdleConns, "db-max-idle-conns", 100, "DB max idle connections")
	flag.StringVar(&cfg.db.maxIdleTime, "db-max-idle-time", "15m", "DB max idle time")
	flag.Float64Var(&cfg.limiter.rps, "limiter-rps", 2, "Rate limiter maximum requests per second")
	flag.IntVar(&cfg.limiter.burst, "limiter-burst", 4, "Rate limiter maximum burst size")
	flag.BoolVar(&cfg.limiter.enabled, "limiter-enabled", true, "Rate limiter enabled")
	//flag.StringVar(&cfg.db.dsn, "db-dsn", os.Getenv("dsn"), "PostgreSQL DSN")
	cfg.db.dsn = os.Getenv("dsn")

	smtpPort, _ := strconv.Atoi(os.Getenv("smtp_port"))
	// mail config
	flag.StringVar(&cfg.smtp.host, "smtp-host", os.Getenv("smtp_host"), "SMTP host")
	flag.IntVar(&cfg.smtp.port, "smtp-port", smtpPort, "SMTP port")
	flag.StringVar(&cfg.smtp.username, "smtp-username", os.Getenv("smtp_username"), "SMTP username")
	flag.StringVar(&cfg.smtp.password, "smtp-password", os.Getenv("smtp_password"), "SMTP password")
	flag.StringVar(&cfg.smtp.sender, "smtp-sender", os.Getenv("smtp_sender"), "SMTP sender")

	flag.Parse()
	// Initialize a new logger which writes messages to the standard out stream.
	// Prefixed with the current date and time
	logger := jsonlog.New(os.Stdout, jsonlog.LevelInfo)

	// Connect to the database
	db, err := openDB(cfg)
	if err != nil {
		logger.PrintFatal(err, nil)
	}
	defer db.Close()
	logger.PrintInfo("Database connection pool established", nil)

	// Initialize a new application struct, passing in the config struct and the logger
	app := &application{
		config: cfg,
		logger: logger,
		models: data.NewModels(db),
		mailer: mailer.New(cfg.smtp.host, cfg.smtp.port, cfg.smtp.username, cfg.smtp.password, cfg.smtp.sender),
	}

	err = app.serve()
	if err != nil {
		logger.PrintFatal(err, nil)
	}

}

// openDB opens a connection to the database using the config settings.
func openDB(cfg config) (*sql.DB, error) {
	db, err := sql.Open("postgres", cfg.db.dsn)
	if err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(cfg.db.maxOpenConns)

	duration, err := time.ParseDuration(cfg.db.maxIdleTime)
	if err != nil {
		return nil, err
	}
	db.SetConnMaxIdleTime(duration)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	err = db.PingContext(ctx)
	if err != nil {
		return nil, err
	}
	return db, nil
}
