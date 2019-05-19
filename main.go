package main

import (
	"log"
	"net/http"
	"os"
	"strconv"
	"teste/db"
	"teste/service"
	"teste/util"

	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
)

var (
	config *db.Config
)

func signUpHandler(c echo.Context) error {
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

	// generate JWT token and return on response
	return c.JSON(http.StatusOK, "OK")
}

func signinHandler(c echo.Context) error {
	u := &db.User{}
	if err := c.Bind(u); err != nil {
		return err
	}
	user, err := db.ConfirmUser(config.Connection(), u)
	if err != nil {
		log.Println("Hash does not match")
		return c.JSON(http.StatusForbidden, "Login credentials are not correct")
	}
	// generate JWT token and return on response
	return c.JSON(http.StatusOK, user)
}

func getBookmarkedProjectsHandler(c echo.Context) error {
	// verify jwt token
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
	// verify jwt token
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
	port, err := strconv.Atoi(os.Getenv("PORT"))
	if err != nil {
		log.Fatal(err)
	}
	config = db.NewConfig(
		os.Getenv("HOST"),
		port,
		os.Getenv("USER"),
		os.Getenv("PASS"),
		os.Getenv("DBNAME"))
	e := echo.New()
	e.Use(middleware.CORS())
	e.POST("/users", signUpHandler)
	e.POST("/users/login", signinHandler)
	e.GET("/users/:id/bookmarked_projects", getBookmarkedProjectsHandler)
	e.POST("/users/:id/bookmarked_projects", addBookmarkedProjectHandler)
	e.Logger.Fatal(e.Start(":1323"))
}
