package mailer

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// Checker is separate from Sender: authentication and contact handlers only send.
// Future providers can implement their own non-sending setup check.
type Checker interface {
	Check(context.Context) (string, error)
}

func (s *Resend) Check(ctx context.Context) (string, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", "https://api.resend.com/domains?limit=1", nil)
	if err != nil {
		return "", fmt.Errorf("Could not prepare the Resend connection check.")
	}
	req.Header.Set("Authorization", "Bearer "+s.settings.APIKey)
	resp, err := s.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("Could not connect to Resend. Check your connection and try again.")
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusOK {
		return "Connected to Resend.\nNo email was sent; delivery has not been tested.", nil
	}
	var body struct {
		Name string `json:"name"`
	}
	// Only the documented error code is inspected. Never display provider bodies.
	if resp.StatusCode == http.StatusUnauthorized && json.NewDecoder(io.LimitReader(resp.Body, 4096)).Decode(&body) == nil && body.Name == "restricted_api_key" {
		return "Connected to Resend with a sending-only key.\nDomain verification and delivery have not been tested. No email was sent.", nil
	}
	return "", fmt.Errorf("Resend could not verify the connection (HTTP %d). Check your API key and try again. Nothing was saved.", resp.StatusCode)
}
