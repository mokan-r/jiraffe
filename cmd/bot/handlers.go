package main

import (
	"fmt"
	tg "github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/PaulSonOfLars/gotgbot/v2/ext"
	"github.com/mokan-r/jiraffe/internal/topics"
	"log"
	"strings"
)

// HandlerCreate creates new topic and inserts campus in DB
func (app *application) HandlerCreate(b *tg.Bot, ctx *ext.Context) error {
	if !checkAdministrator(b, ctx) {
		return nil
	}
	args := ctx.Args()
	if len(args) != 2 {
		_, err := b.SendMessage(
			ctx.EffectiveChat.Id,
			fmt.Sprint("Invalid number of arguments. Expected: /create <string>"),
			nil,
		)
		return err
	}

	campuses, err := app.DB.GetCampuses()
	if err != nil {
		_, err := b.SendMessage(
			ctx.EffectiveChat.Id,
			fmt.Sprint("Error while fetching campuses from database"),
			nil,
		)
		return err
	}

	for _, c := range campuses {
		if c.Name == args[1] {
			_, err := b.SendMessage(
				ctx.EffectiveChat.Id,
				fmt.Sprintf("Campus *%s* already exists", args[1]),
				&tg.SendMessageOpts{ParseMode: "markdown"},
			)
			return err
		}
	}

	topic, err := b.CreateForumTopic(topics.ChatID, args[1], nil)
	err = app.DB.InsertCampus(topic.Name, topic.MessageThreadId)

	if err != nil {
		_, err = b.SendMessage(
			ctx.EffectiveChat.Id,
			fmt.Sprintf("Error while insertin new campus to DB"),
			&tg.SendMessageOpts{ParseMode: "markdown"},
		)
		return err
	}
	_, err = b.SendMessage(
		ctx.EffectiveChat.Id,
		fmt.Sprintf("Topic *%s* created successfully", topic.Name),
		&tg.SendMessageOpts{ParseMode: "markdown"},
	)

	return nil
}

// HandlerRegister creates new user for current topic
func (app *application) HandlerRegister(b *tg.Bot, ctx *ext.Context) error {
	if !checkAdministrator(b, ctx) {
		return nil
	}
	args := ctx.Args()
	if len(args) != 2 {
		_, err := b.SendMessage(
			ctx.EffectiveChat.Id,
			fmt.Sprintf("Invalid number of arguments. Expected: /register <string>"),
			&tg.SendMessageOpts{
				MessageThreadId: ctx.EffectiveMessage.MessageThreadId,
				ParseMode:       "markdown",
			},
		)
		return err
	}
	messageInCampusThread := false
	campuses, err := app.DB.GetCampuses()
	for _, c := range campuses {
		if c.TopicID == ctx.Message.MessageThreadId {
			messageInCampusThread = true
			break
		}
	}

	if !messageInCampusThread {
		_, err := b.SendMessage(
			ctx.EffectiveChat.Id,
			fmt.Sprintf("You can register only inside campus topics"),
			&tg.SendMessageOpts{
				MessageThreadId: ctx.EffectiveMessage.MessageThreadId,
				ParseMode:       "markdown",
			},
		)
		return err
	}

	users, err := app.DB.GetUsers()
	if err != nil {
		_, err := b.SendMessage(
			ctx.EffectiveChat.Id,
			fmt.Sprint("Error while fetching campuses from database"),
			nil,
		)
		return err
	}

	for _, c := range users {
		if c.Name == args[1] && c.Campus.TopicID == ctx.EffectiveMessage.MessageThreadId {
			_, err := b.SendMessage(
				ctx.EffectiveChat.Id,
				fmt.Sprintf("User *%s* already exists in this campus thread", args[1]),
				&tg.SendMessageOpts{ParseMode: "markdown"},
			)
			return err
		}
	}

	err = app.DB.InsertUser(args[1], ctx.EffectiveMessage.MessageThreadId)

	if err != nil {
		_, err = b.SendMessage(
			ctx.EffectiveChat.Id,
			fmt.Sprintf("Error while insertin new user to DB"),
			&tg.SendMessageOpts{
				MessageThreadId: ctx.EffectiveMessage.MessageThreadId,
				ParseMode:       "markdown",
			},
		)
		return err
	}

	_, err = b.SendMessage(
		ctx.EffectiveChat.Id,
		fmt.Sprintf("User *%s* created successfully", args[1]),
		&tg.SendMessageOpts{
			MessageThreadId: ctx.EffectiveMessage.MessageThreadId,
			ParseMode:       "markdown",
		},
	)

	if err != nil {
		return err
	}

	return nil
}

func (app *application) HandlerDeleteUser(b *tg.Bot, ctx *ext.Context) error {
	if !checkAdministrator(b, ctx) {
		return nil
	}
	args := ctx.Args()
	if len(args) != 2 {
		_, err := b.SendMessage(
			ctx.EffectiveChat.Id,
			"Invalid number of arguments. Expected: /unregister <string>",
			&tg.SendMessageOpts{
				MessageThreadId: ctx.EffectiveMessage.MessageThreadId,
				ParseMode:       "markdown",
			},
		)
		return err
	}

	err := app.DB.DeleteUser(args[1])
	if err != nil {
		return err
	}
	return nil
}

func (app *application) StartProgressCB(b *tg.Bot, ctx *ext.Context) error {
	cb := ctx.Update.CallbackQuery

	issueKey := strings.Split(cb.Data, "_")[0]

	_, _, err := cb.Message.EditReplyMarkup(b, &tg.EditMessageReplyMarkupOpts{
		ReplyMarkup: tg.InlineKeyboardMarkup{
			InlineKeyboard: [][]tg.InlineKeyboardButton{{
				{Text: "Low", CallbackData: issueKey + "_low_priority"},
				{Text: "Medium", CallbackData: issueKey + "_medium_priority"},
			}},
		},
	})
	if err != nil {
		return fmt.Errorf("failed to edit start message text: %w", err)
	}

	_, err = cb.Answer(b, &tg.AnswerCallbackQueryOpts{
		Text: "Progress starting...",
	})

	if err != nil {
		return fmt.Errorf("failed to answer start callback query: %w", err)
	}

	return nil
}

func (app *application) PriorityCB(b *tg.Bot, ctx *ext.Context) error {
	cb := ctx.Update.CallbackQuery

	issueKey := strings.Split(cb.Data, "_")[0]
	priority := strings.Split(cb.Data, "_")[1]

	issue, err := app.jiraClient.GetIssue(issueKey)
	if err != nil {
		return err
	}

	issue.Priority = priority

	_, _, err = cb.Message.EditText(b, issue.String(), &tg.EditMessageTextOpts{ParseMode: "markdown"})
	if err != nil {
		return err
	}

	markup, err := app.getAssignersInlineKeyboardMarkup(cb.Message.MessageThreadId, priority, issueKey)
	if err != nil {
		return fmt.Errorf("failed to get reply markup with assigners: %w", err)
	}
	_, _, err = cb.Message.EditReplyMarkup(b, &tg.EditMessageReplyMarkupOpts{
		ReplyMarkup: markup,
	})
	if err != nil {
		return fmt.Errorf("failed to edit reply markup: %w", err)
	}

	_, err = cb.Answer(b, &tg.AnswerCallbackQueryOpts{
		Text: "Priority picked",
	})

	if err != nil {
		return fmt.Errorf("failed to answer priority callback query: %w", err)
	}

	return nil
}

func (app *application) AssignerCB(b *tg.Bot, ctx *ext.Context) error {
	cb := ctx.Update.CallbackQuery

	issueKey := strings.Split(cb.Data, "_")[0]
	priority := strings.Split(cb.Data, "_")[1]
	assigner := strings.Split(cb.Data, "_")[2]

	issue, err := app.jiraClient.GetIssue(issueKey)
	if err != nil {
		return err
	}

	issue.Priority = priority
	issue.Assignee = assigner

	_, _, err = cb.Message.EditText(b, issue.String(), &tg.EditMessageTextOpts{ParseMode: "markdown"})
	if err != nil {
		return err
	}

	answer, err := app.jiraClient.TransitionIssue(&issue)
	if err != nil {
		return err
	}

	_, err = cb.Answer(b, &tg.AnswerCallbackQueryOpts{
		Text: answer,
	})

	if err != nil {
		return fmt.Errorf("failed to answer assigner callback query: %w", err)
	}

	return nil
}

func (app *application) getAssignersInlineKeyboardMarkup(campusID int64, issueKey string, priority string) (tg.InlineKeyboardMarkup, error) {
	users, err := app.DB.GetUsersInCampus(campusID)
	if err != nil {
		return tg.InlineKeyboardMarkup{}, err
	}
	buttons := make([]tg.InlineKeyboardButton, 0, len(users))
	for _, user := range users {
		buttons = append(buttons, tg.InlineKeyboardButton{
			Text:         user.Name,
			CallbackData: issueKey + "_" + priority + "_" + user.Name + "_assign",
		})
	}
	return tg.InlineKeyboardMarkup{InlineKeyboard: [][]tg.InlineKeyboardButton{buttons}}, nil
}

func checkAdministrator(b *tg.Bot, ctx *ext.Context) bool {
	administrators, _ := ctx.EffectiveChat.GetAdministrators(b, nil)
	for _, a := range administrators {
		if ctx.EffectiveSender.Id() == a.GetUser().Id {
			return true
		}
	}
	_, err := b.SendMessage(
		ctx.EffectiveChat.Id,
		"You must be admin of the group to execute this command",
		&tg.SendMessageOpts{
			MessageThreadId: ctx.EffectiveMessage.MessageThreadId,
			ParseMode:       "markdown",
		},
	)
	if err != nil {
		log.Println(err)
	}
	return false
}
