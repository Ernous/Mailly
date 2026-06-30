package oauth

import (
	"fmt"
)

func GenerateXOAuth2String(user, accessToken string) string {
	return fmt.Sprintf("user=%s\x01auth=Bearer %s\x01\x01", user, accessToken)
}
