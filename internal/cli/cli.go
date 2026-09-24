package cli

import (
	"context"
	"flag"
	"fmt"
	"io"
	"io/fs"
	"log"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"flea/internal/buildinfo"
	"flea/internal/contact"
	"flea/internal/editor"
	"flea/internal/mailer"
	"flea/internal/setup"
	"flea/internal/site"
	"flea/templates"
)

const usage = `FleaCMS
Small websites. Plain HTML. Editable content.

Usage
  flea <command>

Build your site
  new <site-name>  Create a new site
  dev              Start the local development server
  status           Check your site for errors
  admin [email]    Add an owner who can edit content
  email            Set up email delivery

Publish your site
  setup            Install Flea on your server
  deploy           Publish the current Git branch
  https            Configure HTTPS for your site

Other commands
  version          Show the installed version
  help             Show this help

Get started
  flea new my-site
  cd my-site
  flea dev

Command details
  flea <command> --help
`

// Run returns 0 on success, 1 on project/runtime failure, 2 on usage errors.
func Run(ctx context.Context, args []string, out, errOut io.Writer) int {
	if len(args) == 0 {
		fmt.Fprint(out, usage)
		return 0
	}
	fail := func(err error) int { fmt.Fprintln(errOut, "flea:", err); return 1 }
	switch args[0] {
	case "email":
		if len(args) == 2 && (args[1] == "--help" || args[1] == "-h") {
			fmt.Fprintln(out, "Usage: flea email\n\nSet up Resend email delivery from your site folder.\nPrompts for a hidden API key, sender name, and verified-domain address.\nChecks the connection and saves settings to .env. No email is sent.\nAdmin sign-in loads settings automatically; contact forms need a dev restart.")
			return 0
		}
		if len(args) != 1 {
			fmt.Fprintln(errOut, "Usage: flea email")
			return 2
		}
		if err := setup.Email(ctx, ".", os.Stdin, out); err != nil {
			return fail(err)
		}
		return 0
	case "admin":
		if len(args) == 2 && (args[1] == "--help" || args[1] == "-h") {
			fmt.Fprintln(out, "Usage: flea admin [email]\n\nAdd an owner and set up email if needed. Local access updates automatically.\nView or remove owners in admin.emails in _/flea.config.\nCommit and deploy to apply live changes. No invitation is sent.")
			return 0
		}
		if len(args) > 2 || (len(args) == 2 && strings.HasPrefix(args[1], "-")) {
			fmt.Fprintln(errOut, "Usage: flea admin [email]")
			return 2
		}
		email := ""
		if len(args) == 2 {
			email = args[1]
		} else {
			var err error
			email, err = setup.Prompt(ctx, os.Stdin, out, "Site owner email address")
			if err != nil {
				return fail(err)
			}
		}
		if err := ctx.Err(); err != nil {
			return fail(err)
		}
		email, added, err := site.AddAdmin(".", email)
		if err != nil {
			return fail(err)
		}
		if added {
			fmt.Fprintf(out, "\nAdded %s to _/flea.config.\n", email)
		} else {
			fmt.Fprintf(out, "\n%s already has owner access in this site's configuration. Nothing changed.\n", email)
		}
		settings, err := mailer.LocalSettings(".")
		if err != nil {
			return fail(err)
		}
		if settings.Validate() != nil {
			fmt.Fprintln(out, "\nEmail needs setting up.")
			if err := setup.Email(ctx, ".", os.Stdin, out); err != nil {
				return fail(err)
			}
		}
		fmt.Fprintln(out, "\nSign in at /admin. Local access updates automatically.\nFor the live site, commit and run flea deploy.")
		return 0
	case "version", "--version":
		if len(args) == 2 && (args[1] == "--help" || args[1] == "-h") {
			fmt.Fprintln(out, "Usage: flea version\n\nShow the installed version.")
			return 0
		}
		fmt.Fprintln(out, "FleaCMS "+buildinfo.Version)
		return 0
	case "deploy", "https":
		var s setup.DeploySettings
		flags := flag.NewFlagSet(args[0], flag.ContinueOnError)
		flags.SetOutput(errOut)
		flags.Usage = func() {
			fmt.Fprintf(flags.Output(), "Usage: flea %s [options]\n\nConfigure the current Git branch's destination and publish it.\nUse -configure to edit saved settings in _/deploy.json.\nDeployment requires a clean, committed Git repository.\n\nOptions\n", args[0])
			flags.PrintDefaults()
		}
		if len(args) == 2 && (args[1] == "--help" || args[1] == "-h") {
			flags.SetOutput(out)
		}
		flags.StringVar(&s.Target, "target", s.Target, "SSH target (user@host or SSH alias)")
		flags.StringVar(&s.Domain, "domain", s.Domain, "comma-separated hostnames or HTTP IPv4 addresses")
		flags.StringVar(&s.Name, "name", s.Name, "stable site identifier (suggested from directory and branch)")
		flags.BoolVar(&s.HTTPS, "https", s.HTTPS, "enable automatic HTTPS")
		flags.StringVar(&s.Email, "email", s.Email, "Let's Encrypt contact email")
		flags.BoolVar(&s.AgreeTOS, "agree-tos", s.AgreeTOS, "accept Let's Encrypt terms: https://letsencrypt.org/repository/")
		flags.BoolVar(&s.WWWRedirect, "www-redirect", true, "redirect WWW to the primary domain")
		configure := flags.Bool("configure", false, "confirm and save changes to this branch's deployment settings")
		expectedIPs := flags.String("expected-ips", "", "expected DNS destination IPs for NAT or proxy setups")
		if err := flags.Parse(args[1:]); err != nil {
			if err == flag.ErrHelp {
				return 0
			}
			return 2
		}
		if flags.NArg() != 0 {
			fmt.Fprint(errOut, usage)
			return 2
		}
		provided := map[string]bool{}
		flags.Visit(func(f *flag.Flag) {
			if f.Name != "configure" {
				provided[f.Name] = true
			}
		})
		if *expectedIPs != "" {
			for _, ip := range strings.Split(*expectedIPs, ",") {
				s.ExpectedIPs = append(s.ExpectedIPs, strings.TrimSpace(ip))
			}
		}
		s, err := setup.SelectDeployment(ctx, ".", s, provided, *configure, args[0] == "https", os.Stdin, out)
		if err != nil {
			return fail(err)
		}
		if err := setup.Deploy(ctx, ".", s, args[0] == "https", os.Stdin, out, errOut); err != nil {
			return fail(err)
		}
		return 0
	case "setup":
		flags := flag.NewFlagSet("setup", flag.ContinueOnError)
		flags.SetOutput(errOut)
		flags.Usage = func() {
			fmt.Fprintln(flags.Output(), "Usage: flea setup\n\nInstall Flea on your server over SSH.\nYou will be asked for the server address.")
		}
		if len(args) == 2 && (args[1] == "--help" || args[1] == "-h") {
			flags.SetOutput(out)
		}
		if err := flags.Parse(args[1:]); err != nil {
			if err == flag.ErrHelp {
				return 0
			}
			return 2
		}
		if flags.NArg() != 0 {
			fmt.Fprint(errOut, usage)
			return 2
		}
		if err := setup.Run(ctx, os.Stdin, out, errOut); err != nil {
			return fail(err)
		}
		return 0
	case "help", "-h", "--help":
		fmt.Fprint(out, usage)
		return 0
	case "new":
		skipGit := false
		var names []string
		for _, arg := range args[1:] {
			switch arg {
			case "--skip-git", "-skip-git":
				skipGit = true
			case "-h", "--help":
				fmt.Fprintln(out, "Usage: flea new <site-name> [--skip-git]\n\nCreates a site and starts a Git repository on main.\nUse --skip-git to leave Git setup to you.")
				return 0
			default:
				names = append(names, arg)
			}
		}
		if len(names) != 1 || !regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9_-]*$`).MatchString(names[0]) {
			fmt.Fprint(errOut, usage)
			return 2
		}
		name := names[0]
		if err := Create(name); err != nil {
			return fail(err)
		}
		fmt.Fprintf(out, "\nCreated your site: %s\n", name)
		if !skipGit {
			gitPath, err := exec.LookPath("git")
			if err != nil {
				fmt.Fprintln(out, "\nGit isn't installed. You can develop your site now.\nInstall Git before deploying, then run git init -b main in your site folder.")
			} else {
				cmd := exec.CommandContext(ctx, gitPath, "init", "--quiet", "-b", "main")
				cmd.Dir = name
				// Do not let inherited Git location overrides initialize another repo.
				for _, entry := range os.Environ() {
					key, _, _ := strings.Cut(entry, "=")
					if key != "GIT_DIR" && key != "GIT_WORK_TREE" && key != "GIT_COMMON_DIR" && key != "GIT_INDEX_FILE" {
						cmd.Env = append(cmd.Env, entry)
					}
				}
				if err := cmd.Run(); err != nil {
					return fail(fmt.Errorf("Your site was created, but Git setup failed.\n\nOpen the %s folder and run:\n  git init -b main", name))
				}
				fmt.Fprintln(out, "\nGit is ready on the main branch. No commit has been made yet.")
			}
		}
		fmt.Fprintf(out, "\nStart developing:\n  cd %s\n  flea dev\n\nWhen you're ready to publish:\n  flea deploy\n\nFleaCMS will save your settings and offer to commit your changes before deploying.\n\n", name)
		return 0
	case "status":
		if len(args) == 2 && (args[1] == "--help" || args[1] == "-h") {
			fmt.Fprintln(out, "Usage: flea status\n\nCheck the current site's configuration, templates, and content.\nRun this command from your site folder.")
			return 0
		}
		if len(args) != 1 {
			fmt.Fprint(errOut, usage)
			return 2
		}
		p, err := site.Load(".")
		if err != nil {
			return fail(err)
		}
		content, _ := p.Content.Index()
		fmt.Fprintf(out, "Valid local project: %s\n%d pages, %d assets, %d fleas\n", p.Site["name"], len(p.Pages), len(p.Assets), len(p.Fleas)+len(content))
		if len(p.Admin.Emails) > 0 {
			fmt.Fprintf(out, "Local sessions: %s (%d-day idle expiry)\n", filepath.Join(p.Root, "_", "sessions.json"), p.Admin.SessionIdleDays)
		}
		for _, warning := range p.Warnings {
			fmt.Fprintf(errOut, "Warning: %s\n", warning)
		}
		for _, route := range p.Routes() {
			fmt.Fprintln(out, " ", route)
		}
		return 0
	case "dev":
		flags := flag.NewFlagSet("dev", flag.ContinueOnError)
		flags.SetOutput(errOut)
		flags.Usage = func() {
			fmt.Fprintln(flags.Output(), "Usage: flea dev [options]\n\nStart the local development server from your site folder.\nChanges reload automatically.\n\nOptions")
			flags.PrintDefaults()
		}
		if len(args) == 2 && (args[1] == "--help" || args[1] == "-h") {
			flags.SetOutput(out)
		}
		port := flags.Int("port", 3000, "local HTTP port (0 selects an available port)")
		if err := flags.Parse(args[1:]); err != nil {
			if err == flag.ErrHelp {
				return 0
			}
			return 2
		}
		if flags.NArg() != 0 || *port < 0 || *port > 65535 {
			fmt.Fprint(errOut, usage)
			return 2
		}
		p, err := site.LoadEditing(ctx, ".")
		if err != nil {
			return fail(err)
		}
		if err := prepareLocalSessions(ctx, p); err != nil {
			return fail(err)
		}
		mailSettings, err := mailer.LocalSettings(".")
		if err != nil {
			return fail(err)
		}
		var sender mailer.Sender
		if mailSettings.Validate() == nil {
			sender = mailer.NewResend(mailSettings)
		}
		listener, err := net.Listen("tcp", fmt.Sprintf("127.0.0.1:%d", *port))
		if err != nil {
			return fail(err)
		}
		logger := log.New(errOut, "", 0)
		if p.FleasUpdated {
			logger.Print("Fleas updated in _/fleas.json.")
		}
		for _, warning := range p.Warnings {
			logger.Printf("Warning: %s", warning)
		}
		handler := site.NewServer(p, logger)
		var httpHandler http.Handler = handler
		{
			dashboard := devAdmin(p.Root, fmt.Sprint(p.Site["name"]), handler, logger)
			httpHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if editor.IsPath(r.URL.Path) {
					dashboard.ServeHTTP(w, r)
				} else {
					handler.ServeHTTP(w, r)
				}
			})
			if len(p.Admin.Emails) > 0 && sender == nil {
				logger.Print("Owner sign-in needs email setup. Run flea email.")
			}
		}
		if p.Contact.To != "" {
			form := contact.New(p.Contact, p.Root, sender, &contact.Budget{}, true, logger)
			next := httpHandler
			httpHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path == contact.Path {
					form.ServeHTTP(w, r)
				} else {
					next.ServeHTTP(w, r)
				}
			})
			if sender == nil {
				logger.Print("Contact forms need email setup. Run flea email, then restart dev.")
			}
		}
		server := &http.Server{Handler: httpHandler, ReadHeaderTimeout: 5 * time.Second, ReadTimeout: 30 * time.Second, WriteTimeout: 30 * time.Second, IdleTimeout: 30 * time.Second}
		watchCtx, cancel := context.WithCancel(ctx)
		defer cancel()
		watchDone := make(chan error, 1)
		go func() { watchDone <- handler.Watch(watchCtx, p.Root, logger) }()
		done := make(chan error, 1)
		go func() { done <- server.Serve(listener) }()
		fmt.Fprintf(out, "\nFleaCMS running at http://%s\nWatching templates, assets, configuration and fleas.\n\nPress Ctrl-C to stop.\n", listener.Addr())
		select {
		case err := <-watchDone:
			server.Close()
			<-done
			if err != nil {
				return fail(err)
			}
			return 0
		case <-ctx.Done():
			shutdownCtx, stop := context.WithTimeout(context.Background(), 5*time.Second)
			defer stop()
			if err := server.Shutdown(shutdownCtx); err != nil {
				server.Close()
			}
			<-done
		case err := <-done:
			cancel()
			<-watchDone
			if err != http.ErrServerClosed {
				return fail(err)
			}
			return 0
		}
		cancel()
		<-watchDone
		fmt.Fprintln(out, "Stopped.")
		return 0
	default:
		fmt.Fprint(errOut, usage)
		return 2
	}
}

// Create reserves a new directory first and never overwrites an existing target.
func Create(target string) error {
	if err := os.Mkdir(target, 0755); err != nil {
		return err
	}
	err := fs.WalkDir(templates.Files, ".", func(name string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() {
			return nil
		}
		body, err := templates.Files.ReadFile(name)
		if err != nil {
			return err
		}
		dest := filepath.Join(target, filepath.FromSlash(name))
		if err := os.MkdirAll(filepath.Dir(dest), 0755); err != nil {
			return err
		}
		f, err := os.OpenFile(dest, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0644)
		if err != nil {
			return err
		}
		_, err = f.Write(body)
		closeErr := f.Close()
		if err != nil {
			return err
		}
		if closeErr != nil {
			return closeErr
		}
		return nil
	})
	if err != nil {
		return err
	}
	_, err = site.Load(target)
	return err
}
