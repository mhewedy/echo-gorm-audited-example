```go
func GormJWTInjector(db *gorm.DB) func(next echo.HandlerFunc) echo.HandlerFunc {
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
```
