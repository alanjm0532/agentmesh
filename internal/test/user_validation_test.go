package testutil_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/alanjm0532/agentmesh/internal/server/config"
	"github.com/alanjm0532/agentmesh/internal/server/entity"
	"github.com/alanjm0532/agentmesh/internal/server/handler"
	"github.com/alanjm0532/agentmesh/internal/server/infra"
	"github.com/alanjm0532/agentmesh/internal/server/repository"
	"github.com/alanjm0532/agentmesh/internal/server/service"
	"go.uber.org/zap"
)

func loginValidationConfig() config.Config {
	return config.Config{
		Server: config.ServerConfig{Mode: "test", SessionSecret: strings.Repeat("x", 32)},
		Database: config.DatabaseConfig{
			Driver: "sqlite",
			DSN:    ":memory:?_pragma=foreign_keys(1)",
		},
	}
}

func submitLogin(router http.Handler, body string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodPost, "/web/login", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()
	router.ServeHTTP(response, req)
	return response
}

func TestLoginRejectsInvalidParameters(t *testing.T) {
	// A nil service ensures invalid requests are rejected before business logic runs.
	router := infra.NewRouter(loginValidationConfig(), handler.NewHandler(nil), zap.NewNop())
	cases := []struct {
		name string
		body string
	}{
		{"empty body", ""},
		{"malformed JSON", "{"},
		{"null body", "null"},
		{"missing fields", "{}"},
		{"missing username", `{"password":"secret"}`},
		{"missing password", `{"username":"alice"}`},
		{"empty username", `{"username":"","password":"secret"}`},
		{"empty password", `{"username":"alice","password":""}`},
		{"blank username", `{"username":" \t\n\u3000","password":"secret"}`},
		{"null username", `{"username":null,"password":"secret"}`},
		{"null password", `{"username":"alice","password":null}`},
		{"wrong field type", `{"username":42,"password":"secret"}`},
		{"username over character limit", `{"username":"` + strings.Repeat("a", 256) + `","password":"secret"}`},
		{"Unicode username over character limit", `{"username":"` + strings.Repeat("界", 256) + `","password":"secret"}`},
		{"password over byte limit", `{"username":"alice","password":"` + strings.Repeat("a", 73) + `"}`},
		{"Unicode password over byte limit", `{"username":"alice","password":"` + strings.Repeat("密", 25) + `"}`},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			response := submitLogin(router, tt.body)
			if response.Code != http.StatusBadRequest {
				t.Fatalf("status = %d, want 400", response.Code)
			}
			if !bytes.Equal(bytes.TrimSpace(response.Body.Bytes()), []byte(`{"message":"Invalid login request"}`)) {
				t.Fatal("response should contain only the public validation message")
			}
			if len(response.Result().Cookies()) != 0 {
				t.Fatal("invalid request created a session")
			}
		})
	}
}

func TestLoginAcceptsBoundaryValues(t *testing.T) {
	cfg := loginValidationConfig()
	client, cleanup, err := infra.NewDatabase(cfg)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(cleanup)
	users := service.NewUserService(repository.NewUserRepository(client))
	router := infra.NewRouter(cfg, handler.NewHandler(users), zap.NewNop())
	cases := []struct {
		name     string
		username string
		password string
	}{
		{"minimum values", "a", "b"},
		{"ASCII limits", strings.Repeat("a", 255), strings.Repeat("b", 72)},
		{"Unicode limits", strings.Repeat("界", 255), strings.Repeat("密", 24)},
		{"password spaces preserved", "spaces", " secret "},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			body, err := json.Marshal(entity.LoginRequest{Username: tt.username, Password: tt.password})
			if err != nil {
				t.Fatal(err)
			}
			// Exercise both automatic registration and subsequent password verification.
			for range 2 {
				response := submitLogin(router, string(body))
				if response.Code != http.StatusOK {
					t.Fatalf("status = %d, want 200", response.Code)
				}
				var profile entity.ProfileResponse
				if err := json.Unmarshal(response.Body.Bytes(), &profile); err != nil {
					t.Fatal(err)
				}
				if profile.Username != tt.username || profile.ID == "" || len(response.Result().Cookies()) == 0 {
					t.Fatal("valid credentials did not produce the expected profile and session")
				}
			}
		})
	}
}
