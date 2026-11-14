package mailer

import (
	"itsy/internal/localenv"
	"os"
)

// LocalSettings never changes the process environment. Explicit environment
// values, including empty values, take precedence over the site's literal .env.
func LocalSettings(directory string) (Settings, error) {
	values, err := localenv.Read(directory)
	if err != nil {
		return Settings{}, err
	}
	get := func(key string) string {
		if value, ok := os.LookupEnv(key); ok {
			return value
		}
		return values[key]
	}
	return Settings{APIKey: get("RESEND_API_KEY"), From: get("FLEA_EMAIL_FROM"), Name: get("FLEA_EMAIL_NAME")}, nil
}
