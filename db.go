package main

import (
	"database/sql"
	"fmt"
	"net/http"

	"github.com/labstack/echo"
	_ "github.com/lib/pq"
)

const (
	host     = "localhost"
	port     = 5432
	userPg   = "lucas"
	password = "teste"
	dbname   = "foo_dev"
)

type project struct {
	Id          int
	Name        string `json:"name"`
	Description string `json:"description"`
	Author      string `json:"author"`
	Language    string `json:"language"`
	Url         string `json:"url"`
}

type user struct {
	Id       int
	Username string
	Password string
	Email    string
}

func addProject(db *sql.DB, proj *project) error {
	sqlStmt := `INSERT INTO projects(name, description, author, language, url) VALUES ($1, $2, $3, $4, $5) RETURNING id`
	err := db.QueryRow(sqlStmt, proj.Name, proj.Description, proj.Author, proj.Language, proj.Url).Scan(&proj.Id)
	if err != nil {
		panic(err)
	}
	fmt.Println("New record is is: ", proj.Id)
	return nil
}

func bookmarkProject(db *sql.DB, userId int, projectId int) error {
	sqlStmt := `INSERT INTO bookmarked_projects(user_id, project_id) VALUES ($1, $2)`
	_, err := db.Exec(sqlStmt, userId, projectId)
	if err != nil {
		panic(err)
	}
	return nil
}

func projectExists(db *sql.DB, name string, author string, language string) (int, error) {
	sqlStmt := `SELECT id FROM projects WHERE name = $1 AND author = $2 AND language = $3`
	id := 0
	err := db.QueryRow(sqlStmt, name, author, language).Scan(&id)
	if err != nil {
		panic(err)
		return 0, err
	}
	return id, nil
}

func fetchProjects(db *sql.DB) []project {
	var projects []project
	rows, err := db.Query(`
		SELECT id, name, description, author, language, url from projects
	`)
	if err != nil {
		panic(err)
	}
	defer rows.Close()
	for rows.Next() {
		project := project{}
		err := rows.Scan(
			&project.Id,
			&project.Name,
			&project.Description,
			&project.Author,
			&project.Language,
			&project.Url)
		if err != nil {
			panic(err)
		}
		projects = append(projects, project)
	}
	err = rows.Err()
	if err != nil {
		panic(err)
	}
	return projects
}

func fetchUsers(db *sql.DB) []user {
	var users []user
	rows, err := db.Query(`
		SELECT id, username, password, email from users`)
	if err != nil {
		panic(err)
	}
	defer rows.Close()
	for rows.Next() {
		us := user{}
		err = rows.Scan(
			&us.Id,
			&us.Username,
			&us.Password,
			&us.Email)
		if err != nil {
			panic(err)
		}
		users = append(users, us)
	}
	err = rows.Err()
	if err != nil {
		panic(err)
	}
	return users
}

func fetchUserBookmarkedProjects(db *sql.DB, userId int) []project {
	var projsFromUser []project
	rows, err := db.Query(`
		select p.id, p.name, p.description, p.author, p.language, p.url from projects p inner join bookmarked_projects bp on p.id = bp.project_id where bp.user_id = $1`, userId)
	if err != nil {
		panic(err)
	}
	defer rows.Close()
	for rows.Next() {
		project := project{}
		err := rows.Scan(
			&project.Id,
			&project.Name,
			&project.Description,
			&project.Author,
			&project.Language,
			&project.Url)
		if err != nil {
			panic(err)
		}
		projsFromUser = append(projsFromUser, project)
	}
	err = rows.Err()
	if err != nil {
		panic(err)
	}
	return projsFromUser
}

func main() {
	connString := fmt.Sprintf("host=%s port=%d user=%s "+"password=%s dbname=%s sslmode=disable",
		host, port, userPg, password, dbname)
	db, err := sql.Open("postgres", connString)
	if err != nil {
		panic(err)
	}
	defer db.Close()

	err = db.Ping()
	if err != nil {
		panic(err)
	}

	fmt.Println("Successfully connected")

	e := echo.New()
	e.GET("/", func(c echo.Context) error {
		return c.String(http.StatusOK, "Hello World")
	})
	e.POST("/users/:id/bookmarked_projects", func(c echo.Context) error {
		p := &project{}
		//	id := c.Param("id")
		if err := c.Bind(p); err != nil {
			return err
		}
		return c.JSON(http.StatusCreated, p)
	})
	e.Logger.Fatal(e.Start(":1323"))
}

//addProject(db, &project{Name: "teste", Description: "Too", Author: "Ulver", Language: "JS", Url: "www.language.com"})
//bookmarkProject(db)
//id, err := projectExists(db, "alco", "miasma", "elixir")
//fmt.Println(id)
//fmt.Println(fetchUserBookmarkedProjects(db, 3))
