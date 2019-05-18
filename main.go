package main

import (
	"log"
	"net/http"
	"strconv"
	"teste/db"
	"teste/service"
	"teste/util"

	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
)

const (
	host     = "localhost"
	port     = 5432
	userPg   = "lucas"
	password = "teste"
	dbname   = "foo_dev"
)

var (
	config *db.Config
)

func addNewUserHandler(c echo.Context) error {
	u := &db.User{}
	if err := c.Bind(u); err != nil {
		return c.JSON(http.StatusBadRequest, "Error signing up new user - 1")
	}
	hash, err := util.HashPassword(u.Password)
	if err != nil {
		log.Println("Error hashing password")
		return c.JSON(http.StatusBadRequest, "Error signing up new user - 2")
	}
	u.Password = hash
	sqlStmt := `INSERT INTO users(username, password, email, created_on) VALUES ($1, $2, $3, $4) RETURNING id`
	err = config.Connection().QueryRow(sqlStmt, u.Username, u.Password, u.Email, "now()").Scan(&u.Id)
	if err != nil {
		log.Println(err)
		return c.JSON(http.StatusBadRequest, "Error signing up new user - 3")
	}
	log.Println("New user is id: ", u.Id)
	if err = service.SendEmail(u.Username, u.Email); err != nil {
		log.Println("Error sending welcome email!")
	}

	return c.JSON(http.StatusOK, "OK")
}

func loginUserHandler(c echo.Context) error {
	u := &db.User{}
	if err := c.Bind(u); err != nil {
		return err
	}
	user, err := db.ConfirmUser(config.Connection(), u)
	if err != nil {
		log.Println("Hash does not match")
		return c.JSON(http.StatusForbidden, "Login credentials are not correct")
	}
	return c.JSON(http.StatusOK, user)
}

func logoutUserHandler(c echo.Context) error {
	return nil
}

func getBookmarkedProjectsHandler(c echo.Context) error {
	userId, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return err
	}
	projects := db.FetchUserBookmarkedProjects(config.Connection(), userId)
	if projects != nil {
		return c.JSON(http.StatusOK, projects)
	}
	return c.JSON(http.StatusOK, projects)
}

func addBookmarkedProjectHandler(c echo.Context) error {
	p := &db.Project{}
	userId, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return err
	}
	if err := c.Bind(p); err != nil {
		return err
	}
	projectId, err := db.ProjectExists(config.Connection(), p.Name, p.Author, p.Language)
	if err != nil {
		newProjectId, err := db.AddProject(config.Connection(), p)
		if err != nil {
			return err
		}
		_, err = db.BookmarkProject(config.Connection(), userId, newProjectId)
		if err != nil {
			log.Println("Problems bookmarking project 1")
			return c.JSON(http.StatusBadRequest, "Failed to bookmark project -> 1")
		}
		return c.JSON(http.StatusOK, "Bookmarked project 1")
	}
	_, err = db.BookmarkProject(config.Connection(), userId, projectId)
	if err != nil {
		log.Println("Problems bookmarking project 2")
		return c.JSON(http.StatusBadRequest, "Failed to bookmark project -> 2")
	}
	return c.JSON(http.StatusOK, "bookmarked project 2")
}

func main() {
	config = db.NewConfig(host, port, userPg, password, dbname)
	e := echo.New()
	e.Use(middleware.CORS())
	e.POST("/users", addNewUserHandler)
	e.POST("/users/login", loginUserHandler)
	e.POST("/users/logout", logoutUserHandler)
	e.GET("/users/:id/bookmarked_projects", getBookmarkedProjectsHandler)
	e.POST("/users/:id/bookmarked_projects", addBookmarkedProjectHandler)
	e.Logger.Fatal(e.Start(":1323"))
}
