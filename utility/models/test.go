package models

import (
	"log"
	"reakgo/utility"

	"github.com/jmoiron/sqlx"
)

type TestStruct2 struct {
	Id    int64
	Name  string
	Email string
}

type AuthStruct struct {
	DB *sqlx.DB
}

func (AuthStruct) Add_into_admin(name string, email string) error {
	_, err := utility.Db1.NamedExec("INSERT INTO admin_auth (name,email) VALUES ( :Name,:Email)", map[string]interface{}{"Name": name, "Email": email})

	if err != nil {
		log.Println("err in admin model:", err)
	}
	return err
}
func (AuthStruct) Add_into_user(name string, email string) error {
	_, err := utility.Db.NamedExec("INSERT INTO users (name,email) VALUES ( :Name,:Email)", map[string]interface{}{"Name": name, "Email": email})

	if err != nil {
		log.Println("err in user model:", err)
	}
	return err
}

func (AuthStruct) GetAllUsers() ([]TestStruct2, error) {
	var users []TestStruct2
	err := utility.Db.Select(&users, "SELECT id, name, email FROM users")
	if err != nil {
		log.Println("err in getting users:", err)
		return nil, err
	}
	return users, nil
}
func (AuthStruct) GetUserByEmail() (TestStruct2, error) {
	var selectedRow TestStruct2
	err := utility.Db.Get(&selectedRow, "SELECT `id`, `name`, `email` FROM users WHERE email = ?", "prmsyne@gmail.com")
	if err != nil {
		log.Println("err in getting user:", err)

	}
	return selectedRow, err
}
