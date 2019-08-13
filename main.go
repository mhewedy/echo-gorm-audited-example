package main

import (
	"github.com/dgrijalva/jwt-go"
	"github.com/labstack/echo/middleware"
	"github.com/qor/audited"
	"net/http"
	"time"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
	"github.com/labstack/echo"
)

const (
	getTokenURL = "/get-token"
	gromDB      = "gorm.db"
)

type User struct {
	gorm.Model
}

type Product struct {
	gorm.Model
	audited.AuditedModel
	code string
}

func main() {
	e := echo.New()
	db := initDB()
	defer db.Close()

	// JWT middleware
	e.Use(middleware.JWTWithConfig(middleware.JWTConfig{
		Skipper: func(c echo.Context) bool {
			if c.Request().URL.String() == getTokenURL {
				return true
			}
			return false
		},
		SigningKey: []byte("secret"),
	}))

	e.Use(gormJWTInjector(db))

	// apis
	e.POST("/create-product", func(c echo.Context) error {
		db := c.Get(gromDB).(*gorm.DB)
		db.Create(&Product{code: "my-prod-code"})
		return c.String(http.StatusOK, "Hello, World!")
	})

	e.GET(getTokenURL, func(c echo.Context) error {
		token, _ := createToken(101)
		return c.JSON(http.StatusOK, map[string]string{"message": token})
	})
	e.Logger.Fatal(e.Start(":1323"))
}

func initDB() *gorm.DB {
	db, err := gorm.Open("sqlite3", "test.db")
	if err != nil {
		panic(err)
	}
	audited.RegisterCallbacks(db)
	db.AutoMigrate(&User{}, &Product{})
	return db
}

func createToken(id int) (string, error) {
	token := jwt.New(jwt.SigningMethodHS256)
	claims := token.Claims.(jwt.MapClaims)
	claims["id"] = id
	claims["exp"] = time.Now().Add(time.Hour * 24 * 10).Unix()
	t, err := token.SignedString([]byte("secret"))
	return t, err
}

func gormJWTInjector(db *gorm.DB) func(next echo.HandlerFunc) echo.HandlerFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {

			if userToken := c.Get("user"); userToken != nil {
				claims := userToken.(*jwt.Token).Claims.(jwt.MapClaims)
				user := User{
					gorm.Model{
						ID: uint(claims["id"].(float64)),
					},
				}
				db = db.Set("audited:current_user", user)
				c.Set(gromDB, db)
			}
			return next(c)
		}
	}
}
