package controllers

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"reakgo/models"
	"reakgo/utility"

	"github.com/gorilla/sessions"
	"golang.org/x/crypto/bcrypt"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

type LoginAuth struct {
	Email      string `json:"email"`
	Password   string `json:"password"`
	RememberMe string `json:"rememberMe"`
	Token      string `json:"token"`
}

func Login(w http.ResponseWriter, r *http.Request) bool {
	response := utility.AjaxResponse{StatusCode: "500", Message: "Please login in the website to continue"}
	var loginCredentials LoginAuth
	var template string

	if r.Method == "POST" {

		if utility.IsCurlApiRequest(r) {
			// json decode and the save in struct
			err := json.NewDecoder(r.Body).Decode(&loginCredentials)
			if err != nil {
				log.Println(err)
				response.Message = "Unable to login as server has received invalid data.  " + utility.ContactAdministratorMessage(true)
				utility.RenderTemplate(w, r, template, response)
				return true
			}

		} else {
			template = "login"
			// Check for any form parsing error
			err := r.ParseForm()
			if err != nil {
				utility.Logger(err, true)
				response.Message = "Unable to login as server has received invalid data" + utility.ContactAdministratorMessage(true)
				utility.RenderTemplate(w, r, template, response)
				return true
			}

			// Parsing form went fine, Now we can access all the values
			loginCredentials.Email = r.FormValue("email")
			loginCredentials.Password = r.FormValue("password")
			loginCredentials.RememberMe = r.FormValue("rememberMe")
		}

		if loginCredentials.Email == "" || loginCredentials.Password == "" {
			response.Message = "Please fill all required fields and try again"
			utility.RenderTemplate(w, r, template, response)
			return true
		}

		row, err := Db.authentication.GetUserByColumn("email", loginCredentials.Email)
		log.Println(row)
		log.Println(err)
		if err != nil {
			exists, errString := utility.CheckSqlError(err, " no rows in result set")
			response.Message = errString
			if exists {
				response.Message = "Incorrect credentials, please re-check and try again"
			}
			utility.RenderTemplate(w, r, template, response)
			return true
		}

		match := bcrypt.CompareHashAndPassword([]byte(row.Password), []byte(loginCredentials.Password))
		if match != nil {
			response.Message = "Credentials didn't match, Please try again."
			utility.RenderTemplate(w, r, template, response)
			return true
		}

		//checking user Active(verified email) or not
		if !row.IsActive {
			response.Message = "Please verify your email to continue"
			utility.RenderTemplate(w, r, template, response)
			return true
		}

		// checking reason: skip redirection from api request
		if utility.IsCurlApiRequest(r) {
			response.StatusCode = "200"
			response.Payload = row
			response.Message = "Login Success!"
			utility.RenderTemplate(w, r, template, response)
			return false

		} else {
			// Password match has been a success
			utility.SessionSet(w, r, utility.Session{Key: "id", Value: row.Id})
			utility.SessionSet(w, r, utility.Session{Key: "email", Value: row.Email})
			utility.SessionSet(w, r, utility.Session{Key: "type", Value: row.Type})

			if loginCredentials.RememberMe == "on" {
				MaxSessionTimeInMin := os.Getenv("MAX_SESSION_TIME_IN_MIN")
				sessionTime := utility.StrToInt(MaxSessionTimeInMin) //sessionTime is in min
				if sessionTime == 0 {
					fmt.Println("MAX_SESSION_TIME_IN_MIN fetched from env is not valid-", MaxSessionTimeInMin)
					sessionTime = 15
				}
				utility.Store.Options = &sessions.Options{
					MaxAge: (3600 * sessionTime), // Max age is in sec
				}

			} else {
				DefaultSessionTimeInMin := os.Getenv("DEFAULT_SESSION_TIME_IN_MIN")
				sessionTime := utility.StrToInt(DefaultSessionTimeInMin)
				if sessionTime == 0 {
					fmt.Println("DEFAULT_SESSION_TIME_IN_MIN fetched from env is not valid-", DefaultSessionTimeInMin)
					sessionTime = 5
				}
				utility.Store.Options = &sessions.Options{
					MaxAge: (3600 * sessionTime), // Max age is in sec
				}
			}

			utility.RedirectTo(w, r, "dashboard")
			return false
		}
	}
	utility.RenderTemplate(w, r, "login", response)
	return false
}

func Signup(w http.ResponseWriter, r *http.Request) bool {
	response := utility.AjaxResponse{StatusCode: "500"}
	var template string
	if r.Method == "POST" || r.Method == "PUT" {
		var data models.Authentication

		if utility.IsCurlApiRequest(r) {
			// json decode and the save in struct
			err := json.NewDecoder(r.Body).Decode(&data)
			if err != nil {
				utility.Logger(err, true)
				response.Message = "Unable to signup as server has received invalid data. " + utility.ContactAdministratorMessage(true)
				utility.RenderTemplate(w, r, template, response)
				return true
			}

		} else {
			template = "signup"
			err := r.ParseForm()
			if err != nil {
				utility.Logger(err, true)
				response.Message = "Unable to signup as server has received invalid data. " + utility.ContactAdministratorMessage(true)
				utility.RenderTemplate(w, r, template, response)
				return true
			}

			// Parsing form went fine, Now we can access all the values
			data.Email = r.FormValue("email")
			data.Password = r.FormValue("password")
			data.Type = r.FormValue("type")
			data.FirstName = r.FormValue("firstName")
			data.LastName = r.FormValue("lastName")
		}
		// generate a token which will be used for email verification
		generateRandomToken, err := utility.GenerateRandomString(60)
		// If generateRandomString function fails, we will create it from combination of email and password.
		if err != nil || generateRandomToken == "" {
			generateRandomToken = utility.GenerateSHA(data.Email + time.Now().String())
		}
		data.Token = generateRandomToken
		data.TokenTimestamp = time.Now().Unix()

		// Generating password hash from password to store securely in db
		encrypt, encryptedPassword := utility.NewPasswordHash(data.Password)
		if !encrypt {
			response.Message = "Unable to register due to failure of password encryption. " + utility.ContactAdministratorMessage(true)
			utility.RenderTemplate(w, r, template, response)
			return true
		}
		data.Password = encryptedPassword

		data.AuthToken = utility.GenerateSHA(data.Email + encryptedPassword)

		if data.AuthToken == "" {
			// will stop the exceution in this case as Auth_Token is important column
			response.Message = "We were unable to register as a result of technical problems. " + utility.ContactAdministratorMessage(true)
			utility.RenderTemplate(w, r, "", response)
			return true
		}

		// Using transaction so that can rollback db operation if the email sends fails
		tx := utility.Db.MustBegin()
		// Inserting data to database
		err = Db.authentication.Signup(data, tx)
		if err != nil {
			tx.Rollback()
			response.Message = utility.GetSqlErrorString(err)
			utility.RenderTemplate(w, r, template, response)
			return true
		}

		/* Send Email section starts
		   will fetch required variables from env for email */
		emailTokenExpire, err := (strconv.ParseInt(os.Getenv("EMAIL_TOKEN_EXPIRE"), 10, 64))
		var changeSecondsToMintue int64

		if err != nil {
			fmt.Println(err)
			/* In this case only the email expire time will be missing in email so we don't need to stop the exceution.
			that's why hardcoded the minimum expire time should be */
			changeSecondsToMintue = 30
		} else {
			changeSecondsToMintue = emailTokenExpire / 60
		}
		appName := os.Getenv("APP_NAME")
		frontendURL := os.Getenv("FRONTEND_URL")
		if appName == "" {
			appName = "REAK"
			fmt.Println("APP_NAME not found in env")
		}
		if frontendURL == "" {
			frontendURL = "http://localhost:4000/"
			fmt.Println("FRONTEND_URL not found in env")
		}

		// map for email details
		email_data := make(map[string]interface{})
		userEmailId := []string{data.Email} // set email address
		email_data["subject"] = "Email verification mail "
		email_data["email"] = userEmailId
		email_data["link_to_verify_data"] = "verifyEmail?email=" + data.Email + "&token=" + generateRandomToken
		email_data["change_seconds_to_minute"] = changeSecondsToMintue
		email_data["name"] = cases.Title(language.Und).String(data.FirstName)
		email_data["App_name"] = appName
		email_data["App_url"] = frontendURL

		if !utility.SendEmail(userEmailId, "verifyEmailTemplate", email_data) {
			tx.Rollback()
			response.Message = "Unable to sign up as email could not be sent at this time, please verify your email and internet connection"
			utility.RenderTemplate(w, r, template, response)
			return true
		}

		/* send Email section ends */

		// will commit db operation as email sent successfully
		err = tx.Commit()
		if err != nil {
			log.Println(err)
			tx.Rollback()
			response.Message = "We were unable to register as a result of technical problems. " + utility.ContactAdministratorMessage(true)
			utility.RenderTemplate(w, r, template, response)
			return true
		}

		response.StatusCode = "200"
		response.Message = "Congratulations, you have successfully enrolled. Please verify your email"
		utility.RenderTemplate(w, r, template, response)
		return false
	}
	utility.RenderTemplate(w, r, "signup", response)
	return false
}

type ForgotEmailStruct struct {
	Email string
}

func ForgotPassword(w http.ResponseWriter, r *http.Request) bool {
	response := utility.AjaxResponse{StatusCode: "500"}

	if r.Method == "POST" {
		var credential ForgotEmailStruct
		var template string
		var successMessage string
		if utility.IsCurlApiRequest(r) {
			// json decode and save in struct
			err := json.NewDecoder(r.Body).Decode(&credential)

			if err != nil {
				response.Message = "Could not process your forgot password request because the server has received invalid data. " + utility.ContactAdministratorMessage(true)
				utility.RenderTemplate(w, r, template, response)
				return true
			}
			successMessage = "New password has been sent to you via email, Please check your inbox"

		} else {
			template = "forgotpassword"
			err := r.ParseForm()
			if err != nil {
				//critical error checking
				utility.Logger(err, true)
				response.Message = "Could not process your forgot password request because the server has received invalid data. " + utility.ContactAdministratorMessage(true)
				utility.RenderTemplate(w, r, template, response)
				return true
			}
			credential.Email = r.FormValue("email")
			successMessage = "Email with password reset link has sent has been sent, Please check your inbox"
		}

		if credential.Email == "" {
			response.Message = "Please provide email and try again"
			utility.RenderTemplate(w, r, template, response)
			return true
		}

		authData, err := Db.authentication.GetUserByColumn("email", credential.Email)

		if err != nil {

			exists, errString := utility.CheckSqlError(err, " no rows in result set")
			message := errString
			if exists {
				utility.Logger(errors.New(credential.Email+"requesting forgot password request but not registered"), false)
				time.Sleep(2 * time.Second)
				message = successMessage
			}
			response.Message = message
			utility.RenderTemplate(w, r, template, response)
			return true
		}

		// set email data and send email
		data := make(map[string]interface{})
		userEmailId := []string{credential.Email} // set email address
		data["subject"] = " Forgot Password "
		data["name"] = cases.Title(language.Und).String(authData.FirstName)
		data["APP_NAME"] = os.Getenv("APP_NAME")
		var emailTemplate string

		if utility.IsCurlApiRequest(r) {
			randonStringSize, err := strconv.Atoi(os.Getenv("PASSWORD_STRING_SIZE"))
			if err != nil {
				log.Println(err)
				randonStringSize = 8
			}

			generateRandomToken, err := utility.GenerateRandomString(randonStringSize)
			if err != nil {
				log.Println(err)
				generateRandomToken = utility.GenerateSHA(credential.Email + time.Now().String())
			}

			encrypt, encryptedPassword := utility.NewPasswordHash(generateRandomToken)

			if !encrypt {
				response.Message = "Unable to process your missed password request due to a failed password encryption " + utility.ContactAdministratorMessage(true)
				utility.RenderTemplate(w, r, "", response)
				return true
			}

			//transaction start
			tx := utility.Db.MustBegin()
			err = Db.authentication.UpdatePasswordUser(tx, credential.Email, encryptedPassword)
			if err != nil {
				tx.Rollback()
				response.Message = utility.GetSqlErrorString(err)
				utility.RenderTemplate(w, r, "", response)
				return true
			}
			data["password"] = generateRandomToken
			emailTemplate = "newPassword"

		} else {
			// User returned successfully, Send email
			tokenNew, err := Db.authentication.ForgotPassword(authData.Id)
			if err != nil {
				response.Message = utility.GetSqlErrorString(err)
				utility.RenderTemplate(w, r, template, response)
				return true
			}
			data["email"] = credential.Email
			data["App_url"] = os.Getenv("FRONTEND_URL")
			data["link_to_verify_data"] = "changePassword?email=" + credential.Email + "&token=" + tokenNew
			emailTemplate = "emailResetPassword"
		}

		if !utility.SendEmail(userEmailId, emailTemplate, data) {
			response.Message = "The email could not be sent at this moment, please verify your email and internet connection. " + utility.ContactAdministratorMessage(true)
			utility.RenderTemplate(w, r, template, response)
			return true
		}

		response.StatusCode = "200"
		response.Message = successMessage
		utility.RenderTemplate(w, r, template, response)
		return false
	}
	utility.RenderTemplate(w, r, "forgotpassword", response)
	return false
}

// this function open by email link ,when new password created.
func ChangePassword(w http.ResponseWriter, r *http.Request) bool {

	if r.Method == "POST" {
		err := r.ParseForm()
		if err != nil {
			//critical error checking
			utility.Logger(err, true)
			utility.AddFlash("error", "Could not process your change password request because the server has received invalid data."+utility.ContactAdministratorMessage(true), w, r)
			utility.RenderTemplate(w, r, "changePassword", nil)
			return true
		}
		token := r.URL.Query().Get("token")
		newPassword := r.FormValue("newpassword")
		confirmpassword := r.FormValue("confirmpassword")
		//if token equal to empty string show failure message and show failure message(email) for user
		if token == "" {
			utility.AddFlash("error", "Could not process your change password request because the server has received invalid data."+utility.ContactAdministratorMessage(true), w, r)
			utility.RenderTemplate(w, r, "changePassword", nil)
			return true
		}
		//newPassword and confirmPassword equal to empty string open changepassword template
		if newPassword == "" && confirmpassword == "" {
			utility.AddFlash("error", "Please fill all required field and retry", w, r)
			utility.RenderTemplate(w, r, "changePassword", nil)
			return true
		}
		////newPassword and confirmPassword do not match show error.
		if newPassword != confirmpassword {
			utility.AddFlash("error", "Password and confirm password does not match", w, r)
			utility.RenderTemplate(w, r, "changePassword", nil)
			return true
		}
		//token verify
		boolvalue, err := Db.authentication.TokenVerify(token, newPassword)
		if err != nil {
			//critical error checking
			exists, errString := utility.CheckSqlError(err, " no rows in result set")
			message := errString
			if exists {
				message = "Authentication Failure, please click on valid Email"
			}
			utility.AddFlash("error", message, w, r)
			utility.RenderTemplate(w, r, "changePassword", nil)
			return true
		}
		//if token verify fails
		if boolvalue {
			utility.AddFlash("error", "The reset Password link was expired. Please request the email again.", w, r)
			utility.RenderTemplate(w, r, "changePassword", nil)
		}
		utility.AddFlash("error", "Password reset successfully", w, r)
		utility.RenderTemplate(w, r, "changePassword", nil)
		return false
	}
	utility.RenderTemplate(w, r, "changePassword", nil)
	return false
}

func VerifyEmail(w http.ResponseWriter, r *http.Request) bool {
	response := utility.AjaxResponse{StatusCode: "500"}
	var template string
	var data LoginAuth

	if utility.IsCurlApiRequest(r) {
		err := json.NewDecoder(r.Body).Decode(&data)
		if err != nil {
			utility.Logger(err, true)
			response.Message = "Unable to verify the email as server has received invalid data. " + utility.ContactAdministratorMessage(true)
			utility.RenderTemplate(w, r, template, response)
			return true
		}

	} else {
		template = "emailVerifyFailureTemplate"
		data.Token = r.URL.Query().Get("token")
		data.Email = r.URL.Query().Get("email")
	}

	/* backend Validation of required fields*/
	if data.Token == "" || data.Email == "" {
		utility.Logger(errors.New("Either received empty token- "+data.Token+"or email- "+data.Token), false)
		response.Message = "Unable to verify the email as it lacks the required values. " + utility.ContactAdministratorMessage(true)
		utility.RenderTemplate(w, r, template, response)
		return true
	}

	verifyStatus := Db.authentication.VerifyEmail(data.Token, data.Email)

	switch verifyStatus {

	case "success":
		response.Message = "Email Verification was successful, Please continue by logging in"
		template = "emailVerifySucessTemplate"
		break

	case "linkexpire":
		response.Message = "Email verification link was expired. " + utility.ContactAdministratorMessage(false)
		template = "linkTimeExpireTemplate"
		break

	case "accessDenied":
		response.Message = "Authentication Failure, please click on valid Email"
		template = "emailVerifyFailureTemplate"
		break

	case "active":
		response.Message = "Please login to continue using the application"
		template = "emailVerifySucessTemplate"
		break

	default:
		response.Message = verifyStatus
		template = "emailVerifyFailureTemplate"
		break
	}

	if utility.IsCurlApiRequest(r) {
		template = ""
	}
	utility.RenderTemplate(w, r, template, response)
	return false
}

func EmailVerifyFailureTemplate(w http.ResponseWriter, r *http.Request) {
	utility.RenderTemplate(w, r, "emailVerifyFailureTemplate", nil)
}

func EmailVerifySucessTemplate(w http.ResponseWriter, r *http.Request) {
	utility.RenderTemplate(w, r, "emailVerifySucessTemplate", nil)
}

func LinkTimeExpireTemplate(w http.ResponseWriter, r *http.Request) {
	utility.RenderTemplate(w, r, "linkTimeExpireTemplate", nil)
}

func Forbidden(w http.ResponseWriter, r *http.Request) {
	response := utility.AjaxResponse{StatusCode: "404", Message: "Request Not Found", Payload: ""}
	var template string
	if !utility.IsCurlApiRequest(r) {
		template = "login"
	}
	utility.RenderTemplate(w, r, template, response)
}

func Logout(w http.ResponseWriter, r *http.Request) bool {
	response := utility.AjaxResponse{StatusCode: "500", Message: "", Payload: []interface{}{}}
	res := utility.SessionDestroy(w, r)
	if res {
		response.StatusCode = "200"
		utility.RenderTemplate(w, r, "", response)
		return false
	}
	response.StatusCode = "500"
	utility.RenderTemplate(w, r, "", response)
	return true
}
