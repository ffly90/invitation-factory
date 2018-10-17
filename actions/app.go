package actions

import (
	"net/http"

	"github.com/gobuffalo/buffalo"
	"github.com/gobuffalo/buffalo-pop/pop/popmw"
	"github.com/gobuffalo/envy"
	"github.com/gobuffalo/mw-forcessl"
	paramlogger "github.com/gobuffalo/mw-paramlogger"
	"github.com/gorilla/sessions"
	"github.com/unrolled/secure"

	"github.com/gobuffalo/mw-csrf"
	"github.com/gobuffalo/mw-i18n"
	"github.com/gobuffalo/packr"
	"github.com/invitation/models"
)

// ENV is used to help switch settings based on where the
// application is being run. Default is "development".
var ENV = envy.Get("GO_ENV", "development")
var app *buffalo.App
var store *sessions.CookieStore

// T is the translator
var T *i18n.Translator

// App is where all routes and middleware for buffalo
// should be defined. This is the nerve center of your
// application.
func App() *buffalo.App {
	if app == nil {
		var hashKey = []byte("wjdpqjwdwqnbdpqwjdpqwupoqwjdqwbdoibqwiodjpoqwudpqwze98123e9z1wpdjpoqdnj1bediu1dh")
		store = sessions.NewCookieStore(hashKey)
		store.Options = &sessions.Options{
			HttpOnly: true,
			MaxAge:   86400 * 7,
			Secure:   true,
			SameSite: http.SameSiteLaxMode,
			Path:     "/",
		}
		app = buffalo.New(buffalo.Options{
			Env:          ENV,
			SessionName:  "__Secure-__Host-_invitation_session",
			SessionStore: store,
		})

		// Automatically redirect to SSL
		app.Use(forceSSL())

		if ENV == "development" {
			app.Use(paramlogger.ParameterLogger)
		}

		// Protect against CSRF attacks. https://www.owasp.org/index.php/Cross-Site_Request_Forgery_(CSRF)
		// Remove to disable this.
		app.Use(csrf.New)

		// Wraps each request in a transaction.
		//  c.Value("tx").(*pop.PopTransaction)
		// Remove to disable this.
		app.Use(popmw.Transaction(models.DB))

		// Setup and use translations:
		app.Use(translations())

		app.Use(SRIHandler)

		app.Use(SetSecurityHeaders)

		app.GET("/", HomeHandler)

		app.Use(SetCurrentUser)
		app.Use(Authorize)
		app.GET("/users/new", UsersNew)
		app.POST("/users", UsersCreate)
		app.GET("/signin", AuthNew)
		app.POST("/signin", AuthCreate)
		app.DELETE("/signout", AuthDestroy)
		app.GET("/invitations/{invitation_id}/send", InvitMailSend)
		app.GET("/invitations/{invitation_id}/guests/{guest_id}", StatusResponse)
		app.POST("/invitations/{invitation_id}/guests/{guest_id}", SetStatusResponse)
		app.GET("/invitations/delete_guest/{guest_id}", DeleteGuestFromUnsubscribe)
		app.GET("/users/{user_id}/verify", VerifyUser)
		app.Middleware.Skip(Authorize, HomeHandler, UsersNew, UsersCreate, AuthNew, AuthCreate, DeleteGuestFromUnsubscribe, VerifyUser, StatusResponse, SetStatusResponse)
		app.Resource("/invitations", InvitationsResource{})
		app.ServeFiles("/", assetsBox) // serve files from the public directory
	}

	return app
}

// translations will load locale files, set up the translator `actions.T`,
// and will return a middleware to use to load the correct locale for each
// request.
// for more information: https://gobuffalo.io/en/docs/localization
func translations() buffalo.MiddlewareFunc {
	var err error
	if T, err = i18n.New(packr.NewBox("../locales"), "en-US"); err != nil {
		app.Stop(err)
	}
	return T.Middleware()
}

// forceSSL will return a middleware that will redirect an incoming request
// if it is not HTTPS. "http://example.com" => "https://example.com".
// This middleware does **not** enable SSL. for your application. To do that
// we recommend using a proxy: https://gobuffalo.io/en/docs/proxy
// for more information: https://github.com/unrolled/secure/
func forceSSL() buffalo.MiddlewareFunc {
	return forcessl.Middleware(secure.Options{
		SSLRedirect:     ENV == "production",
		SSLProxyHeaders: map[string]string{"X-Forwarded-Proto": "https", "Strict-Transport-Security": "max-age=31536000; includeSubDomains; preload"},
	})
}
