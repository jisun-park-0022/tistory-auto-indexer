package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"

	"github.com/joho/godotenv"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

const (
	scope       = "https://www.googleapis.com/auth/webmasters"
	redirectURI = "http://localhost:9090/callback"
	callbackPort = ":9090"
)

func main() {
	_ = godotenv.Load(".env")

	clientID := os.Getenv("GOOGLE_CLIENT_ID")
	clientSecret := os.Getenv("GOOGLE_CLIENT_SECRET")
	if clientID == "" || clientSecret == "" {
		slog.Error("GOOGLE_CLIENT_ID and GOOGLE_CLIENT_SECRET must be set in .env")
		os.Exit(1)
	}

	cfg := &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		RedirectURL:  redirectURI,
		Scopes:       []string{scope},
		Endpoint:     google.Endpoint,
	}

	authURL := cfg.AuthCodeURL("state", oauth2.AccessTypeOffline, oauth2.ApprovalForce)
	fmt.Println("========================================")
	fmt.Println("아래 URL을 브라우저에서 열어 Google 계정으로 로그인하세요:")
	fmt.Println()
	fmt.Println(authURL)
	fmt.Println()
	fmt.Println("승인 후 자동으로 refresh token이 출력됩니다.")
	fmt.Println("========================================")

	codeCh := make(chan string, 1)
	srv := &http.Server{Addr: callbackPort}

	http.HandleFunc("/callback", func(w http.ResponseWriter, r *http.Request) {
		code := r.URL.Query().Get("code")
		if code == "" {
			http.Error(w, "code not found", http.StatusBadRequest)
			return
		}
		fmt.Fprintln(w, "인증 완료! 터미널로 돌아가세요.")
		codeCh <- code
	})

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("callback server error", "error", err)
			os.Exit(1)
		}
	}()

	code := <-codeCh
	_ = srv.Shutdown(context.Background())

	token, err := cfg.Exchange(context.Background(), code)
	if err != nil {
		slog.Error("failed to exchange code for token", "error", err)
		os.Exit(1)
	}

	fmt.Println()
	fmt.Println("========================================")
	fmt.Println("GOOGLE_REFRESH_TOKEN (아래 값을 .env에 붙여넣기):")
	fmt.Println()
	fmt.Println(token.RefreshToken)
	fmt.Println("========================================")
}
