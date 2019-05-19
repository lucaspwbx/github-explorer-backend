package db

import (
	"database/sql"
	"fmt"
	"log"

	"github.com/lib/pq"
)

const (
	host     = "localhost"
	port     = 5432
	userPg   = "lucas"
	password = "teste"
	dbname   = "foo_dev"
)

type Config struct {
	conn *sql.DB
}

func NewConfig(host string, port int, user string, password string, dbName string) *Config {
	connString := fmt.Sprintf("host=%s port=%d user=%s "+"password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbName)
	conn, err := sql.Open("postgres", connString)
	if err != nil {
		panic(err)
	}
	//defer conn.Close()

	err = conn.Ping()
	if err != nil {
		panic(err)
	}

	log.Println("Successfully connected")
	return &Config{conn}
}

func (c *Config) Connection() *sql.DB {
	return c.conn
}

type Project struct {
	Id          int    `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Author      string `json:"author"`
	Language    string `json:"language"`
	Url         string `json:"url"`
}

type User struct {
	Id               int    `json:"id"`
	Username         string `json:"username"`
	Password         string `json:"password"`
	Email            string `json:"email"`
	Languages        string `json:"languages"`
	Frequency        string `json:"frequency"`
	FavoriteLanguage string `json:"favorite_language"`
}

func AddProject(db *sql.DB, proj *Project) (int, error) {
	sqlStmt := `INSERT INTO projects(name, description, author, language, url) VALUES ($1, $2, $3, $4, $5) RETURNING id`
	err := db.QueryRow(sqlStmt, proj.Name, proj.Description, proj.Author, proj.Language, proj.Url).Scan(&proj.Id)
	if err != nil {
		log.Println("Add project", err)
		return 0, err
	}
	log.Println("New record is is: ", proj.Id)
	return proj.Id, nil
}

func BookmarkProject(db *sql.DB, userId int, projectId int) (bool, error) {
	sqlStmt := `INSERT INTO bookmarked_projects(user_id, project_id) VALUES ($1, $2)`
	_, err := db.Exec(sqlStmt, userId, projectId)
	if err != nil {
		if err, ok := err.(*pq.Error); ok {
			if err.Code == "23505" {
				log.Println("Project is already bookmarked!")
				return false, err
			}
		}
		log.Fatal("Bookmark project: ", err)
	}
	return true, nil
}

func ProjectExists(db *sql.DB, name string, author string, language string) (int, error) {
	sqlStmt := `SELECT id FROM projects WHERE name = $1 AND author = $2 AND language = $3`
	id := 0
	err := db.QueryRow(sqlStmt, name, author, language).Scan(&id)
	if err != nil {
		log.Println("Project exists: ", err)
		return 0, err
	}
	return id, nil
}

func ConfirmUser(db *sql.DB, u *User) (*User, error) {
	sqlStmt := `SELECT id, languages, frequency, favorite_language FROM users WHERE username = $1 AND email = $2 AND password = $3;`
	user := &User{}
	err := db.QueryRow(sqlStmt, u.Username, u.Email, u.Password).Scan(&user.Id, &user.Languages, &user.Frequency, &user.FavoriteLanguage)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func FetchUserBookmarkedProjects(db *sql.DB, userId int) []Project {
	var projsFromUser []Project
	projsFromUser = []Project{}
	rows, err := db.Query(`
		select p.id, p.name, p.description, p.author, p.language, p.url from projects p inner join bookmarked_projects bp on p.id = bp.project_id where bp.user_id = $1`, userId)
	if err != nil {
		log.Println(err)
	}
	defer rows.Close()
	for rows.Next() {
		project := Project{}
		err := rows.Scan(
			&project.Id,
			&project.Name,
			&project.Description,
			&project.Author,
			&project.Language,
			&project.Url)
		if err != nil {
			log.Println(err)
		}
		projsFromUser = append(projsFromUser, project)
	}
	err = rows.Err()
	if err != nil {
		log.Println(err)
	}
	return projsFromUser
}
