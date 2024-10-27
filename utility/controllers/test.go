package controllers

import (
	"encoding/json"
	"log"
	"net/http"
	"reakgo/utility"
)

type TestStruct1 struct {
	Email string `json:"email"`
	Name  string `json:"name"`
}

func Add_into_admin(w http.ResponseWriter, r *http.Request) bool {
	response := utility.AjaxResponse{StatusCode: "500", Message: "Please login in the website to continue"}
	var loginCredentials TestStruct1
	err := json.NewDecoder(r.Body).Decode(&loginCredentials)
	if err != nil {
		log.Println(err)
		response.Message = "Unable to login as server has received invalid data.  " + utility.ContactAdministratorMessage(true)
		utility.RenderTemplate(w, r, "", response)
		return true
	}
	err = Db.authStruct.Add_into_admin(loginCredentials.Name, loginCredentials.Email)
	if err != nil {
		log.Println("error in admin controllers:", err)
		utility.RenderTemplate(w, r, "", response)
		return true
	}
	response.StatusCode = "200"
	response.Message = "successfully added"
	utility.RenderTemplate(w, r, "", response)
	return false
}

func Add_into_user(w http.ResponseWriter, r *http.Request) bool {
	response := utility.AjaxResponse{StatusCode: "500", Message: "Please login in the website to continue"}
	var loginCredentials TestStruct1
	err := json.NewDecoder(r.Body).Decode(&loginCredentials)
	if err != nil {
		log.Println(err)
		response.Message = "Unable to login as server has received invalid data.  " + utility.ContactAdministratorMessage(true)
		utility.RenderTemplate(w, r, "", response)
		return true
	}
	err = Db.authStruct.Add_into_user(loginCredentials.Name, loginCredentials.Email)
	if err != nil {
		log.Println("error in user controllers:", err)
		utility.RenderTemplate(w, r, "", response)
		return true
	}
	response.StatusCode = "200"
	response.Message = "successfully added"
	utility.RenderTemplate(w, r, "", response)
	return false
}

func Get_all_users(w http.ResponseWriter, r *http.Request) bool {
	response := utility.AjaxResponse{StatusCode: "500", Message: "Please login in the website to continue"}

	// users, err := Db.authStruct.GetAllUsers()
	users, err := Db.authStruct.GetUserByEmail()
	if err != nil {
		log.Println("error in user controllers:", err)
		utility.RenderTemplate(w, r, "", response)
		return true
	}
	response.StatusCode = "200"
	response.Message = "successfully fetch"
	response.Payload = users
	utility.RenderTemplate(w, r, "", response)
	return false
}
