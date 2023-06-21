package telegram

import (
	"fmt"
	tg "github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/PaulSonOfLars/gotgbot/v2/ext"
	"github.com/mokan-r/jiraffe/internal/topics"
	"github.com/mokan-r/jiraffe/pkg/models"
	"log"
	"net/http"
	"os"
	"time"
)

type Client struct {
	Api     *tg.Bot
	Updater *ext.Updater
}

func New() (*Client, error) {
	// Get token from the environment variable
	token := os.Getenv("TELEGRAM_TOKEN")
	if token == "" {
		return nil, fmt.Errorf("TELEGRAM_TOKEN environment variable is empty")
	}

	// Create bot from environment value.
	bot, err := tg.NewBot(token, &tg.BotOpts{
		Client: http.Client{},
		DefaultRequestOpts: &tg.RequestOpts{
			Timeout: time.Second * 20,
			APIURL:  tg.DefaultAPIURL,
		},
	})
	if err != nil {
		panic("failed to create new bot: " + err.Error())
	}

	commands := []tg.BotCommand{
		{
			Command:     "create",
			Description: "Create new topic for Campus. Usage: /create <Topic Name>",
		},
		{
			Command:     "register",
			Description: "Register new user for Campus. Usage /register <LDAP user name>",
		},
		{
			Command:     "unregister",
			Description: "Unregister user from current campus topic. Usage /unregister <LDAP user name>",
		},
	}

	_, err = bot.SetMyCommands(commands, &tg.SetMyCommandsOpts{
		Scope: tg.BotCommandScopeChat{ChatId: topics.ChatID},
	})
	if err != nil {
		log.Fatal(err)
	}

	// Create updater and dispatcher.
	updater := ext.NewUpdater(&ext.UpdaterOpts{
		Dispatcher: ext.NewDispatcher(&ext.DispatcherOpts{
			// If an error is returned by a handler, log it and continue going.
			Error: func(b *tg.Bot, ctx *ext.Context, err error) ext.DispatcherAction {
				log.Println("an error occurred while handling update:", err.Error())
				return ext.DispatcherActionNoop
			},
			MaxRoutines: ext.DefaultMaxRoutines,
		}),
	})

	return &Client{Api: bot, Updater: updater}, nil
}

func (tc *Client) Start() error {
	err := tc.Updater.StartPolling(tc.Api, &ext.PollingOpts{
		DropPendingUpdates: true,
		GetUpdatesOpts: tg.GetUpdatesOpts{
			Timeout:     10,
			RequestOpts: &tg.RequestOpts{Timeout: time.Second * 20},
		},
	})

	return err
}

func (tc *Client) NotifyAboutNewIssue(issue models.Issue, topicID int64) error {
	log.Println(topicID, "\n", issue.String())
	_, err := tc.Api.SendMessage(
		topics.ChatID,
		fmt.Sprintf(issue.String(), tc.Api.User.Username),
		&tg.SendMessageOpts{
			ParseMode:       "Markdown",
			MessageThreadId: topicID,
			ReplyMarkup: tg.InlineKeyboardMarkup{
				InlineKeyboard: [][]tg.InlineKeyboardButton{{
					{Text: "Start progress", CallbackData: issue.Key + "_start"},
				}},
			},
		})
	if err != nil {
		return fmt.Errorf("failed to send new issue message: %w", err)
	}
	return nil
}
