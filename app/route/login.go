package route

import (
	"encoding/base64"
	"fmt"
	"net/http"
	"os"
	"strconv"

	"github.com/matryer/way"
	"golang.org/x/exp/slog"

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
	success := c.Sess.CSRF(r)
	if !success {
		return http.StatusBadRequest, nil
	}

	username := r.FormValue("username")
	password := r.FormValue("password")
	mfa := r.FormValue("mfa")
	remember := r.FormValue("remember")

	allowedUsername := os.Getenv("PBB_USERNAME")
	if len(allowedUsername) == 0 {
		slog.ErrorCtx(r.Context(), "Environment variable missing: PBB_USERNAME")
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}

	hash := os.Getenv("PBB_PASSWORD_HASH")
	if len(hash) == 0 {
		slog.ErrorCtx(r.Context(), "Environment variable missing: PBB_PASSWORD_HASH")
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}

	// Get the MFA key - if the environment variable doesn't exist, then
	// let the MFA pass.
	mfakey := os.Getenv("PBB_MFA_KEY")
	mfaSuccess := true
	if len(mfakey) > 0 {
		imfa := 0
		imfa, err = strconv.Atoi(mfa)
		if err != nil {
			http.Redirect(w, r, "/", http.StatusFound)
			return
		}

		mfaSuccess, err = totp.Authenticate(imfa, mfakey)
		if err != nil {
			return http.StatusInternalServerError, err
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
		fmt.Printf("Login attempt failed. Username: %v (expected: %v) | Password match: %v | MFA success: %v\n", username, allowedUsername, passMatch, mfaSuccess)
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}

	fmt.Printf("Login attempt successful.\n")

	c.Sess.SetUser(r, username)
	if remember == "on" {
		c.Sess.RememberMe(r, true)
	} else {
		c.Sess.RememberMe(r, false)
	}

	http.Redirect(w, r, "/dashboard", http.StatusFound)
	return
}

func (c *AuthUtil) logout(w http.ResponseWriter, r *http.Request) (status int, err error) {
	c.Sess.Logout(r)

	http.Redirect(w, r, "/", http.StatusFound)
	return
}
