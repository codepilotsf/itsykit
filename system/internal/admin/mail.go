package admin

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"itsy/internal/mailer"
)

func MailReady(configPath string) bool {
	name := filepath.Join(filepath.Dir(configPath), "mail.env")
	info, err := os.Lstat(name)
	if err != nil || !info.Mode().IsRegular() || info.Mode().Perm() != 0600 {
		return false
	}
	data, err := os.ReadFile(name)
	if err != nil || len(data) > 4096 {
		return false
	}
	var s mailer.Settings
	for _, line := range strings.Split(string(data), "\n") {
		k, v, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}
		value, err := strconv.Unquote(v)
		if err != nil {
			return false
		}
		switch k {
		case "RESEND_API_KEY":
			s.APIKey = value
		case "FLEA_EMAIL_FROM":
			s.From = value
		case "FLEA_EMAIL_NAME":
			s.Name = value
		}
	}
	return s.Validate() == nil
}

// ConfigureMail accepts credentials on stdin, never command-line arguments.
// systemd reads this root-only file before dropping to the service account.
func (m Manager) ConfigureMail(input io.Reader) error {
	unlock, err := m.lock()
	if err != nil {
		return err
	}
	defer unlock()
	if _, err = os.Lstat(m.pending()); !os.IsNotExist(err) {
		return fmt.Errorf("Recover the unfinished deployment before changing email settings.")
	}
	var s mailer.Settings
	d := json.NewDecoder(io.LimitReader(input, 4097))
	d.DisallowUnknownFields()
	if d.Decode(&s) != nil || d.Decode(new(any)) != io.EOF {
		return fmt.Errorf("Invalid email settings. Nothing was saved.")
	}
	if err = s.Validate(); err != nil {
		return err
	}
	data := []byte("RESEND_API_KEY=" + strconv.Quote(s.APIKey) + "\nFLEA_EMAIL_FROM=" + strconv.Quote(s.From) + "\nFLEA_EMAIL_NAME=" + strconv.Quote(s.Name) + "\n")
	return atomicWrite(filepath.Join(filepath.Dir(m.ConfigPath), "mail.env"), data, 0600)
}
