package main

import (
	"context"
	"flag"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"mauk14.library/internal/data"
	"mauk14.library/internal/jsonlog"
	"mauk14.library/internal/mailer"
	"os"
	"sync"
	"time"
)

type config struct {
	port    int
	env     string
	isMongo bool
	limiter struct {
		rps     float64
		burst   int
		enabled bool
	}
	db struct {
		dsn string
	}
	smtp struct {
		host     string
		port     int
		username string
		password string
		sender   string
	}
}

const (
	version = "1.0.0"
)

type application struct {
	config config
	logger *jsonlog.Logger
	models data.Models
	mailer mailer.Mailer
	wg     sync.WaitGroup
}

func main() {
	var cfg config

	flag.IntVar(&cfg.port, "port", 4000, "API server port")
	flag.StringVar(&cfg.env, "env", "development", "Environment (development|staging|production)")

	flag.StringVar(&cfg.db.dsn, "db-dsn", "POSTGRES_URI", "PostgreSQL DSN")

	flag.Float64Var(&cfg.limiter.rps, "limiter-rps", 2, "Rate limiter maximum requests per second")
	flag.IntVar(&cfg.limiter.burst, "limiter-burst", 4, "Rate limiter maximum burst")
	flag.BoolVar(&cfg.limiter.enabled, "limiter-enabled", true, "Enable rate limiter")

	flag.BoolVar(&cfg.isMongo, "mongo", false, "Mongo use or not?")

	flag.StringVar(&cfg.smtp.host, "smtp-host", "SMTP_HOST", "SMTP host")
	flag.IntVar(&cfg.smtp.port, "smtp-port", 587, "SMTP port")
	flag.StringVar(&cfg.smtp.username, "smtp-username", "SMTP_USERNAME", "SMTP username")
	flag.StringVar(&cfg.smtp.password, "smtp-password", "SMTP_PASSWORD", "SMTP password")
	flag.StringVar(&cfg.smtp.sender, "smtp-sender", "SMTP_SENDER", "_SMTP sender")

	flag.Parse()

	logger := jsonlog.New(os.Stdout, jsonlog.LevelInfo)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var db data.DB

	if cfg.isMongo {
		client, err := clientMongoDb(os.Getenv("MONGODB_URI"))

		if err != nil {
			logger.PrintFatal(err, nil)
		}

		defer func(client *mongo.Client, ctx context.Context) {
			err := client.Disconnect(ctx)
			if err != nil {
				logger.PrintFatal(err, nil)
			}
		}(client, ctx)

		db = data.NewMongo(client, "Library")
	} else {
		database, err := OpenDb(cfg)
		if err != nil {
			logger.PrintFatal(err, nil)
		}
		db = database

	}
	logger.PrintInfo("database connection pool established", nil)

	app := &application{
		config: cfg,
		logger: logger,
		models: data.NewModels(db),
		mailer: mailer.New(cfg.smtp.host, cfg.smtp.port, cfg.smtp.username, cfg.smtp.password, cfg.smtp.sender),
	}

	err := app.serve()
	if err != nil {
		logger.PrintFatal(err, nil)
	}

}

func clientMongoDb(uri string) (*mongo.Client, error) {
	client, err := mongo.NewClient(options.Client().ApplyURI(uri))
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)

	defer cancel()

	err = client.Connect(ctx)
	if err != nil {
		return nil, err
	}
	err = client.Ping(ctx, readpref.Primary())
	if err != nil {
		return nil, err
	}

	return client, nil
}

func OpenDb(cfg config) (*data.Postgres, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	db, err := pgxpool.New(ctx, cfg.db.dsn)
	if err != nil {
		return nil, err
	}

	err = db.Ping(ctx)
	if err != nil {
		return nil, err
	}

	return &data.Postgres{DB: db}, nil
}
