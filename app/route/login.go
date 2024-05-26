package route

import (
	"encoding/base64"
	"log/slog"
	"net/http"
	"os"
	"strconv"

	"github.com/matryer/way"

	"go.yhsif.com/pandablog/app/lib/envdetect"
	"go.yhsif.com/pandablog/app/lib/passhash"
	"go.yhsif.com/pandablog/app/lib/totp"
)

// AuthUtil -
type AuthUtil struct {
	*Core
}

func registerAuthUtil(c *AuthUtil) {
	c.Router.Get("/login/:slug", c.login)
	c.Router.Post("/login/:slug", c.loginPost)
	c.Router.Get("/dashboard/logout", c.logout)
}

// login allows a user to login to the dashboard.
func (c *AuthUtil) login(w http.ResponseWriter, r *http.Request) (status int, err error) {
	site, err := c.Storage.Site.Load(r.Context())
	if err != nil {
		return http.StatusInternalServerError, err
	}

	slug := way.Param(r.Context(), "slug")
	if slug != site.LoginURL {
		return http.StatusNotFound, nil
	}

	vars := make(map[string]any)
	vars["title"] = "Login"
	vars["token"] = c.Sess.SetCSRF(r)

	return c.Render.Template(w, r, "base", "login", vars)
}

func (c *AuthUtil) loginPost(w http.ResponseWriter, r *http.Request) (status int, err error) {
	site, err := c.Storage.Site.Load(r.Context())
	if err != nil {
		return http.StatusInternalServerError, err
	}

	slug := way.Param(r.Context(), "slug")
	if slug != site.LoginURL {
		return http.StatusNotFound, nil
	}

	r.ParseForm()

	// CSRF protection.
	if !c.Sess.CSRF(r) {
		slog.ErrorContext(r.Context(), "Login attempt failed.", slog.Group(
			"login",
			"csrfPassed", false,
		))
		return http.StatusBadRequest, nil
	}

	username := r.FormValue("username")
	password := r.FormValue("password")
	mfa := r.FormValue("mfa")
	remember := r.FormValue("remember") == "on"

	allowedUsername := os.Getenv("PBB_USERNAME")
	if len(allowedUsername) == 0 {
		slog.ErrorContext(r.Context(), "Environment variable missing: PBB_USERNAME")
		http.Redirect(w, r, "/", http.StatusFound)
		return http.StatusFound, nil
	}

	hash := os.Getenv("PBB_PASSWORD_HASH")
	if len(hash) == 0 {
		slog.ErrorContext(r.Context(), "Environment variable missing: PBB_PASSWORD_HASH")
		http.Redirect(w, r, "/", http.StatusFound)
		return http.StatusFound, nil
	}

	// Get the MFA key - if the environment variable doesn't exist, then
	// let the MFA pass.
	mfakey := os.Getenv("PBB_MFA_KEY")
	mfaSuccess := true
	if len(mfakey) > 0 {
		imfa := 0
		imfa, err = strconv.Atoi(mfa)
		if err != nil {
			mfaSuccess = false
		}

		mfaSuccess, err = totp.Authenticate(imfa, mfakey)
		if err != nil {
			mfaSuccess = false
		}
	}

	// When running locally, let any MFA pass.
	if envdetect.RunningLocalDev() {
		mfaSuccess = true
	}

	// Decode the hash - this is to allow it to be stored easily since dollar
	// signs are difficult to work with.
	hashDecoded, err := base64.StdEncoding.DecodeString(hash)
	if err != nil {
		return http.StatusInternalServerError, err
	}
	passMatch := passhash.MatchString(string(hashDecoded), password)

	// If the username and password don't match, then just redirect.
	if username != allowedUsername || !passMatch || !mfaSuccess {
		slog.ErrorContext(r.Context(), "Login attempt failed.", slog.Group(
			"login",
			"method", "password",
			slog.Group(
				"matched",
				"password", passMatch,
				"mfa", mfaSuccess,
				slog.Group(
					"username",
					"got", username,
					"want", allowedUsername,
				),
			),
		))
		http.Redirect(w, r, "/", http.StatusFound)
		return http.StatusFound, nil
	}

	slog.WarnContext(r.Context(), "Login attempt successful.", slog.Group(
		"login",
		"method", "password",
		"mfa", len(mfakey) > 0,
		"remember", remember,
	))

	c.Sess.SetUser(r, username)
	c.Sess.RememberMe(r, remember)

	http.Redirect(w, r, "/dashboard", http.StatusFound)
	return http.StatusFound, nil
}

func (c *AuthUtil) logout(w http.ResponseWriter, r *http.Request) (status int, err error) {
	c.Sess.Logout(r)

	http.Redirect(w, r, "/", http.StatusFound)
	return http.StatusFound, nil
}
