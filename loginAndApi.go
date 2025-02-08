package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"time"

	"github.com/chromedp/chromedp"
	"github.com/coreos/go-oidc/v3/oidc"
	"golang.org/x/oauth2"
)

// getLocalStorageAction returns a chromedp.ActionFunc that retrieves the value
// associated with the specified key from the browser's local storage.
func getLocalStorageAction(key string, value *string) chromedp.ActionFunc {
	return chromedp.ActionFunc(func(ctx context.Context) error {
		return chromedp.Evaluate(
			`localStorage.getItem('`+key+`')`,
			value,
		).Do(ctx)
	})
}

// loginAndApi performs a login operation and executes an API request.
// It initializes a Chrome instance, and then uses it to perform login tasks in non-headless mode.
// It performs login tasks to retrieve tokens from local storage,
// and then uses these tokens to execute an API request.
func loginAndApi(ctx context.Context) error {
	ctx, cancel := chromedp.NewExecAllocator(
		ctx,
		append(chromedp.DefaultExecAllocatorOptions[:],
			chromedp.Flag("headless", false),
			chromedp.Flag("disable-gpu", false),
			chromedp.Flag("hide-scrollbars", false),
		)...,
	)
	defer cancel()

	// create chrome instance
	ctx, cancel = chromedp.NewContext(ctx)
	defer cancel()

	ctxWithTimeout, cancel := context.WithTimeout(ctx, 120*time.Second)
	defer cancel()

	var accessToken, idToken, refreshToken string
	if err := chromedp.Run(ctxWithTimeout,
		loginTasks(ctx, testServerUrl, &accessToken, &idToken, &refreshToken),
	); err != nil {
		return err
	}

	log.Printf("Access Token: %s", accessToken)
	log.Printf("ID Token: %s", idToken)
	log.Printf("Refresh Token: %s", refreshToken)

	// Execute API request
	testToken := Token{
		AccessToken:  accessToken,
		IdToken:      idToken,
		RefreshToken: refreshToken,
	}

	if err := execApi(ctx, testToken, testServerUrl, providerUrl); err != nil {
		return fmt.Errorf("failed to execute API: %v", err)
	}
	log.Println("API executed successfully")

	return nil
}

func loginTasks(_ context.Context, baseURL string, accessToken, idToken, refreshToken *string) chromedp.Tasks {
	return chromedp.Tasks{
		// Navigate to base URL
		chromedp.Navigate(baseURL),

		// Display message to wait for user login
		chromedp.ActionFunc(func(ctx context.Context) error {
			log.Println("Waiting for user to complete login...")
			return nil
		}),

		// Wait for dashboard page to be displayed (login completion check)
		chromedp.WaitVisible(`#userInfo`),

		// Get tokens from LocalStorage
		getLocalStorageAction("access_token", accessToken),
		getLocalStorageAction("id_token", idToken),
		getLocalStorageAction("refresh_token", refreshToken),

		// Output log
		chromedp.ActionFunc(func(ctx context.Context) error {
			log.Println("Login completed and tokens retrieved")
			return nil
		}),
	}
}

type Token struct {
	AccessToken  string `json:"access_token"`
	IdToken      string `json:"id_token"`
	RefreshToken string `json:"refresh_token"`
}

func execApi(ctx context.Context, token Token, baseUrl string, providerUrl string) error {
	// OIDC provider configuration
	provider, err := oidc.NewProvider(ctx, providerUrl)
	if err != nil {
		return fmt.Errorf("failed to get provider: %v", err)
	}

	// OAuth2 configuration
	oauth2Config := oauth2.Config{
		ClientID:    "app",
		RedirectURL: "http://localhost:3000/callback",
		Endpoint:    provider.Endpoint(),
		Scopes:      []string{oidc.ScopeOpenID, "profile", "email"},
	}

	// Token configuration
	ts := oauth2Config.TokenSource(ctx, &oauth2.Token{
		RefreshToken: token.RefreshToken,
		TokenType:    "Bearer",
	})

	// Create HTTP client
	client := oauth2.NewClient(ctx, ts)

	// Execute API request
	resp, err := client.Get(baseUrl + "/api/hello")
	if err != nil {
		return fmt.Errorf("failed to call API: %v", err)
	}
	defer resp.Body.Close()

	// Read response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response: %v", err)
	}

	log.Printf("API Response: %s", string(body))
	return nil
}
