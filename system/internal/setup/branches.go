package setup

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"itsy/internal/sitefiles"
	"net"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"itsy/internal/admin"
	"itsy/internal/site"
)

type DeploymentConfig struct {
	Branches map[string]DeploySettings `json:"branches"`
}

func readDeploymentConfig(root string) (DeploymentConfig, error) {
	c := DeploymentConfig{Branches: map[string]DeploySettings{}}
	data, err := site.ReadManagedFile(root, "itsy.config.json")
	if os.IsNotExist(err) {
		return c, nil
	}
	if err != nil {
		return c, err
	}
	var shape map[string]json.RawMessage
	if err = json.Unmarshal(data, &shape); err != nil {
		return c, err
	}
	if shape == nil {
		return c, fmt.Errorf("itsy.config.json must be an object")
	}
	data, exists := shape["deploy"]
	if !exists {
		return c, nil
	}
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.DisallowUnknownFields()
	err = dec.Decode(&c)
	if err != nil {
		return c, fmt.Errorf("itsy.config.json: %w", err)
	}
	if c.Branches == nil {
		c.Branches = map[string]DeploySettings{}
	}
	return c, nil
}

func saveDeploymentConfig(root string, c DeploymentConfig) error {
	data, err := site.ReadManagedFile(root, "itsy.config.json")
	if err != nil {
		return err
	}
	var fields map[string]json.RawMessage
	if err = json.Unmarshal(data, &fields); err != nil || fields == nil {
		return fmt.Errorf("itsy.config.json must be a valid object")
	}
	fields["deploy"], err = json.Marshal(c)
	if err != nil {
		return err
	}
	data, err = json.MarshalIndent(fields, "", "  ")
	if err != nil {
		return err
	}
	return site.WriteManagedFile(root, "itsy.config.json", append(data, '\n'))
}

func (s DeploySettings) domains() []string {
	result := admin.Domains(s.Domain)
	if s.WWWRedirect {
		result = append(result, "www."+s.Domain)
	}
	return result
}

func validateMapping(branch string, s DeploySettings, c DeploymentConfig) error {
	if _, err := targetName(s.Target); err != nil {
		return err
	}
	o := admin.Options{Archive: "snapshot", Name: s.Name, Domains: s.domains(), HTTPS: s.HTTPS, Email: s.Email, AgreeTOS: s.AgreeTOS}
	if err := o.Validate(); err != nil {
		return err
	}
	if s.WWWRedirect && (net.ParseIP(s.Domain) != nil || strings.Contains(s.Domain, ",") || strings.HasPrefix(s.Domain, "www.")) {
		return fmt.Errorf("WWW redirect requires one primary domain without the www prefix")
	}
	for _, ip := range s.ExpectedIPs {
		if net.ParseIP(ip) == nil {
			return fmt.Errorf("invalid expected IP %q", ip)
		}
	}
	for other, v := range c.Branches {
		if other == branch {
			continue
		}
		if v.Target == s.Target && v.Name == s.Name {
			return fmt.Errorf("site name %s on %s already belongs to branch %s", s.Name, s.Target, other)
		}
		for _, a := range s.domains() {
			for _, b := range v.domains() {
				if a == b {
					return fmt.Errorf("hostname %s already belongs to branch %s", a, other)
				}
			}
		}
	}
	return nil
}

// SelectDeployment confirms and persists a branch mapping before any SSH or
// cleanliness check. A newly saved tracked config must be committed first.
func SelectDeployment(ctx context.Context, root string, overrides DeploySettings, provided map[string]bool, reconfigure, httpsOnly bool, in io.Reader, out io.Writer) (DeploySettings, error) {
	state, err := currentGit(ctx, root, false)
	if err != nil {
		return DeploySettings{}, err
	}
	c, err := readDeploymentConfig(root)
	if err != nil {
		return DeploySettings{}, err
	}
	s, exists := c.Branches[state.Branch]
	if exists && !reconfigure {
		if len(provided) > 0 {
			return s, fmt.Errorf("branch %s has a saved destination; use -configure to confirm and save changes", state.Branch)
		}
		if httpsOnly && !s.HTTPS {
			return s, fmt.Errorf("HTTPS is disabled for branch %s; use itsy https -configure", state.Branch)
		}
		if err = validateMapping(state.Branch, s, c); err != nil {
			return s, err
		}
		if !strings.Contains(s.Target, "@") {
			fmt.Fprintf(out, "\nWhich SSH account should connect to %s?\n", s.Target)
			s.Target, err = sshUsername(ctx, s.Target, in, out)
			if err != nil {
				return s, err
			}
			if err = validateMapping(state.Branch, s, c); err != nil {
				return s, err
			}
			c.Branches[state.Branch] = s
			if err = saveDeploymentConfig(root, c); err != nil {
				return s, err
			}
			fmt.Fprintf(out, "\nSaved your SSH username in %s.\nItsyKit will offer to commit this change before deploying.\n", sitefiles.ResolveAt(root, sitefiles.ConfigName))
		}
		s.Branch = state.Branch
		return s, nil
	}
	if !exists {
		s.HTTPS = true
		s.WWWRedirect = true
		base, _ := filepath.Abs(root)
		slug := regexp.MustCompile(`[^A-Za-z0-9_-]+`).ReplaceAllString(filepath.Base(base), "-")
		s.Name = slug + "-" + regexp.MustCompile(`[^A-Za-z0-9_-]+`).ReplaceAllString(state.Branch, "-")
		if main, ok := c.Branches["main"]; ok {
			s.Target = main.Target
			s.Email = main.Email
			s.ExpectedIPs = main.ExpectedIPs
			if !strings.Contains(main.Domain, ",") && net.ParseIP(main.Domain) == nil {
				s.Domain = strings.ToLower(regexp.MustCompile(`[^a-zA-Z0-9-]+`).ReplaceAllString(state.Branch, "-")) + "." + main.Domain
				s.WWWRedirect = false
			}
		}
	}
	for key := range provided {
		switch key {
		case "target":
			s.Target = overrides.Target
		case "domain":
			s.Domain = overrides.Domain
		case "name":
			s.Name = overrides.Name
		case "https":
			s.HTTPS = overrides.HTTPS
		case "email":
			s.Email = overrides.Email
		case "agree-tos":
			s.AgreeTOS = overrides.AgreeTOS
		case "www-redirect":
			s.WWWRedirect = overrides.WWWRedirect
		case "expected-ips":
			s.ExpectedIPs = overrides.ExpectedIPs
		}
	}
	if httpsOnly {
		s.HTTPS = true
	}
	ask := func(label, def string) (string, error) { return ask(ctx, in, out, label, def) }
	yes := func(label string, def bool) (bool, error) {
		d := "n"
		if def {
			d = "y"
		}
		for {
			v, err := ask(label+" (y/n)", d)
			if err != nil {
				return false, err
			}
			switch strings.ToLower(v) {
			case "y", "yes":
				return true, nil
			case "n", "no":
				return false, nil
			}
			fmt.Fprintln(out, "Enter y or n.")
		}
	}
	fmt.Fprintf(out, "\nCurrent branch: %s\n\n", state.Branch)
	if s.Target, err = ask("Server hostname or IP", s.Target); err != nil {
		return s, err
	}
	if s.Target, err = sshUsername(ctx, s.Target, in, out); err != nil {
		return s, err
	}
	if s.Domain, err = ask(fmt.Sprintf("Site domain name for %s branch", state.Branch), s.Domain); err != nil {
		return s, err
	}
	s.Domain = strings.ToLower(s.Domain)
	if net.ParseIP(s.Domain) == nil {
		if s.WWWRedirect, err = yes("Redirect WWW to "+s.Domain+"?", s.WWWRedirect); err != nil {
			return s, err
		}
		fmt.Fprintln(out, "Let's Encrypt terms: https://letsencrypt.org/repository/")
		if s.HTTPS, err = yes("Enable HTTPS and accept Let's Encrypt terms?", s.HTTPS); err != nil {
			return s, err
		}
		s.AgreeTOS = s.HTTPS
	} else {
		s.HTTPS = false
		s.WWWRedirect = false
		fmt.Fprintln(out, "IP destinations use HTTP without a WWW redirect.")
	}
	if s.HTTPS {
		if s.Email == "" {
			if email, gitErr := git(ctx, root, "config", "--get", "user.email"); gitErr == nil {
				s.Email = strings.TrimSpace(email)
			}
		}
		if s.Email, err = ask("Certificate contact email", s.Email); err != nil {
			return s, err
		}
	}
	if err = validateMapping(state.Branch, s, c); err != nil {
		return s, err
	}
	if err := ctx.Err(); err != nil {
		return s, err
	}
	c.Branches[state.Branch] = s
	if err = saveDeploymentConfig(root, c); err != nil {
		return s, err
	}
	fmt.Fprintf(out, "Saved %s.\n", sitefiles.ResolveAt(root, sitefiles.ConfigName))
	s.Branch = state.Branch
	return s, nil
}
