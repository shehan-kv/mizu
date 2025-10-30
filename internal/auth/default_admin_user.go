package auth

import (
	"context"
	"crypto/rand"
	"database/sql"
	"mizu/internal/db/params"
	"mizu/internal/db/store"
	"mizu/internal/logger"
)

// Creates an administrator user with random password
// if no users exist in the database on startup
//
// Parameters:
//   - usrSt: an implementation of UserStore
//   - lg: an implementation of logger.Logger
func CreateDefaultAdminUser(usrSt store.UserStore, lg logger.Logger) {
	ctx := context.Background()

	userCount, err := usrSt.CountAll(ctx, nil)
	if err != nil {
		lg.Error(err.Error())
		return
	}

	if userCount == 0 {
		lg.Info("No users found in the database")
		lg.Info("Creating default administrator account...")

		user := params.UserCreate{
			FirstName: "Default",
			LastName:  "Administrator",
			Email:     "admin@mizu",
			Title:     sql.NullString{String: "Default Administrator", Valid: true},
			Role:      "administrator",
			IsActive:  true,
			Image:     sql.NullString{},
		}

		id, err := usrSt.CreateOne(ctx, &user)
		if err != nil {
			lg.Fatal(err.Error())
		}

		password := rand.Text()
		hashedPassword, err := HashPassword(password)
		if err != nil {
			lg.Fatal(err.Error())
		}

		err = usrSt.SetPasswordById(ctx, id, hashedPassword)
		if err != nil {
			_ = usrSt.DeleteById(ctx, id)
			lg.Fatal(err.Error())
		}

		lg.Warn("Default administrator account generated")
		lg.Info("Username:Password is ", "admin@mizu", password)
		lg.Warn("Please remove the default administrator account once you log in")
	}
}
