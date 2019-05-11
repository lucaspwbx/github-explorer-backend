package main

import (
	"database/sql"
	"fmt"

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
	Name        string
	Description string
	Author      string
	Language    string
	Url         string
}

type user struct {
	Id       int
	Username string
	Password string
	Email    string
}

func addProject(db *sql.DB, proj *project) {
	sqlStmt := `INSERT INTO projects(name, description, author, language, url) VALUES ($1, $2, $3, $4, $5) RETURNING id`
	//id := 0
	err := db.QueryRow(sqlStmt, proj.Name, proj.Description, proj.Author, proj.Language, proj.Url).Scan(&proj.Id)
	if err != nil {
		panic(err)
	}
	fmt.Println("New record is is: ", proj.Id)
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
	fmt.Println(projects)

	var users []user
	rows, err = db.Query(`
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
	fmt.Println(users)

	var projFromuser []project
	rows, err = db.Query(`
		select p.id, p.name, p.description, p.author, p.language, p.url from projects p inner join bookmarked_projects bp on p.id = bp.project_id where bp.user_id = 3;
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
		projFromuser = append(projFromuser, project)
	}
	err = rows.Err()
	if err != nil {
		panic(err)
	}
	fmt.Println(projFromuser)

	//addProject(db, &project{Name: "teste", Description: "Too", Author: "Ulver", Language: "JS", Url: "www.language.com"})
}
