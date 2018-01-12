package actions

import (
	"fmt"
	"net/http"

	"github.com/derhabicht/rmuse/models"
	"github.com/gobuffalo/buffalo"
	"github.com/markbates/pop"
	"golang.org/x/crypto/bcrypt"
)

// UserCreate default implementation.
func UserCreate(c buffalo.Context) error {
	type argument struct {
		FirstName string `json:"firstname"`
		LastName  string `json:"lastname"`
		Email     string `json:"email"`
		Username  string `json:"username"`
		Artist    bool   `json:"artist"`
		Password  string `json:"password"`
	}

	arg := &argument{}
	if err := c.Bind(arg); err != nil {
		return c.Render(http.StatusUnprocessableEntity, r.JSON("{\"error\":\"malformed argument body\"}"))
	}

	ph, err := bcrypt.GenerateFromPassword([]byte(arg.Password), bcrypt.DefaultCost)
	if err != nil {
		return c.Render(http.StatusInternalServerError, r.JSON("{\"error\":\"cannot hash password\"}"))
	}

	u := &models.User{
		FirstName:    arg.FirstName,
		LastName:     arg.LastName,
		Email:        arg.Email,
		Username:     arg.Username,
		Artist:       arg.Artist,
		PasswordHash: string(ph),
	}

	tx := c.Value("tx").(*pop.Connection)
	verrs, err := u.Create(tx)
	if err != nil {
		// TODO: Double check validations here to see why they fail
		return c.Render(http.StatusInternalServerError, r.JSON("{\"error\":\"failed to create user\"}"))
	}

	if verrs.HasAny() {
		return c.Render(http.StatusUnprocessableEntity, r.JSON(verrs))
	}

	ts, err := u.CreateJWTToken()
	if err != nil {
		return c.Render(http.StatusInternalServerError, r.JSON("{\"error\":\"failed to create token\"}"))
	}

	res := struct {
		Token string       `json:"token"`
		User  *models.User `json:"user"`
	}{
		Token: ts,
		User:  u,
	}

	return c.Render(http.StatusOK, r.JSON(res))
}

func UserUpdate(c buffalo.Context) error {
	cu, ok := c.Value("user").(*models.User)

	if !ok {
		return c.Render(http.StatusUnauthorized, r.JSON("{\"error\":\"not authorized to update user\""))
	}

	type argument struct {
		FirstName string `json:"firstname"`
		LastName  string `json:"lastname"`
		Email     string `json:"email"`
		Username  string `json:"username"`
		Artist    bool   `json:"artist"`
		Password  string `json:"password"`
	}

	arg := &argument{}
	if err := c.Bind(arg); err != nil {
		return c.Render(http.StatusUnprocessableEntity, r.JSON("{\"error\":\"malformed argument body\"}"))
	}

	ph, err := bcrypt.GenerateFromPassword([]byte(arg.Password), bcrypt.DefaultCost)
	if err != nil {
		return c.Render(http.StatusInternalServerError, r.JSON("{\"error\":\"cannot hash password\"}"))
	}

	u := &models.User{
		FirstName:    arg.FirstName,
		LastName:     arg.LastName,
		Email:        arg.Email,
		Username:     arg.Username,
		Artist:       arg.Artist,
		PasswordHash: string(ph),
	}

	if cu.FirstName != u.FirstName {
		c.Logger().Debug(cu.FirstName)
		c.Logger().Debug(u.FirstName)
		cu.FirstName = u.FirstName
	}
	if cu.LastName != u.LastName {
		cu.LastName = u.LastName
	}
	if cu.Email != u.Email {
		cu.Email = u.Email
	}
	if cu.Username != u.Username {
		cu.Username = u.Username
	}
	if cu.Artist != u.Artist {
		cu.Artist = u.Artist
	}
	if cu.PasswordHash != u.PasswordHash {
		cu.PasswordHash = u.PasswordHash
	}

	tx := c.Value("tx").(*pop.Connection)
	verrs, err := cu.Update(tx)
	if err != nil {
		return c.Render(http.StatusInternalServerError, r.JSON("{\"error\":\"failed to create user\"}"))
	}

	if verrs.HasAny() {
		return c.Render(http.StatusUnprocessableEntity, r.JSON(verrs))
	}

	return c.Render(http.StatusOK, r.JSON(cu))
}

func UserRead(c buffalo.Context) error {
	return c.Render(http.StatusOK, r.JSON(c.Value("user")))
}

func UserPageFetch(c buffalo.Context) error {
	username := c.Param("username")
	tx := c.Value("tx").(*pop.Connection)

	m, err := models.GetMediaByUsername(tx, username)

	if err != nil {
		resp := struct{
			Errors []string `json:"errors"`
		}{
			Errors: []string{fmt.Sprintf("user %s not found", username)},
		}
		return c.Render(http.StatusNotFound, r.JSON(resp))
	}

	u, ok := c.Value("user").(*models.User)

	if !ok {
		u = nil
	}

	res := struct{
		Following bool          `json:"following"`
		Media     *models.Media `json:"images"`
	}{
		Following: u.Follows(tx, username),
		Media: m,
	}

	return c.Render(http.StatusOK, r.JSON(res))
}

func UserFollow(c buffalo.Context) error {
	tx := c.Value("tx").(*pop.Connection)
	u, ok := c.Value("user").(*models.User)

	if !ok {
		return c.Render(http.StatusUnauthorized, r.JSON("{\"error\":\"must be logged in to follow\"}"))
	}

	username := c.Param("username")
	fu, err := models.GetUserByUsername(tx, username)

	if err != nil {
		emsg := struct{
			Error string `json:"error"`
		}{
			Error: fmt.Sprintf("user %s does not exist", username),
		}
		return c.Render(http.StatusUnprocessableEntity, r.JSON(emsg))
	}

	f := &models.Follow{
		Follower: u.ID,
		Followed: fu.ID,
	}

	_, err = f.Create(tx)

	if err != nil {
		emsg := struct{
			Error string `json:"error"`
		}{
			Error: fmt.Sprintf("unable to follow user %s", username),
		}
		return c.Render(http.StatusInternalServerError, r.JSON(emsg))
	}

	return c.Render(http.StatusOK, r.JSON(""))
}

func UserUnfollow(c buffalo.Context) error {
	tx := c.Value("tx").(*pop.Connection)
	u, ok := c.Value("user").(*models.User)

	if !ok {
		return c.Render(http.StatusUnauthorized, r.JSON("{\"error\":\"must be logged in to unfollow\"}"))
	}

	username := c.Param("username")
	fu, err := models.GetUserByUsername(tx, username)

	if err != nil {
		return c.Render(http.StatusOK, r.JSON(""))
	}

	f := &models.Follow{
		Follower: u.ID,
		Followed: fu.ID,
	}

	f.Delete(tx)

	return c.Render(http.StatusOK, r.JSON(""))
}
