package main

import (
	"context"
	"flag"
	"github.com/PaulSonOfLars/gotgbot/v2/ext/handlers"
	"github.com/PaulSonOfLars/gotgbot/v2/ext/handlers/filters/callbackquery"
	_ "github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	"github.com/mokan-r/jiraffe/internal/jira"
	"github.com/mokan-r/jiraffe/internal/telegram"
	"github.com/mokan-r/jiraffe/pkg/db/postgresql"
	"log"
)

func init() {
	if err := godotenv.Load(); err != nil {
		log.Print("No .env file found")
	}
}

func main() {
	dsn := flag.String(
		"dsn",
		"postgres://jiraffe:jiraffe@localhost:5432/jiraffe",
		"Postgres data source name",
	)
	flag.Parse()

	db, err := openDB(*dsn)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("DB connected")
	defer db.Close()

	jiraClient, err := jira.New()
	if err != nil {
		log.Fatal("Error while creating jira client: " + err.Error())
	}
	telegramClient, err := telegram.New()
	if err != nil {
		panic("Error while creating telegram client: " + err.Error())
	}
	app := application{
		jiraClient:     jiraClient,
		telegramClient: telegramClient,
		DB:             &postgresql.PostgreSQL{DB: db},
	}

	campuses, err := app.DB.GetCampuses()
	if err != nil {
		panic("Couldn't fetch initial campus list from DB")
	}

	campusNamesList := make([]string, 0, len(campuses))

	for _, camp := range campuses {
		campusNamesList = append(campusNamesList, camp.Name)
	}

	app.jiraClient.CampusList = campusNamesList

	app.telegramClient.Updater.Dispatcher.AddHandler(handlers.NewCommand("create", app.HandlerCreate))
	app.telegramClient.Updater.Dispatcher.AddHandler(handlers.NewCommand("register", app.HandlerRegister))
	app.telegramClient.Updater.Dispatcher.AddHandler(handlers.NewCommand("unregister", app.HandlerDeleteUser))
	app.telegramClient.Updater.Dispatcher.AddHandler(handlers.NewCallback(callbackquery.Suffix("start"), app.StartProgressCB))
	app.telegramClient.Updater.Dispatcher.AddHandler(handlers.NewCallback(callbackquery.Suffix("start"), app.StartProgressCB))

	err = app.telegramClient.Start()
	if err != nil {
		log.Fatal("Error while trying to start polling: " + err.Error())
	}

	app.Serve()
}

func openDB(dsn string) (*pgxpool.Pool, error) {
	pool, err := pgxpool.New(context.Background(), dsn)
	if err != nil {
		return nil, err
	}
	if err = pool.Ping(context.Background()); err != nil {
		return nil, err
	}
	return pool, nil
}
