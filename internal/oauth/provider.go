package oauth

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"golang.org/x/oauth2"
)

type Provider string

const (
	ProviderGoogle    Provider = "google"
	ProviderMicrosoft Provider = "microsoft"
)

type ProviderConfig struct {
	Provider    Provider
	DisplayName string
	IMAPHost    string
	IMAPPort    int
	SMTPHost    string
	SMTPPort    int
	OAuthConfig *oauth2.Config
}

var Providers = map[Provider]*ProviderConfig{}

func RegisterGoogle(clientID, clientSecret, redirectURI string) {
	if clientID == "" || clientSecret == "" {
		return
	}
	Providers[ProviderGoogle] = &ProviderConfig{
		Provider:    ProviderGoogle,
		DisplayName: "Google (Gmail)",
		IMAPHost:    "imap.gmail.com",
		IMAPPort:    993,
		SMTPHost:    "smtp.gmail.com",
		SMTPPort:    587,
		OAuthConfig: &oauth2.Config{
			ClientID:     clientID,
			ClientSecret: clientSecret,
			Endpoint: oauth2.Endpoint{
				AuthURL:  "https://accounts.google.com/o/oauth2/v2/auth",
				TokenURL: "https://oauth2.googleapis.com/token",
			},
			Scopes:      []string{"https://mail.google.com/", "openid", "email", "profile"},
			RedirectURL: redirectURI,
		},
	}
}

func RegisterMicrosoft(clientID, clientSecret, redirectURI string) {
	if clientID == "" || clientSecret == "" {
		return
	}
	Providers[ProviderMicrosoft] = &ProviderConfig{
		Provider:    ProviderMicrosoft,
		DisplayName: "Microsoft (Outlook)",
		IMAPHost:    "outlook.office365.com",
		IMAPPort:    993,
		SMTPHost:    "smtp-mail.outlook.com",
		SMTPPort:    587,
		OAuthConfig: &oauth2.Config{
			ClientID:     clientID,
			ClientSecret: clientSecret,
			Endpoint: oauth2.Endpoint{
				AuthURL:  "https://login.microsoftonline.com/consumers/oauth2/v2.0/authorize",
				TokenURL: "https://login.microsoftonline.com/consumers/oauth2/v2.0/token",
			},
			Scopes: []string{
				"User.Read",
				"IMAP.AccessAsUser.All",
				"SMTP.Send",
				"offline_access",
				"openid",
				"email",
				"profile",
			},
			RedirectURL: redirectURI,
		},
	}
}

func GetAuthURL(provider Provider, state string) (string, error) {
	cfg, ok := Providers[provider]
	if !ok {
		return "", fmt.Errorf("unknown provider: %s", provider)
	}
	return cfg.OAuthConfig.AuthCodeURL(state, oauth2.AccessTypeOffline, oauth2.SetAuthURLParam("prompt", "consent")), nil
}

func ExchangeCode(ctx context.Context, provider Provider, code string) (*oauth2.Token, error) {
	cfg, ok := Providers[provider]
	if !ok {
		return nil, fmt.Errorf("unknown provider: %s", provider)
	}
	return cfg.OAuthConfig.Exchange(ctx, code)
}

func GetProviderConfig(provider Provider) (*ProviderConfig, bool) {
	cfg, ok := Providers[provider]
	return cfg, ok
}

func RefreshToken(ctx context.Context, provider Provider, oldToken *oauth2.Token) (*oauth2.Token, error) {
	cfg, ok := Providers[provider]
	if !ok {
		return nil, fmt.Errorf("unknown provider: %s", provider)
	}
	ts := cfg.OAuthConfig.TokenSource(ctx, oldToken)
	newToken, err := ts.Token()
	if err != nil {
		return nil, fmt.Errorf("token refresh failed: %w", err)
	}
	return newToken, nil
}

type UserInfo struct {
	Email    string
	Name     string
	PhotoURL string
}

func GetUserInfo(ctx context.Context, provider Provider, accessToken string) (*UserInfo, error) {
	var url string
	switch provider {
	case ProviderGoogle:
		url = "https://www.googleapis.com/oauth2/v2/userinfo"
	case ProviderMicrosoft:
		url = "https://graph.microsoft.com/v1.0/me"
	default:
		return nil, fmt.Errorf("unknown provider: %s", provider)
	}

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("userinfo request failed (%d): %s", resp.StatusCode, string(body))
	}

	var raw map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		return nil, err
	}

	info := &UserInfo{}

	if email, ok := raw["email"].(string); ok {
		info.Email = email
	} else if emails, ok := raw["mail"].(string); ok {
		info.Email = emails
	}

	if name, ok := raw["name"].(string); ok {
		info.Name = name
	} else if display, ok := raw["displayName"].(string); ok {
		info.Name = display
	}

	if picture, ok := raw["picture"].(string); ok {
		info.PhotoURL = picture
	}

	return info, nil
}
