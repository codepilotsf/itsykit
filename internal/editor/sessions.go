package editor

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"syscall"
	"time"

	"itsy/internal/state"
)

// Sessions are runtime state, never release content. Each transaction reads the
// current file under a separate lock, so a stale handler cannot revive a logout.
// Pending email codes intentionally remain in memory.
type savedSession struct {
	Email      string    `json:"email"`
	Host       string    `json:"host"`
	CSRF       string    `json:"csrf"`
	LastActive time.Time `json:"last_active"`
	Expires    time.Time `json:"expires"`
}

func privateRegular(info os.FileInfo) bool {
	stat, ok := info.Sys().(*syscall.Stat_t)
	return info.Mode().IsRegular() && info.Mode().Perm()&0077 == 0 && ok && stat.Nlink == 1
}

func (h *Handler) sessionDuration() time.Duration {
	return time.Duration(h.options.SessionIdleDays) * 24 * time.Hour
}

// Caller holds h.mu. Do not acquire the content/deployment lock here: session
// writes are independent of content saves and release activation.
func (h *Handler) sessionTransaction(change func(map[[32]byte]session) error) error {
	if h.options.SessionIdleDays < 1 || h.options.SessionIdleDays > 365 || h.options.SessionPath == "" {
		return fmt.Errorf("session storage is not configured")
	}
	parent := filepath.Dir(h.options.SessionPath)
	info, err := os.Lstat(parent)
	if err != nil || !info.IsDir() {
		return fmt.Errorf("session directory must be a real directory")
	}
	root, err := os.OpenRoot(parent)
	if err != nil {
		return err
	}
	defer root.Close()
	opened, err := root.Stat(".")
	if err != nil || !os.SameFile(info, opened) {
		return fmt.Errorf("session directory changed")
	}
	const lockName = ".sessions-lock"
	if info, err := root.Lstat(lockName); err == nil {
		if !privateRegular(info) {
			return fmt.Errorf("session lock must be a private regular file")
		}
	} else if !os.IsNotExist(err) {
		return err
	}
	unlock, err := state.LockRoot(root, lockName)
	if err != nil {
		return err
	}
	defer unlock()
	name := filepath.Base(h.options.SessionPath)
	var before []byte
	if info, err := root.Lstat(name); err == nil {
		if !privateRegular(info) {
			return fmt.Errorf("sessions.json must be a private regular file (0600)")
		}
		f, err := root.Open(name)
		if err != nil {
			return err
		}
		actual, err := f.Stat()
		if err != nil || !os.SameFile(info, actual) {
			f.Close()
			return fmt.Errorf("session file changed")
		}
		before, err = io.ReadAll(io.LimitReader(f, (2<<20)+1))
		f.Close()
		if err != nil {
			return err
		}
		if len(before) > 2<<20 {
			return fmt.Errorf("session file exceeds 2 MiB")
		}
	} else if !os.IsNotExist(err) {
		return err
	}
	stored := map[string]savedSession{}
	if before != nil {
		decoder := json.NewDecoder(bytes.NewReader(before))
		decoder.DisallowUnknownFields()
		if err := decoder.Decode(&stored); err != nil {
			return fmt.Errorf("invalid session file")
		}
		if stored == nil || decoder.Decode(new(any)) != io.EOF {
			return fmt.Errorf("invalid session file")
		}
	}
	if len(stored) > maxState {
		return fmt.Errorf("too many stored sessions")
	}
	values := map[[32]byte]session{}
	now := h.now()
	for digest, v := range stored {
		b, err := hex.DecodeString(digest)
		if err != nil || len(b) != 32 || len(v.CSRF) != 64 || v.Host == "" || v.Email == "" || v.LastActive.IsZero() || v.Expires.Before(v.LastActive) {
			return fmt.Errorf("invalid session record")
		}
		// A shorter configured idle window takes effect for existing sessions too.
		if limit := v.LastActive.Add(h.sessionDuration()); limit.Before(v.Expires) {
			v.Expires = limit
		}
		if now.Before(v.Expires) && h.allowed[v.Email] {
			values[[32]byte(b)] = session{email: v.Email, host: v.Host, csrf: v.CSRF, expires: v.Expires, lastActive: v.LastActive}
		}
	}
	if change != nil {
		if err := change(values); err != nil {
			return err
		}
	}
	if len(values) > maxState {
		return fmt.Errorf("too many sessions")
	}
	after := map[string]savedSession{}
	for digest, v := range values {
		after[hex.EncodeToString(digest[:])] = savedSession{v.email, v.host, v.csrf, v.lastActive, v.expires}
	}
	data, err := json.MarshalIndent(after, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	if !bytes.Equal(before, data) && (before != nil || len(values) != 0) {
		suffix, err := token()
		if err != nil {
			return err
		}
		temp := ".sessions-" + suffix
		f, err := root.OpenFile(temp, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
		if err != nil {
			return err
		}
		defer root.Remove(temp)
		_, err = f.Write(data)
		if err == nil {
			err = f.Sync()
		}
		closeErr := f.Close()
		if err != nil {
			return err
		}
		if closeErr != nil {
			return closeErr
		}
		if err = root.Rename(temp, name); err != nil {
			return err
		}
		dir, err := root.Open(".")
		if err != nil {
			return err
		}
		err = dir.Sync()
		dir.Close()
		if err != nil {
			return err
		}
	}
	h.sessions = values
	return nil
}

// PruneSessions applies the current access/expiration policy at startup or after
// a local configuration change. Host config validation must remain read-only.
func PruneSessions(o Options) error {
	h := New(o)
	return h.sessionTransaction(nil)
}

func (h *Handler) sessionError(w http.ResponseWriter, err error) {
	if h.options.Logger != nil {
		h.options.Logger.Printf("Admin session storage unavailable: %v", err)
	}
	http.Error(w, "Sign-in is temporarily unavailable. Please try again shortly.", 503)
}

func (h *Handler) renew(w http.ResponseWriter, r *http.Request) bool {
	h.mu.Lock()
	err := h.sessionTransaction(func(values map[[32]byte]session) error {
		k := key(h.getCookie(r, "session"))
		s, ok := values[k]
		if !ok || s.host != authority(r) {
			return fmt.Errorf("session no longer active")
		}
		s.lastActive = h.now()
		s.expires = s.lastActive.Add(h.sessionDuration())
		values[k] = s
		return nil
	})
	h.mu.Unlock()
	if err != nil {
		h.sessionError(w, err)
		return false
	}
	h.cookie(w, "session", h.getCookie(r, "session"), int(h.sessionDuration().Seconds()))
	return true
}
