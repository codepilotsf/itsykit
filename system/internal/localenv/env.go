// Package localenv reads literal local configuration. It never executes shell code.
package localenv

import (
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
)

const maxSize = 64 << 10

func read(root *os.Root, name string) ([]byte, error) {
	info, err := root.Lstat(name)
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil || !info.Mode().IsRegular() {
		return nil, fmt.Errorf("%s must be a regular file, not a link.", name)
	}
	f, err := root.Open(name)
	if err != nil {
		return nil, fmt.Errorf("Could not read %s.", name)
	}
	defer f.Close()
	opened, err := f.Stat()
	if err != nil || !os.SameFile(info, opened) {
		return nil, fmt.Errorf("%s changed while reading it.", name)
	}
	data, err := io.ReadAll(io.LimitReader(f, maxSize+1))
	if err != nil || len(data) > maxSize {
		return nil, fmt.Errorf("%s could not be read or exceeds 64 KiB.", name)
	}
	return data, nil
}

func parse(data []byte) (map[string]string, error) {
	values := map[string]string{}
	for n, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		line = strings.TrimPrefix(line, "export ")
		key, value, ok := strings.Cut(line, "=")
		key, value = strings.TrimSpace(key), strings.TrimSpace(value)
		if !ok || key == "" || strings.ContainsAny(key, " \t'\"$") {
			return nil, fmt.Errorf("Check .env line %d: use NAME=value assignments.", n+1)
		}
		if strings.HasPrefix(value, "\"") {
			var err error
			value, err = strconv.Unquote(value)
			if err != nil {
				return nil, fmt.Errorf("Check the quoted value on .env line %d.", n+1)
			}
		} else if strings.HasPrefix(value, "'") {
			if len(value) < 2 || !strings.HasSuffix(value, "'") {
				return nil, fmt.Errorf("Check the quoted value on .env line %d.", n+1)
			}
			value = value[1 : len(value)-1]
		}
		values[key] = value
	}
	return values, nil
}

func Read(directory string) (map[string]string, error) {
	root, err := os.OpenRoot(directory)
	if err != nil {
		return nil, err
	}
	defer root.Close()
	data, err := read(root, ".env")
	if err != nil {
		return nil, err
	}
	return parse(data)
}

func replace(root *os.Root, name string, data []byte, mode os.FileMode) error {
	if _, err := read(root, name); err != nil {
		return err
	}
	// Create an exclusive temporary file inside the pinned root.
	var f *os.File
	var temp string
	for i := 0; i < 100; i++ {
		temp = fmt.Sprintf(".env.flea-%d-%d", os.Getpid(), i)
		var err error
		f, err = root.OpenFile(temp, os.O_WRONLY|os.O_CREATE|os.O_EXCL, mode)
		if os.IsExist(err) {
			continue
		}
		if err != nil {
			return fmt.Errorf("Could not prepare %s.", name)
		}
		break
	}
	if f == nil {
		return fmt.Errorf("Could not prepare %s.", name)
	}
	defer root.Remove(temp)
	_, err := f.Write(data)
	if err == nil {
		err = f.Sync()
	}
	closeErr := f.Close()
	if err == nil {
		err = closeErr
	}
	if err != nil {
		return fmt.Errorf("Could not save %s.", name)
	}
	if err := root.Rename(temp, name); err != nil {
		return fmt.Errorf("Could not replace %s.", name)
	}
	return nil
}

// Save preserves unrelated lines, replaces only supplied keys, and writes secrets
// with owner-only permissions. The caller must first refuse a Git-tracked .env.
func Save(directory string, values map[string]string, keys []string) error {
	root, err := os.OpenRoot(directory)
	if err != nil {
		return err
	}
	defer root.Close()
	data, err := read(root, ".env")
	if err != nil {
		return err
	}
	if _, err := parse(data); err != nil {
		return err
	}
	ignore, err := read(root, ".gitignore")
	if err != nil {
		return err
	}
	// Append last so an earlier negation cannot expose this site's .env.
	const ignoreRules = "/.env.flea-*\n/.env"
	if !strings.HasSuffix(strings.TrimSpace(string(ignore)), "\n"+ignoreRules) && strings.TrimSpace(string(ignore)) != ignoreRules {
		if err := replace(root, ".gitignore", []byte(strings.TrimRight(string(ignore), "\n")+"\n"+ignoreRules+"\n"), 0644); err != nil {
			return err
		}
	}
	var lines []string
	for _, line := range strings.Split(strings.TrimRight(string(data), "\n"), "\n") {
		key, _, _ := strings.Cut(strings.TrimPrefix(strings.TrimSpace(line), "export "), "=")
		if _, ok := values[strings.TrimSpace(key)]; !ok {
			lines = append(lines, line)
		}
	}
	for _, key := range keys {
		lines = append(lines, key+"="+strconv.Quote(values[key]))
	}
	if len(strings.Join(lines, "\n"))+1 > maxSize {
		return fmt.Errorf("The updated .env would exceed 64 KiB. Email settings were not saved.")
	}
	return replace(root, ".env", []byte(strings.Join(lines, "\n")+"\n"), 0600)
}
