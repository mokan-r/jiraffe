package postgresql

import (
	"context"
	"errors"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/mokan-r/jiraffe/pkg/models"
	"log"
)

type PostgreSQL struct {
	DB *pgxpool.Pool
}

func (p *PostgreSQL) InsertIssue(issue models.Issue) error {
	stmt := `
		insert into issues (
			tg_message_id,
			key,
			link,
			priority,
			summary,
			description,
			reporter,
			assignee,
		    campus_id
		) values(
		         $1, $2, $3, $4, $5, $6, $7, $8, (select id from campus where name = $9)
		) returning id;
	`
	conn, err := p.DB.Acquire(context.Background())
	if err != nil {
		return err
	}
	defer conn.Release()
	conn.QueryRow(
		context.Background(),
		stmt,
		issue.TgMessageID,
		issue.Key,
		issue.Link,
		issue.Priority,
		issue.Summary,
		issue.Description,
		issue.Reporter,
		issue.Assignee,
		issue.Campus,
	)

	return nil
}

func (p *PostgreSQL) InsertUser(userName string, campusID int64) error {
	stmt := `INSERT INTO users (name, campus_id) VALUES ($1, (select id from campus c where c.topic_id = $2));`
	conn, err := p.DB.Acquire(context.Background())
	if err != nil {
		return err
	}
	defer conn.Release()

	conn.QueryRow(context.Background(), stmt, userName, campusID)

	return nil
}

func (p *PostgreSQL) DeleteUser(userName string) error {
	stmt := `DELETE FROM users where name = $1`
	conn, err := p.DB.Acquire(context.Background())
	if err != nil {
		return err
	}
	defer conn.Release()

	exec, err := conn.Exec(context.Background(), stmt, userName)

	log.Println(exec.RowsAffected())

	return err
}

func (p *PostgreSQL) GetIssuesKeys() ([]models.Issue, error) {
	stmt := `
		SELECT
		key
		FROM issues;
	`

	conn, err := p.DB.Acquire(context.Background())
	if err != nil {
		return nil, err
	}
	defer conn.Release()

	rows, err := conn.Query(context.Background(), stmt)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var issues []models.Issue

	for rows.Next() {
		i := models.Issue{}
		err := rows.Scan(&i.Key)
		if err != nil {
			return nil, err
		}
		issues = append(issues, i)
	}

	return issues, nil
}

func (p *PostgreSQL) GetUsers() ([]models.User, error) {
	stmt := `SELECT u.name, c.name, c.topic_id FROM users u join campus c on u.campus_id = c.id;`
	conn, err := p.DB.Acquire(context.Background())
	if err != nil {
		return nil, err
	}
	defer conn.Release()

	rows, err := conn.Query(context.Background(), stmt)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var users []models.User

	for rows.Next() {
		user := models.User{}
		err := rows.Scan(&user.Id, &user.Name, &user.Campus.TopicID)
		if err != nil {
			return nil, err
		}
		users = append(users, user)
	}

	return users, nil
}

func (p *PostgreSQL) GetUsersInCampus(campusID int64) ([]models.User, error) {
	stmt := `SELECT u.name, c.name, c.topic_id FROM users u join campus c on u.campus_id = c.id where c.topic_id = $1;`
	conn, err := p.DB.Acquire(context.Background())
	if err != nil {
		return nil, err
	}
	defer conn.Release()

	rows, err := conn.Query(context.Background(), stmt, campusID)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var users []models.User

	for rows.Next() {
		user := models.User{}
		err := rows.Scan(&user.Id, &user.Name, &user.Campus.TopicID)
		if err != nil {
			return nil, err
		}
		users = append(users, user)
	}

	return users, nil
}

func (p *PostgreSQL) InsertCampus(name string, topicID int64) error {
	stmt := `INSERT INTO campus (name, topic_id) VALUES ($1, $2);`
	conn, err := p.DB.Acquire(context.Background())
	if err != nil {
		return err
	}
	defer conn.Release()

	conn.QueryRow(context.Background(), stmt, name, topicID)

	return nil
}

func (p *PostgreSQL) GetCampuses() ([]models.Campus, error) {
	stmt := `SELECT id, name, topic_id FROM campus;`
	conn, err := p.DB.Acquire(context.Background())
	if err != nil {
		log.Println(err)
		return nil, err
	}
	defer conn.Release()

	rows, err := conn.Query(context.Background(), stmt)

	if err != nil {
		log.Println(err)
		return nil, err
	}

	defer rows.Close()

	campuses := make([]models.Campus, 0)

	for rows.Next() {
		campus := models.Campus{}
		err := rows.Scan(&campus.Id, &campus.Name, &campus.TopicID)
		if err != nil {
			log.Println("Scanning from rows err:", err)
			return nil, err
		}
		campuses = append(campuses, campus)
	}

	return campuses, nil
}

func (p *PostgreSQL) IsIssueExists(issueKey string) (bool, error) {
	stmt := `SELECT id FROM issues where key = $1;`
	conn, err := p.DB.Acquire(context.Background())
	if err != nil {
		log.Println(err)
		return false, err
	}
	defer conn.Release()

	row := conn.QueryRow(context.Background(), stmt, issueKey)
	var id int
	err = row.Scan(&id)
	if err != nil && errors.Is(err, pgx.ErrNoRows) {
		return false, nil
	}
	return true, err
}

func (p *PostgreSQL) GetCampusID(campusName string) (int64, error) {
	stmt := `SELECT topic_id FROM campus where name = $1;`
	conn, err := p.DB.Acquire(context.Background())
	if err != nil {
		log.Println(err)
		return 0, err
	}
	defer conn.Release()

	row := conn.QueryRow(context.Background(), stmt, campusName)
	var id int64
	err = row.Scan(&id)
	if err != nil && errors.Is(err, pgx.ErrNoRows) {
		return 0, nil
	}
	return id, err
}
