package oauth

import (
	"context"
	"encoding/base64"
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
				// Full IMAP scope URLs required for personal + work accounts
				"https://outlook.office.com/IMAP.AccessAsUser.All",
				"https://outlook.office.com/SMTP.Send",
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
	switch provider {
	case ProviderGoogle:
		return getGoogleUserInfo(ctx, accessToken)
	case ProviderMicrosoft:
		return getMicrosoftUserInfo(ctx, accessToken)
	default:
		return nil, fmt.Errorf("unknown provider: %s", provider)
	}
}

func getGoogleUserInfo(ctx context.Context, accessToken string) (*UserInfo, error) {
	raw, err := fetchJSON(ctx, "https://www.googleapis.com/oauth2/v2/userinfo", accessToken)
	if err != nil {
		return nil, err
	}
	info := &UserInfo{}
	info.Email, _ = raw["email"].(string)
	info.Name, _ = raw["name"].(string)
	info.PhotoURL, _ = raw["picture"].(string)
	return info, nil
}

func getMicrosoftUserInfo(ctx context.Context, accessToken string) (*UserInfo, error) {
	raw, err := fetchJSON(ctx, "https://graph.microsoft.com/v1.0/me", accessToken)
	if err != nil {
		return nil, err
	}
	info := &UserInfo{}

	// Email: try mail, then userPrincipalName (personal accounts often have only UPN)
	if mail, ok := raw["mail"].(string); ok && mail != "" {
		info.Email = mail
	} else if upn, ok := raw["userPrincipalName"].(string); ok {
		info.Email = upn
	}

	// Display name
	if name, ok := raw["displayName"].(string); ok {
		info.Name = name
	}

	// Avatar — separate endpoint, failure is non-fatal
	photoURL, err := getMicrosoftPhoto(ctx, accessToken)
	if err == nil {
		info.PhotoURL = photoURL
	}

	return info, nil
}

// getMicrosoftPhoto fetches the user's profile photo as a data URI.
func getMicrosoftPhoto(ctx context.Context, accessToken string) (string, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", "https://graph.microsoft.com/v1.0/me/photo/$value", nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return "", fmt.Errorf("photo not available: %d", resp.StatusCode)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	contentType := resp.Header.Get("Content-Type")
	if contentType == "" {
		contentType = "image/jpeg"
	}

	return fmt.Sprintf("data:%s;base64,%s", contentType,
		fmt.Sprintf("%s", encodeBase64(data))), nil
}

func fetchJSON(ctx context.Context, url, accessToken string) (map[string]interface{}, error) {	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
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
		return nil, fmt.Errorf("request failed (%d): %s", resp.StatusCode, string(body))
	}

	var raw map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		return nil, err
	}
	return raw, nil
}

func encodeBase64(data []byte) string {
	return base64.StdEncoding.EncodeToString(data)
}
