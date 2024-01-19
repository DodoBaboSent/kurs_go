package main

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"text/template"
	"time"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const userkey = "user"
const rolekey = "role"

var secret = []byte("secret")

type User struct {
	gorm.Model
	Email    string `gorm:"unique"`
	Password string
	Role     string `gorm:"default:user"`
}

type News struct {
	gorm.Model  `gorm:"embedded"`
	Name        string
	Text        string
	UsrComments []Comments
}
type Comments struct {
	gorm.Model
	UserID uint
	NewsID uint
	User   User
	Text   string
}

var db *gorm.DB
var err error

func main() {
	r := gin.Default()
	r.SetFuncMap(template.FuncMap{
		"formatAsDate": formatAsDate,
		"getUsr":       getUsr,
	})
	r.StaticFile("/assets/app.css", "build/app.css")
	r.StaticFile("/assets/app.js", "build/app.js")
	r.StaticFile("/assets/main.wasm", "build/vendor/main.wasm")
	r.StaticFile("/service_js.js", "build/service_js.js")
	r.Static("/static", "build/static")
	r.Static("/templates", "src/templates")
	r.LoadHTMLFiles("src/templates/index.html", "src/templates/admin.html", "src/templates/new.html", "src/templates/article.html")

	db, err = gorm.Open(sqlite.Open("test.sqlite"), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}

	if err = db.AutoMigrate(&User{}); err == nil && db.Migrator().HasTable(&User{}) {
		if err := db.First(&User{}).Error; errors.Is(err, gorm.ErrRecordNotFound) {
			db.Delete(&User{Email: "admin@example.com"})
			db.Create(&User{Email: "admin@example.com", Password: "admin", Role: "admin"})
		}
	}

	if err = db.AutoMigrate(&News{}); err == nil && db.Migrator().HasTable(&News{}) {
		if err := db.First(&News{}).Error; errors.Is(err, gorm.ErrRecordNotFound) {
			comment := []Comments{Comments{
				User: User{Email: "test@example.com", Password: "test"},
				Text: "Test test",
			}}
			db.AutoMigrate(&Comments{})
			db.Create(&comment).Association("User")
			db.Delete(&News{Name: "Initial Stuff"})
			db.Create(&News{Name: "Initial Stuff", Text: "Lorem Ipsum stuff stuff stuff stuff stuff", UsrComments: comment}).Association("Comments")
			db.Create(&News{Name: "Initial Stuff 2", Text: "Lorem Ipsum stuff stuff stuff stuff stuff 2"})
		}
	}

	r.Use(sessions.Sessions("mysession", cookie.NewStore(secret)))

	r.GET("/", func(ctx *gin.Context) {
		ctx.HTML(http.StatusOK, "index.html", gin.H{})
	})

	r.POST("/login", login)
	r.GET("/logout", logout)
	r.POST("/reg", reg)
	r.GET("/news", func(ctx *gin.Context) {
		var newsScan []News
		var newsGot []*News
		db.Find(&newsGot).Scan(&newsScan)
		ctx.HTML(http.StatusOK, "new.html", gin.H{
			"News": newsScan,
		})
	})
	r.POST("/new-post", func(ctx *gin.Context) {
		name := ctx.PostForm("name")
		text := ctx.PostForm("text")

		post := News{Name: name, Text: text}
		result := db.Create(&post)
		if result.Error != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to reg"})
			return
		}
		ctx.Redirect(302, "/")
	})
	r.GET("/new/:id", func(ctx *gin.Context) {
		id_q := ctx.Param("id")
		id, _ := strconv.Atoi(id_q)
		NewsArticle := News{}
		db.Model(&News{}).Preload(clause.Associations).Find(&NewsArticle, "id = ?", id)
		session := sessions.Default(ctx)
		user := session.Get(userkey)
		ctx.HTML(http.StatusOK, "article.html", gin.H{
			"Article": NewsArticle,
			"CurUser": user,
		})
	})
	r.POST("/post-comment", func(ctx *gin.Context) {
		a_id := ctx.PostForm("id")
		a_id_uint, _ := strconv.Atoi(a_id)
		poster_usr := ctx.PostForm("name")
		text_comm := ctx.PostForm("text")
		var usr User
		db.Find(&usr, "email = ?", poster_usr)
		new_comm := Comments{UserID: usr.ID, NewsID: uint(a_id_uint), Text: text_comm}
		db.Create(&new_comm)
		ctx.Redirect(302, fmt.Sprintf("/new/%s", a_id))
	})

	private := r.Group("/admin")
	private.Use(AuthRequired)
	{
		private.GET("/me", me)
		private.GET("/status", status)
		private.GET("/panel", func(ctx *gin.Context) {
			session := sessions.Default(ctx)
			role_s := session.Get(rolekey)
			println(role_s)
			ctx.HTML(http.StatusOK, "admin.html", gin.H{
				"role": role_s,
			})
		})
	}

	r.Run("127.0.0.1:8080")
}

func formatAsDate(t time.Time) string {
	year, month, day := t.Date()
	return fmt.Sprintf("%02d/%02d/%d", day, month, year)
}

func getUsr(id uint) string {
	var usr User
	db.Find(&usr, "id = ?", id)
	return usr.Email
}

func AuthRequired(c *gin.Context) {
	session := sessions.Default(c)
	user := session.Get(userkey)
	if user == nil {
		// Abort the request with the appropriate error code
		c.Redirect(302, "/templates/login.html")
		return
	}
	// Continue down the chain to handler etc
	c.Next()
}

// login is a handler that parses a form and checks for specific data.
func login(c *gin.Context) {
	session := sessions.Default(c)
	username := c.PostForm("username")
	password := c.PostForm("password")

	// Validate form input
	if strings.Trim(username, " ") == "" || strings.Trim(password, " ") == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Parameters can't be empty"})
		return
	}

	// Check for username and password match, usually from a database
	var user = User{Email: username}
	var selected User
	db.First(&user).Scan(&selected)
	println(selected.Email)
	println(selected.Password)
	if selected.Email != username {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication failed"})
		return
	}
	if selected.Password != password {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication failed"})
		return
	}

	// Save the username in the session
	session.Set(userkey, username) // In real world usage you'd set this to the users ID
	session.Set(rolekey, selected.Role)
	if err := session.Save(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save session"})
		return
	}
	c.Redirect(302, "/")
}

// logout is the handler called for the user to log out.
func logout(c *gin.Context) {
	session := sessions.Default(c)
	user := session.Get(userkey)
	if user == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid session token"})
		return
	}
	session.Delete(userkey)
	if err := session.Save(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save session"})
		return
	}
	c.Redirect(302, "/admin/panel")
}

// me is the handler that will return the user information stored in the
// session.
func me(c *gin.Context) {
	session := sessions.Default(c)
	user := session.Get(userkey)
	c.JSON(http.StatusOK, gin.H{"user": user})
}

// status is the handler that will tell the user whether it is logged in or not.
func status(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "You are logged in"})
}

func reg(c *gin.Context) {
	session := sessions.Default(c)
	username := c.PostForm("username")
	password := c.PostForm("password")

	user := User{Email: username, Password: password}
	result := db.Create(&user)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to reg"})
		return
	}

	var selected User
	db.First(&user).Scan(&selected)
	session.Set(userkey, username)
	session.Set(rolekey, selected.Role)
	if err := session.Save(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save session"})
		return
	}
	c.Redirect(302, "/")
}
