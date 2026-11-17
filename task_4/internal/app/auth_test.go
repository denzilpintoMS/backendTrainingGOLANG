package app_test

import (
	"context"
	"net/http"
	"strings"
	application "task_3/internal/app"
	"testing"

	"golang.org/x/crypto/bcrypt"
	"task_3/internal/model"
	"task_3/internal/repository"
)

type fakeUserRepository struct {
	user model.User
	err  error
}

func (r *fakeUserRepository) CreateUser(_ context.Context, user model.User) (model.User, error) {
	if r.err != nil {
		return model.User{}, r.err
	}
	if r.user.Email == user.Email {
		return model.User{}, repository.ErrEmailInUse
	}
	user.ID = 1
	r.user = user
	return user, nil
}
func (r *fakeUserRepository) GetUserByEmail(_ context.Context, email string) (model.User, error) {
	if r.err != nil {
		return model.User{}, r.err
	}
	if email != r.user.Email {
		return model.User{}, repository.ErrUserNotFound
	}
	return r.user, nil
}

func TestRegisterAndLogin(t *testing.T) {
	users := &fakeUserRepository{}
	app := application.New(&fakeBookRepository{}, users, testTokens())
	var registered map[string]any
	request(t, app, http.MethodPost, "/api/v1/auth/register", `{"name":" Alice ","email":" ALICE@example.com ","password":" secret123 "}`, 201, &registered)
	if users.user.Email != "alice@example.com" || users.user.Name != "Alice" {
		t.Fatalf("unexpected user: %+v", users.user)
	}
	if err := bcrypt.CompareHashAndPassword([]byte(users.user.PasswordHash), []byte(" secret123 ")); err != nil {
		t.Fatal(err)
	}
	public := registered["user"].(map[string]any)
	if len(public) != 5 || public["password"] != nil || public["PasswordHash"] != nil || public["password_hash"] != nil {
		t.Fatalf("unsafe user response: %v", public)
	}
	request(t, app, http.MethodPost, "/api/v1/auth/register", `{"name":"Alice","email":"ALICE@example.com","password":"secret123"}`, 409, nil)
	var login struct {
		AccessToken string `json:"access_token"`
		TokenType   string `json:"token_type"`
		ExpiresIn   int    `json:"expires_in"`
	}
	request(t, app, http.MethodPost, "/api/v1/auth/login", `{"email":"ALICE@example.com","password":" secret123 "}`, 200, &login)
	id, err := testTokens().Verify(login.AccessToken)
	if err != nil || id != 1 || login.TokenType != "Bearer" || login.ExpiresIn != 3600 {
		t.Fatalf("invalid login token response: %+v, %v", login, err)
	}
	request(t, app, http.MethodPost, "/api/v1/books", `{"name":"Authorized book"}`, 201, nil, "Bearer "+login.AccessToken)
	for _, body := range []string{
		`{"email":"alice@example.com","password":"wrong"}`,
		`{"email":"missing@example.com","password":" secret123 "}`,
		`{"email":"alice@example.com","password":"secret123"}`,
	} {
		var response map[string]string
		request(t, app, http.MethodPost, "/api/v1/auth/login", body, 401, &response)
		if response["error"] != "invalid email or password" {
			t.Fatalf("unexpected response: %v", response)
		}
	}
}

func TestInvalidRegistration(t *testing.T) {
	for _, body := range []string{
		`{}`, `null`, `{`, `[]`, `{"name":"A","email":"invalid","password":"secret123"}`,
		`{"name":" ","email":"a@example.com","password":"secret123"}`,
		`{"name":"A","email":"a@example.com","password":"short"}`,
		`{"name":"A","email":"a@example.com","password":"` + strings.Repeat("a", 73) + `"}`,
		`{"name":"A","email":"a@example.com","password":"secret123","role":"admin"}`,
		`{"name":"A","email":"a@example.com","password":"secret123"} {}`,
	} {
		users := &fakeUserRepository{}
		request(t, application.New(&fakeBookRepository{}, users, testTokens()), http.MethodPost, "/api/v1/auth/register", body, 400, nil)
		if users.user.ID != 0 {
			t.Fatal("invalid registration wrote a user")
		}
	}
}

func TestAuthRepositoryErrors(t *testing.T) {
	users := &fakeUserRepository{err: context.DeadlineExceeded}
	app := application.New(&fakeBookRepository{}, users, testTokens())
	request(t, app, http.MethodPost, "/api/v1/auth/register", `{"name":"A","email":"a@example.com","password":"secret123"}`, 500, nil)
	request(t, app, http.MethodPost, "/api/v1/auth/login", `{"email":"a@example.com","password":"secret123"}`, 500, nil)
}
