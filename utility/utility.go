package utility

import (
	"bytes"
	"crypto/rand"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"math/big"
	"net/http"
	"net/smtp"
	"os"
	"runtime"
	"strconv"
	"strings"

	"github.com/gorilla/sessions"
	"github.com/jmoiron/sqlx"
	"golang.org/x/crypto/bcrypt"
)

// Template Pool
var View *template.Template

// Session Store
var Store *sessions.FilesystemStore

// DB Connections
var Db *sqlx.DB
var Db1 *sqlx.DB
type Session struct {
	Key   string
	Value interface{}
}

type Flash struct {
	Type    string
	Message string
}

// type AjaxResponse struct {
// 	Status  string
// 	Message string
// 	Payload interface{}
// }

func RedirectTo(w http.ResponseWriter, r *http.Request, path string) {
	http.Redirect(w, r, os.Getenv("APP_URL")+"/"+path, http.StatusFound)
}

func SessionSet(w http.ResponseWriter, r *http.Request, data Session) {
	session, _ := Store.Get(r, os.Getenv("SESSION_NAME"))
	// Set some session values.
	session.Values[data.Key] = data.Value
	// Save it before we write to the response/return from the handler.
	err := session.Save(r, w)
	if err != nil {
		Logger(err, true)
	}
}

func SessionGet(r *http.Request, key string) interface{} {
	session, _ := Store.Get(r, os.Getenv("SESSION_NAME"))
	// Set some session values.
	return session.Values[key]
}

func CheckACL(w http.ResponseWriter, r *http.Request, minLevel int) bool {
	userType := SessionGet(r, "type")
	var level int = 0
	switch userType {
	case "user":
		level = 1
	case "admin":
		level = 2
	default:
		level = 0
	}
	if level >= minLevel {
		return true
	} else {
		response := AjaxResponse{StatusCode: "401", Message: "Unauthorized access", Payload: ""}
		var template string
		if IsCurlApiRequest(r) {
			RenderTemplate(w, r, template, response)
			return false
		} else {
			//template = "forbidden"
			RedirectTo(w, r, "forbidden")
			return false
		}
	}
}

func AddFlash(flavour string, message string, w http.ResponseWriter, r *http.Request) {
	session, err := Store.Get(r, os.Getenv("SESSION_NAME"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	//flash := make(map[string]string)
	//flash["Flavour"] = flavour
	//flash["Message"] = message
	flash := Flash{
		Type:    flavour,
		Message: message,
	}

	session.AddFlash(flash, "message")
	err = session.Save(r, w)
	if err != nil {
		Logger(err, false)
	}
}

func viewFlash(w http.ResponseWriter, r *http.Request) interface{} {
	session, err := Store.Get(r, os.Getenv("SESSION_NAME"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	fm := session.Flashes("message")
	if fm == nil {
		return nil
	}
	session.Save(r, w)
	return fm
}

func RenderTemplate(w http.ResponseWriter, r *http.Request, template string, data interface{}) {
	session, _ := Store.Get(r, os.Getenv("SESSION_NAME"))
	tmplData := make(map[string]interface{})
	if template == "" {
		jsonresponce, err := json.Marshal(data)
		if err != nil {
			Logger(err, true)
		}
		w.Write([]byte(jsonresponce))
	} else {
		ajaxReq, ok := data.(AjaxResponse)
		if ok {
			AddFlash(ajaxReq.StatusCode, ajaxReq.Message, w, r)
		}
		tmplData["data"] = data
		tmplData["flash"] = viewFlash(w, r)
		tmplData["session"] = session.Values["email"]
		tmplData["appUrl"] = os.Getenv("APP_URL")
		tmplData["supportEmail"] = os.Getenv("SUPPORT_EMAIL")
		tmplData["flashDuration"] = os.Getenv("FLASH_DURATION")
		tmplData["forbiddenPage"] = os.Getenv("FORBIDDEN_PAGE")
		tmplData["forbbidenCode"] = os.Getenv("FORBIDDEN_CODE")

		View.ExecuteTemplate(w, template, tmplData)
	}

}

func GetErrorMessage(currentFilePath string, lineNumbers int, errorMessage error, isCritical bool) bool {
	LOG_FILE := "Errorlogged.txt"
	// open log file
	logFile, err := os.OpenFile(LOG_FILE, os.O_APPEND|os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		log.Println(err)
		return true
	}
	defer logFile.Close()
	// Set log out put and enjoy :)
	log.SetOutput(logFile)

	// optional: log date-time, filename, and line number
	log.SetFlags(log.Lshortfile | log.LstdFlags)
	log.Println("Error message :- ", currentFilePath, lineNumbers, errorMessage)
	if isCritical {
		emailForErrorMessageSend := os.Getenv("EMAIL_FOR_CRITICAL_ERROR")
		email := []string{emailForErrorMessageSend} // set email address
		data := make(map[string]interface{})
		data["subject"] = "Error message file"
		data["errorMessage"] = errorMessage
		data["currentFilePath"] = currentFilePath
		data["lineNumbers"] = lineNumbers
		if !SendEmail(email, "errorMessage", data) {
			log.Println("Email for error message couldn't be sent at the moment, Please try again")
			return true
		}
	}
	return false
}

// logger fo critical errors
// provide currentPath,lineNumbers for err
func Logger(errObject error, isCritical bool) {

	//using 1 indicate actually error
	_, currentFilePath, lineNumbers, ok := runtime.Caller(1)
	if !ok {
		err := errors.New("failed to get filename")
		log.Println(err)
	}
	if os.Getenv("CI_ENVIRONMENT") == "development" { // when development environment is set email not to be sent to developer because of this rise a error
		//calling error function
		log.Println(currentFilePath, lineNumbers, errObject)
	} else {
		go GetErrorMessage(currentFilePath, lineNumbers, errObject, isCritical)
	}
}

/*
remove files from the specified directory
ex- RemoveFile("assets/images/market/testing3375083430.png")
*/
func RemoveFile(filePath string) {
	err := os.Remove(filePath)
	if err != nil {
		Logger(err, false)
	}
}

// get sql error string from sql error
func GetSqlErrorString(err error) string {
	mes := strings.SplitN(err.Error(), ":", -1)
	return mes[1]
}

// match sql error with particular sql error string
func CheckSqlError(err error, errString string) (bool, string) {
	sqlerrorString := GetSqlErrorString(err)
	exists := strings.HasPrefix(sqlerrorString, errString)
	return exists, sqlerrorString
}

/* smtp send email*/
func SendEmailSMTP(to []string, subject string, body string) bool {
	//Sender data.
	from := os.Getenv("FROM_EMAIL")
	// Set up email information.
	header := "MIME-version: 1.0;\nContent-Type: text/html; charset=\"UTF-8\";\n"
	msg := []byte("From: " + from + "\n" + "To: " + strings.Join(to, ",") + "\n" + "Subject: " + subject + "\r\n" + header + body)
	// Sending email.
	// fmt.Println("From: " + from + "\n" + "To: " + strings.Join(to, ",") + "\n" + "Subject: " + subject + "\r\n" + header + "\r\n" + body)
	err := smtp.SendMail(os.Getenv("SMTP_HOST")+":"+os.Getenv("SMTP_PORT"), smtp.PlainAuth("", os.Getenv("FROM_APIKEY"), os.Getenv("EMAIL_SECRATE"), os.Getenv("SMTP_HOST")), from, to, msg)
	if err != nil {
		Logger(err, true)
		return false
	}
	return true
}

/*
* main function to call for send mail
#input

	to : string array
	template : tempatePath
	data : associative array of data which is set on template, by-default app_url and app_name is set
*/
func SendEmail(to []string, template string, data map[string]interface{}) bool {
	buf := new(bytes.Buffer)
	//extra information on email
	data["app_url"] = os.Getenv("APPURL")
	data["app_name"] = os.Getenv("APPNAME")
	// Set up email information.
	err := View.ExecuteTemplate(buf, template, data)
	if err != nil {
		Logger(err, true)
		return false
	}
	return SendEmailSMTP(to, fmt.Sprintf("%v", data["subject"]), buf.String())
}

// upload image file
// upload image file
func UploadFile(r *http.Request, fileName string, controlName string, avatar string) (string, error, string) {
	var message string
	err := r.ParseMultipartForm(10 << 20)
	if err != nil {
		Logger(err, false)
	}
	file, _, err := r.FormFile(controlName)
	if err != nil {
		Logger(err, false)
	} else {
		// files
		defer file.Close()
		tempFile, err := ioutil.TempFile("assets/images/"+avatar, fileName+"*.png")
		if err != nil {
			Logger(err, false)
		}

		defer tempFile.Close()
		fileBytes, err := ioutil.ReadAll(file)
		if err != nil {
			Logger(err, false)
		}
		tempFile.Write(fileBytes)
		fileExtention := http.DetectContentType(fileBytes)

		fileInfo, err := os.Stat(tempFile.Name())
		if err != nil {
			Logger(err, false)
		}
		//check file size
		fileSize := fileInfo.Size()

		allowedFileSize, err := StrToInt64(os.Getenv("UPLOAD_FILE_SIZE_IN_KB"))

		if err != nil {
			Logger(err, false)
			allowedFileSize = 50 //IN KB
		}

		if allowedFileSize == 0 {
			err := errors.New("either upload_file_size in kb env variable not found or found null")
			Logger(err, false)
			allowedFileSize = 50 //IN KB
		}

		//upload file only till 5 mb and fileExtension is jpeg/jpg/png
		if fileSize <= (allowedFileSize*1000) && fileExtention == "image/jpeg" || fileExtention == "image/jpg" || fileExtention == "image/png" {
			return tempFile.Name(), err, message

		} else {
			// to-do calcualate image size and apply remove file
			message = "Image size should not be greater than  " + fmt.Sprint(allowedFileSize) + "kb and accepted formats are jpeg,jpg,png"
		}
	}
	return "", err, message
}

// convert string to int64
func StrToInt64(str string) (int64, error) {
	if str == "" {
		return 0, nil
	}
	strint64, err := strconv.ParseInt(str, 10, 64)
	if err != nil {
		Logger(err, false)
		return int64(0), err
	}
	return strint64, err
}

func SessionDestroy(w http.ResponseWriter, r *http.Request) bool {
	session, err := Store.Get(r, os.Getenv("SESSION_NAME"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return false
	}
	session.Options.MaxAge = -1
	err = session.Save(r, w)
	if err != nil {
		Logger(err, false)
		return false
	}
	// cookie, err := r.Cookie(os.Getenv("SESSION_NAME"))
	// if err != nil {
	//     // Cookie doesn't exist, nothing to delete
	//     return
	// }
	// cookie.Expires = time.Now().AddDate(0, 0, -1) // set expiration time to a day in the past
	// http.SetCookie(w, cookie)
	return true
}

func ContactAdministratorMessage(willRetry bool) string {
	SUPPORT_EMAIL := os.Getenv("SUPPORT_EMAIL")
	if SUPPORT_EMAIL == "" {
		SUPPORT_EMAIL = "support@reak.in"
		fmt.Println("SUPPORT_EMAIL not found in env")
	}

	// link to project administrator email
	emailLink := `<a class="badge" style="background-color: gray;" href=mailto:` + SUPPORT_EMAIL + `">` + SUPPORT_EMAIL + `</a>`

	message := `Please contact the administrator at at ` + emailLink
	if willRetry {
		message = `Please retry and if the problem persists reach out to the administrator at` + emailLink
	}
	return message
}

func GenerateRandomString(n int) (string, error) {
	const letters = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz-"
	ret := make([]byte, n)
	for i := 0; i < n; i++ {
		num, err := rand.Int(rand.Reader, big.NewInt(int64(len(letters))))
		if err != nil {
			Logger(err, false)
			return "", err
		}
		ret[i] = letters[num.Int64()]
	}

	return string(ret), nil
}

func GenerateSHA(key string) string {
	sum := sha256.Sum256([]byte(key))
	sumString := fmt.Sprintf("%x", sum)
	return sumString
}

// Generate NewPasswordHash
func NewPasswordHash(NewPassword string) (bool, string) {
	//NewPassword Change bcrypt code
	newPasswordHash, err := bcrypt.GenerateFromPassword([]byte(NewPassword), 10)
	if err != nil {
		Logger(err, true)
		return false, ""
	}
	//modify NewPassword
	NewPassword = string(newPasswordHash)
	if NewPassword == "" {
		return false, NewPassword
	}
	return true, NewPassword
}

// convert string to int
func StrToInt(num string) int {
	if num != "" {
		intNum, err := strconv.Atoi(num)
		if err != nil {
			Logger(err, false)
		}
		return intNum
	}
	return 0
}

type AjaxResponse struct {
	StatusCode string
	Message    string
	Payload    interface{}
}

func StrToFloat64(floatValue string) float64 {
	if floatValue == "" {
		return 0
	}
	StrFloat, err := strconv.ParseFloat(floatValue, 64)
	if err != nil {
		Logger(err, false)
		return 0
	}
	return StrFloat
}

func IsCurlApiRequest(r *http.Request) bool {
	return r.Header.Get("reak-api") == "true"
}
