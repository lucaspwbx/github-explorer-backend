package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"

	"github.com/labstack/echo"
	"github.com/lib/pq"
	"golang.org/x/crypto/bcrypt"
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
	Username string `"json:username"`
	Password string `"json:password"`
	Email    string `"json:email"`
}

func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	return string(bytes), err
}

func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

func addProject(db *sql.DB, proj *project) (int, error) {
	sqlStmt := `INSERT INTO projects(name, description, author, language, url) VALUES ($1, $2, $3, $4, $5) RETURNING id`
	err := db.QueryRow(sqlStmt, proj.Name, proj.Description, proj.Author, proj.Language, proj.Url).Scan(&proj.Id)
	if err != nil {
		//	panic(err)
		log.Fatal("Add project", err)
		return 0, err
	}
	fmt.Println("New record is is: ", proj.Id)
	return proj.Id, nil
}

func bookmarkProject(db *sql.DB, userId int, projectId int) (bool, error) {
	sqlStmt := `INSERT INTO bookmarked_projects(user_id, project_id) VALUES ($1, $2)`
	_, err := db.Exec(sqlStmt, userId, projectId)
	if err != nil {
		//panic(err)
		if err, ok := err.(*pq.Error); ok {
			if err.Code == "23505" {
				log.Println("Project is already bookmarked")
				return false, err
			}
		}
		log.Fatal("Bookmark project: ", err)
	}
	return true, nil
}

func projectExists(db *sql.DB, name string, author string, language string) (int, error) {
	sqlStmt := `SELECT id FROM projects WHERE name = $1 AND author = $2 AND language = $3`
	id := 0
	err := db.QueryRow(sqlStmt, name, author, language).Scan(&id)
	if err != nil {
		//	log.Fatal("Project exists: ", err)
		//	panic(err)
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

func addNewUserHandler(c echo.Context) error {
	u := &user{}
	db := getDB()
	if err := c.Bind(u); err != nil {
		return c.JSON(http.StatusBadRequest, "Error signing up new user - 1")
	}
	hash, err := HashPassword(u.Password)
	if err != nil {
		log.Fatal("Error hashing password")
		return c.JSON(http.StatusBadRequest, "Error signing up new user - 2")
	}
	u.Password = hash
	log.Println(u)
	sqlStmt := `INSERT INTO users(username, password, email, created_on) VALUES ($1, $2, $3, $4) RETURNING id`
	err = db.QueryRow(sqlStmt, u.Username, u.Password, u.Email, "now()").Scan(&u.Id)
	if err != nil {
		log.Println(err)
		return c.JSON(http.StatusBadRequest, "Error signing up new user - 3")
	}
	log.Println("New user is id: ", u.Id)
	return c.JSON(http.StatusOK, "OK")
}

func loginUserHandler(c echo.Context) error {
	return nil
}

func logoutUserHandler(c echo.Context) error {
	return nil
}

func getDB() *sql.DB {
	connString := fmt.Sprintf("host=%s port=%d user=%s "+"password=%s dbname=%s sslmode=disable",
		host, port, userPg, password, dbname)
	db, err := sql.Open("postgres", connString)
	if err != nil {
		panic(err)
	}
	//defer db.Close()

	err = db.Ping()
	if err != nil {
		panic(err)
	}

	log.Println("Successfully connected")
	return db
}

func main() {

	e := echo.New()
	e.GET("/", func(c echo.Context) error {
		return c.String(http.StatusOK, "Hello World")
	})
	e.POST("/users", addNewUserHandler)
	e.POST("/users/:id/login", loginUserHandler)
	e.POST("/users/:id/logout", logoutUserHandler)
	//e.POST("/users/:id/bookmarked_projects", func(c echo.Context) error {
	//p := &project{}
	//userId, err := strconv.Atoi(c.Param("id"))
	//if err != nil {
	//return err
	//}
	//if err := c.Bind(p); err != nil {
	//return err
	//}
	//projectId, err := projectExists(db, p.Name, p.Author, p.Language)
	//if err != nil {
	//newProjectId, err := addProject(db, p)
	//if err != nil {
	//return err
	//}
	//_, err = bookmarkProject(db, userId, newProjectId)
	//if err != nil {
	//log.Println("Problems bookmarking project 1")
	//return c.JSON(http.StatusBadRequest, "Failed to bookmark project -> 1")
	//}
	//return c.JSON(http.StatusOK, "Bookmarked project 1")
	//}
	//_, err = bookmarkProject(db, userId, projectId)
	//if err != nil {
	//log.Println("Problems bookmarking project 2")
	//return c.JSON(http.StatusBadRequest, "Failed to bookmark project -> 2")
	//}
	//return c.JSON(http.StatusOK, "bookmarked project 2")
	//})
	e.Logger.Fatal(e.Start(":1323"))
}

//addProject(db, &project{Name: "teste", Description: "Too", Author: "Ulver", Language: "JS", Url: "www.language.com"})
//bookmarkProject(db)
//id, err := projectExists(db, "alco", "miasma", "elixir")
//fmt.Println(id)
//fmt.Println(fetchUserBookmarkedProjects(db, 3))
