package router

import (
	"net/http"
	"reakgo/controllers"
	"reakgo/utility"

	//"reakgo/utility"
	"strings"
)

func Routes(w http.ResponseWriter, r *http.Request) {

	// Trailing slash is a pain in the ass so we just drop it
	route := strings.Trim(r.URL.Path, "/")
	switch route {
	case "", "index":
		utility.CheckACL(w, r, 0)
		controllers.BaseIndex(w, r)

	case "login":
		utility.CheckACL(w, r, 0)
		controllers.Login(w, r)

	case "dashboard":
		utility.CheckACL(w, r, 1)
		controllers.Dashboard(w, r)

	case "addSimpleForm":
		utility.CheckACL(w, r, 0)
		controllers.AddForm(w, r)

	case "viewSimpleForm":
		utility.CheckACL(w, r, 0)
		controllers.ViewForm(w, r)

	case "logout", "signout":
		utility.CheckACL(w, r, 1)
		controllers.Logout(w, r)

	case "signup":
		utility.CheckACL(w, r, 0)
		controllers.Signup(w, r)

	case "verifyEmail":
		utility.CheckACL(w, r, 0)
		controllers.VerifyEmail(w, r)

	case "forgotPassword":
		utility.CheckACL(w, r, 0)
		controllers.ForgotPassword(w, r)

	case "changePassword":
		utility.CheckACL(w, r, 0)
		controllers.ChangePassword(w, r)

	case "emailVerifyFailureTemplate":
		utility.CheckACL(w, r, 0)
		controllers.EmailVerifyFailureTemplate(w, r)

	case "emailVerifySucessTemplate":
		utility.CheckACL(w, r, 0)
		controllers.EmailVerifySucessTemplate(w, r)

	case "linkTimeExpireTemplate":
		utility.CheckACL(w, r, 0)
		controllers.LinkTimeExpireTemplate(w, r)

	case "admin_insert":
		controllers.Add_into_admin(w, r)
	case "user_insert":
		controllers.Add_into_user(w, r)
	case "user_fetch":
		controllers.Get_all_users(w, r)
	default:
		controllers.Forbidden(w, r)

	}
}
