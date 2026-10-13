package site

import "os"

// ReadManagedFile and WriteManagedFile reuse project confinement for fixed,
// caller-selected local/runtime paths. Never pass request-selected paths.
func ReadManagedFile(root, name string) ([]byte, error) {
	r, e := os.OpenRoot(root)
	if e != nil {
		return nil, e
	}
	defer r.Close()
	return readFile(r, name)
}
func WriteManagedFile(root, name string, b []byte) error {
	r, e := os.OpenRoot(root)
	if e != nil {
		return e
	}
	defer r.Close()
	return writeSiteFile(r, name, b)
}
