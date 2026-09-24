package editor

import (
	"bytes"
	"crypto/sha256"
	_ "embed"
	"encoding/base64"
)

// Only this exact script is allowed by the admin CSP. Template scripts remain
// blocked; the script takes no server-interpolated values or external inputs.
//
//go:embed activity.js
var activityScript []byte

func activityPolicy() string {
	digest := sha256.Sum256(activityScript)
	return "; connect-src 'self'; script-src 'sha256-" + base64.StdEncoding.EncodeToString(digest[:]) + "'"
}

func withActivity(data []byte) []byte {
	script := append([]byte("<script>"), activityScript...)
	script = append(script, []byte("</script>")...)
	if offset := bytes.LastIndex(bytes.ToLower(data), []byte("</body>")); offset >= 0 {
		result := append([]byte(nil), data[:offset]...)
		result = append(result, script...)
		return append(result, data[offset:]...)
	}
	return append(data, script...)
}
