package models

import (
	"crypto/rand"
	"math/big"
	"os"
	"reakgo/utility"
	"strconv"
	"time"

	"github.com/jmoiron/sqlx"
	"golang.org/x/crypto/bcrypt"
)

type Authentication struct {
	Id             int    `db:"id" json:"id"`
	FirstName      string `db:"first_name" json:"firstName"`
	LastName       string `db:"last_name" json:"lastName"`
	Email          string `db:"email" json:"email"`
	Password       string `db:"password" json:"password"`
	TokenTimestamp int64  `db:"token_timestamp" json:"tokenTimestamp"`
	Token          string `db:"token" json:"token"`
	Type           string `db:"type" json:"type"`
	IsActive       bool   `db:"is_active" json:"isActive"`
	AuthToken      string `db:"auth_token" json:"authToken"`
}

type TwoFactor struct {
	UserId int32  `db:"userId"`
	Secret string `db:"secret"`
}

type AuthenticationModel struct {
	DB *sqlx.DB
}

func (auth AuthenticationModel) GetUserByColumn(column string, value string) (Authentication, error) {
	var selectedRow Authentication
	err := utility.Db.Get(&selectedRow, "SELECT `id`, `first_name`, `last_name`, `email`, `password`, `token`, `token_timestamp`, `is_active`, `type`, `auth_token` FROM authentication WHERE "+column+" = ?", value)
	if err != nil {
		utility.Logger(err, true)
	}
	return selectedRow, err
}

func (auth AuthenticationModel) ForgotPassword(id int) (string, error) {
	Token, err := GenerateRandomString(60)
	if err != nil {
		utility.Logger(err, true)
	}
	TokenTimestamp := time.Now().Unix()
	query, err := utility.Db.Prepare("UPDATE authentication SET token = ?, token_timestamp = ? WHERE id = ?")
	if err != nil {
		utility.Logger(err, true)
	}
	_, err = query.Exec(Token, TokenTimestamp, id)
	if err != nil {
		utility.Logger(err, true)
	}
	return Token, err
}

func GenerateRandomString(n int) (string, error) {
	const letters = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz-"
	ret := make([]byte, n)
	for i := 0; i < n; i++ {
		num, err := rand.Int(rand.Reader, big.NewInt(int64(len(letters))))
		if err != nil {
			return "", err
		}
		ret[i] = letters[num.Int64()]
	}

	return string(ret), nil
}

func (auth AuthenticationModel) TokenVerify(token string, newPassword string) (bool, error) {
	selectedRow, err := auth.GetUserByColumn("token", token)
	if err != nil {
		utility.Logger(err, true)
		return true, err
	}
	if (selectedRow.TokenTimestamp + 360000) > time.Now().Unix() {
		_, err := auth.ChangePassword(newPassword, selectedRow.Id)
		if err != nil {
			utility.Logger(err, true)
			return true, err
		}
		return false, err
	}
	return true, err
}

func (auth AuthenticationModel) ChangePassword(newPassword string, id int) (bool, error) {
	query, err := utility.Db.Prepare("UPDATE authentication SET password = ? WHERE id = ?")
	if err != nil {
		utility.Logger(err, true)
	}
	newPasswordHash, err := bcrypt.GenerateFromPassword([]byte(newPassword), 10)
	if err != nil {
		utility.Logger(err, true)
		return true, err
	}
	_, err = query.Exec(newPasswordHash, id)
	if err != nil {
		utility.Logger(err, true)
		return true, err
	} else {
		return false, err
	}
}

func (auth AuthenticationModel) Signup(add Authentication, tx *sqlx.Tx) error {
	_, err := tx.NamedExec("INSERT INTO authentication (first_name,last_name,email,password,token,token_timestamp, auth_token) VALUES ( :FirstName,:LastName,:Email,:Password,:Token,:TokenTimestamp,:AuthToken)", map[string]interface{}{"FirstName": add.FirstName, "LastName": add.LastName, "Email": add.Email, "Password": add.Password, "Token": add.Token, "TokenTimestamp": add.TokenTimestamp, "AuthToken": add.AuthToken})

	if err != nil {
		utility.Logger(err, true)
	}
	return err
}

func (auth AuthenticationModel) VerifyUser(token string, email string) (bool, Authentication) {
	var selectedRow Authentication
	emailTokenExpire, err := strconv.ParseInt(os.Getenv("EMAIL_TOKEN_EXPIRE"), 10, 64)
	if err != nil {
		utility.Logger(err, false)
		// hard coded here the minimum token expire time should be as the execution should not stop if not got this value from env
		emailTokenExpire = 1800
	}
	rows := utility.Db.QueryRow("SELECT id, is_active, token_timestamp FROM authentication WHERE `email` = ? AND `token` = ?", email, token)
	err = rows.Scan(&selectedRow.Id, &selectedRow.IsActive, &selectedRow.TokenTimestamp)
	if err != nil {
		utility.Logger(err, true)
		return false, selectedRow
	}
	current_time := time.Now().Unix()
	if current_time <= selectedRow.TokenTimestamp+emailTokenExpire {
		return true, selectedRow
	}

	return false, selectedRow
}

func (auth AuthenticationModel) VerifyEmail(token string, email string) string {
	boolean, selectedRow := auth.VerifyUser(token, email)
	if boolean {
		if !selectedRow.IsActive {
			data := make(map[string]interface{})
			data["Id"] = selectedRow.Id
			data["isActive"] = true
			// update is_active true if token is verify
			_, err := utility.Db.NamedExec("UPDATE authentication SET is_active=:isActive WHERE id=:id ", map[string]interface{}{"isActive": data["isActive"], "id": data["Id"]})

			if err != nil {
				utility.Logger(err, true)
				return utility.GetSqlErrorString(err)
			}
			return "success"
		}
		return "active"
	}
	if selectedRow.Id == 0 {
		return "accessDenied"
	} else {
		return "linkexpire"
	}
}

// update user password
func (auth AuthenticationModel) UpdatePasswordUser(tx *sqlx.Tx, email string, password string) error {
	_, err := tx.NamedExec("UPDATE authentication SET password=:Password WHERE email=:Email ", map[string]interface{}{"Password": password, "Email": email})
	if err != nil {
		utility.Logger(err, true)
		return err

	}
	return err
}
