package controllers

import (
	"reakgo/models"
	"reakgo/utility"

	"github.com/jmoiron/sqlx"
)

type Env struct {
	authentication interface {
		ForgotPassword(id int) (string, error)
		TokenVerify(token string, newPassword string) (bool, error)
		Signup(add models.Authentication, tx *sqlx.Tx) error
		GetUserByColumn(column string, value string) (models.Authentication, error)
		VerifyEmail(token string, email string) string
		UpdatePasswordUser(tx *sqlx.Tx, email string, password string) error
	}

	formAddView interface {
		Add(name string, address string)
		View() ([]models.FormAddView, error)
	}
	authStruct interface {
		Add_into_admin(name string, email string) error
		Add_into_user(name string, email string) error
		GetAllUsers() ([]models.TestStruct2, error)
		GetUserByEmail() (models.TestStruct2, error)
	}
}

var Db *Env

func init() {
	Db = &Env{
		authentication: models.AuthenticationModel{DB: utility.Db},
		formAddView:    models.FormAddViewModel{DB: utility.Db},
		authStruct:     models.AuthStruct{DB: utility.Db},
	}
}
