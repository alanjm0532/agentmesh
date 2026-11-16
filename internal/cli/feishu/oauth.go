package feishu

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/alanjm0532/agentmesh/internal/cli/repository"
)

const authorizeEndpoint = "https://accounts.feishu.cn/open-apis/authen/v1/authorize"
const tokenEndpoint = "https://accounts.feishu.cn/oauth/v3/token"

type authorizationResult struct {
	code string
	err  error
}

func (c *Client) Login(ctx context.Context, announce func(string)) error {
	redirect, err := url.Parse(c.config.RedirectURI)
	if err != nil {
		return err
	}
	if redirect.Scheme != "http" || !slices.Contains([]string{"localhost", "127.0.0.1", "::1"}, redirect.Hostname()) {
		return fmt.Errorf("Feishu redirect_uri must use HTTP on localhost or a loopback IP")
	}
	_, port, err := net.SplitHostPort(redirect.Host)
	if err != nil {
		return fmt.Errorf("Feishu redirect_uri must include a local port: %w", err)
	}
	portNumber, err := strconv.Atoi(port)
	if err != nil || portNumber < 1 || portNumber > 65535 {
		return fmt.Errorf("Feishu redirect_uri port must be between 1 and 65535")
	}
	listener, err := net.Listen("tcp", redirect.Host)
	if err != nil {
		return fmt.Errorf("listen for Feishu authorization: %w", err)
	}
	defer func() { _ = listener.Close() }()
	state := randomVerifier()
	verifier := randomVerifier()
	challenge := sha256.Sum256([]byte(verifier))
	parameters := url.Values{
		"client_id": {c.config.AppID}, "response_type": {"code"},
		"redirect_uri": {c.config.RedirectURI},
		"scope":        {"im:message im:message.send_as_user offline_access"},
		"state":        {state}, "code_challenge": {base64.RawURLEncoding.EncodeToString(challenge[:])},
		"code_challenge_method": {"S256"},
	}
	results := make(chan authorizationResult, 1)
	mux := http.NewServeMux()
	mux.HandleFunc("GET /"+strings.TrimPrefix(redirect.Path, "/"), func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-store")
		w.Header().Set("Referrer-Policy", "no-referrer")
		query := r.URL.Query()
		if subtle.ConstantTimeCompare([]byte(query.Get("state")), []byte(state)) != 1 {
			http.Error(w, "Invalid authorization state", http.StatusBadRequest)
			return
		}
		result := authorizationResult{code: query.Get("code")}
		if query.Has("error") {
			result.err = fmt.Errorf("Feishu authorization was denied or failed: %s", query.Get("error"))
		}
		_, _ = fmt.Fprintln(w, "Authorization callback received. Return to the CLI for the result.")
		select {
		case results <- result:
		default:
		}
	})
	server := &http.Server{Handler: mux, ReadHeaderTimeout: 5 * time.Second}
	defer func() { _ = server.Close() }()
	serveErrors := make(chan error, 1)
	go func() { serveErrors <- server.Serve(listener) }()
	announce(authorizeEndpoint + "?" + parameters.Encode())
	select {
	case <-ctx.Done():
		return fmt.Errorf("Feishu login canceled or expired: %w", ctx.Err())
	case err = <-serveErrors:
		return fmt.Errorf("Feishu callback server: %w", err)
	case result := <-results:
		if result.err != nil {
			return result.err
		}
		value, err := c.requestToken(ctx, map[string]string{
			"grant_type": "authorization_code", "code": result.code,
			"redirect_uri": c.config.RedirectURI, "code_verifier": verifier,
		})
		if err != nil {
			return err
		}
		return c.store.SaveToken(ctx, value)
	}
}

func randomVerifier() string {
	var data [32]byte
	_, _ = rand.Read(data[:])
	return base64.RawURLEncoding.EncodeToString(data[:])
}

func (c *Client) refresh(ctx context.Context, refreshToken string) (repository.FeishuToken, error) {
	return c.requestToken(ctx, map[string]string{"grant_type": "refresh_token", "refresh_token": refreshToken})
}

func (c *Client) requestToken(ctx context.Context, form map[string]string) (repository.FeishuToken, error) {
	form["client_id"] = c.config.AppID
	form["client_secret"] = c.config.AppSecret
	var payload struct {
		Code                  int    `json:"code"`
		AccessToken           string `json:"access_token"`
		RefreshToken          string `json:"refresh_token"`
		ExpiresIn             int64  `json:"expires_in"`
		RefreshTokenExpiresIn int64  `json:"refresh_token_expires_in"`
		Scope                 string `json:"scope"`
		Error                 string `json:"error"`
		ErrorDescription      string `json:"error_description"`
	}
	started := time.Now().Unix()
	response, err := c.http.R().SetContext(ctx).SetFormData(form).SetResult(&payload).SetError(&payload).Post(tokenEndpoint)
	if err != nil {
		return repository.FeishuToken{}, fmt.Errorf("request Feishu user token: %w", err)
	}
	if response.StatusCode() != http.StatusOK || payload.Code != 0 || payload.ExpiresIn <= 0 || payload.RefreshTokenExpiresIn <= 0 {
		return repository.FeishuToken{}, fmt.Errorf("Feishu token request failed (HTTP %d, code %d, %s): %s; run agentmesh feishu login to reauthorize", response.StatusCode(), payload.Code, payload.Error, payload.ErrorDescription)
	}
	return repository.FeishuToken{
		AccessToken: payload.AccessToken, RefreshToken: payload.RefreshToken,
		ExpiresAt: started + payload.ExpiresIn, RefreshExpiresAt: started + payload.RefreshTokenExpiresIn,
		Scope: payload.Scope,
	}, nil
}
