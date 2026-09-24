package site

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"strings"

	"flea/internal/fleas"
	"flea/internal/sitefiles"
)

type fleaEdit struct{ Before, After string }

// A write-ahead record makes content-first extraction replayable after a crash.
// Both the original and intended bytes are accepted; unrelated edits are not.
type fleaTransaction struct {
	BeforeContent json.RawMessage
	Content       json.RawMessage
	Edits         map[string]fleaEdit
}

const transactionFile = "_/.fleas-pending.json"

func sameFleaJSON(a, b []byte) bool {
	var left, right bytes.Buffer
	return json.Compact(&left, a) == nil && json.Compact(&right, b) == nil && bytes.Equal(left.Bytes(), right.Bytes())
}

func finishFleaTransaction(root *os.Root, tx fleaTransaction) error {
	if _, err := fleas.Decode(tx.Content); err != nil {
		return err
	}
	current, readErr := readFile(root, "_/fleas.json")
	if readErr != nil && !os.IsNotExist(readErr) {
		return readErr
	}
	if !(os.IsNotExist(readErr) && (len(tx.BeforeContent) == 0 || string(tx.BeforeContent) == "null")) && !sameFleaJSON(current, tx.BeforeContent) && !sameFleaJSON(current, tx.Content) {
		return fmt.Errorf("_/fleas.json changed during extraction; recovery file retained. Reconcile it with _/.fleas-pending.json before retrying")
	}
	for name, e := range tx.Edits {
		if !fs.ValidPath(name) || !sitefiles.Included(name) || !strings.HasSuffix(name, ".html") {
			return fmt.Errorf("invalid flea transaction path")
		}
		data, err := readFile(root, name)
		if err != nil {
			return err
		}
		if string(data) != e.Before && string(data) != e.After {
			return fmt.Errorf("%s changed during extraction; restore the original or pending template before retrying", name)
		}
	}
	if err := writeSiteFile(root, "_/fleas.json", tx.Content); err != nil {
		return err
	}
	for name, e := range tx.Edits {
		current, err := readFile(root, name)
		if err != nil {
			return err
		}
		if string(current) != e.Before && string(current) != e.After {
			return fmt.Errorf("%s changed during extraction; recovery file retained", name)
		}
		if err := writeSiteFile(root, name, []byte(e.After)); err != nil {
			return err
		}
	}
	return root.Remove(transactionFile)
}
