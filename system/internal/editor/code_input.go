package editor

import (
	"crypto/sha256"
	_ "embed"
	"encoding/base64"
	"net/http"
	"strings"
)

//go:embed code_input.js
var codeInputJS string

func appendCodeInput(w http.ResponseWriter, data []byte) []byte {
	return appendTrustedScript(w, data, codeInputJS)
}

func appendTrustedScript(w http.ResponseWriter, data []byte, script string) []byte {
	hash := sha256.Sum256([]byte(script))
	policy := w.Header().Get("Content-Security-Policy")
	token := " 'sha256-" + base64.StdEncoding.EncodeToString(hash[:]) + "'"
	if strings.Contains(policy, "script-src ") {
		policy = strings.Replace(policy, "script-src ", "script-src"+token+" ", 1)
	} else {
		policy += "; script-src" + token
	}
	w.Header().Set("Content-Security-Policy", policy)
	return append(data, []byte("<script>"+script+"</script>")...)
}
