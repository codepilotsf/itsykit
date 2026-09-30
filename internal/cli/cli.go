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
	"flea/internal/devaccess"
	"flea/internal/mailer"
	"flea/internal/mailpreview"
	"flea/internal/selfupdate"
	"flea/internal/setup"
	"flea/internal/site"
	"flea/templates"
)

const usage = `_________
 FleaCMS
‾‾‾‾‾‾‾‾‾
Small websites. Plain HTML. Editable content.

Usage
  flea <command>

Build your site
  new <site-name>  Create a new site
  reset            Reset this local site to a clean start
  dev              Start the local development server
  status           Check your site for errors
  admin [email]    Add an owner who can edit content
  email            Set up email delivery

Publish your site
  pull             Bring live snips and theme into your local site
  setup            Install Flea on your server
  deploy           Publish the current Git branch
  rollback         Restore previous code, keeping live content
  backup           Download a complete site backup
  restore          Restore a site backup after confirmation
  update-server    Update the server without resetting sites
  https            Configure HTTPS for your site

Other commands
  update           Update this CLI and its bundled servers
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
	case "update":
		if len(args) == 2 && (args[1] == "--help" || args[1] == "-h") {
			fmt.Fprint(out, "Usage: flea update\n\nDownload and verify the latest CLI from fleacms.com.\nUpdates this executable and its bundled servers, without changing remote servers.\nNo site folder is required.\n\n")
			return 0
		}
		if len(args) != 1 {
			fmt.Fprintln(errOut, "Usage: flea update")
			return 2
		}
		if err := selfupdate.Run(ctx, buildinfo.Version, out); err != nil {
			return fail(err)
		}
		return 0
	case "update-server":
		flags := flag.NewFlagSet("update-server", flag.ContinueOnError)
		flags.SetOutput(errOut)
		recover := flags.Bool("recover", false, "restore executable after an interrupted upgrade")
		flags.Usage = func() {
			fmt.Fprint(out, "Usage: flea update-server [--recover]\n\nUpdate this branch's server using the bundled executable. All its sites\nrestart briefly. Configuration, credentials and site data are preserved.\nNo automatic format migration or service-unit changes.\n\n")
		}
		if err := flags.Parse(args[1:]); err != nil {
			if err == flag.ErrHelp {
				return 0
			}
			return 2
		}
		if flags.NArg() != 0 {
			return 2
		}
		if err := setup.Upgrade(ctx, ".", *recover, os.Stdin, out, errOut); err != nil {
			return fail(err)
		}
		return 0
	case "backup", "restore":
		if len(args) == 2 && (args[1] == "--help" || args[1] == "-h") {
			fmt.Fprintf(out, "Usage: flea %s <backup-file>\n\nUse this branch's saved destination. Backups include code, snips and theme,\nnot credentials, certificates or sessions. Restore replaces live site data\nafter confirmation. Backup refuses to overwrite an existing file.\n\n", args[0])
			return 0
		}
		if len(args) != 2 || strings.HasPrefix(args[1], "-") {
			fmt.Fprintf(errOut, "Usage: flea %s <backup-file>\n", args[0])
			return 2
		}
		var err error
		if args[0] == "backup" {
			err = setup.Backup(ctx, ".", args[1], os.Stdin, out, errOut)
		} else {
			err = setup.Restore(ctx, ".", args[1], os.Stdin, out, errOut)
		}
		if err != nil {
			return fail(err)
		}
		return 0
	case "rollback":
		flags := flag.NewFlagSet("rollback", flag.ContinueOnError)
		flags.SetOutput(errOut)
		list := flags.Bool("list", false, "list retained releases")
		flags.Usage = func() {
			fmt.Fprint(out, "Usage: flea rollback [--list] [release-id]\n\nRestore previous code for this branch. Keep live snips and theme.\nWithout an ID, use the preceding release. Confirmation is required.\n\n")
		}
		if err := flags.Parse(args[1:]); err != nil {
			if err == flag.ErrHelp {
				return 0
			}
			return 2
		}
		if flags.NArg() > 1 {
			return 2
		}
		target := ""
		if flags.NArg() == 1 {
			target = flags.Arg(0)
		}
		if err := setup.Rollback(ctx, ".", target, *list, os.Stdin, out, errOut); err != nil {
			return fail(err)
		}
		return 0
	case "email":
		if len(args) >= 2 && args[1] == "preview" {
			if len(args) != 3 || (args[2] != "sign-in" && args[2] != "contact") {
				fmt.Fprintln(errOut, "Usage: flea email preview <sign-in|contact>")
				return 2
			}
			p, err := site.Load(".")
			if err != nil {
				return fail(err)
			}
			message, err := p.RenderEmail(args[2], map[string]string{"code": "123456", "host": "localhost", "name": "Joe Appleseed", "address": "joe@example.com", "message": "Hello! I'd like to learn more about your services."})
			if err != nil {
				return fail(err)
			}
			fmt.Fprint(out, message.HTML)
			return 0
		}
		if len(args) == 2 && (args[1] == "--help" || args[1] == "-h") {
			fmt.Fprintln(out, "Usage: flea email\n       flea email preview <sign-in|contact>\n\nSet up Resend delivery, or print a sample HTML email without sending it.\nTemplates live in mailers/. Setup saves credentials in .env.\nContact settings need a dev restart.")
			fmt.Fprintln(out)
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
			fmt.Fprintln(out, "Usage: flea admin [email]\n\nAdd an owner and set up email if needed. Local access updates automatically.\nView or remove owners in admin.emails in .flea.config.json.\nCommit and deploy to apply live changes. No invitation is sent.")
			fmt.Fprintln(out)
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
			fmt.Fprintf(out, "\nAdded %s to .flea.config.json.\n", email)
		} else {
			fmt.Fprintf(out, "\n%s already has owner access in this site's configuration. Nothing changed.\n", email)
		}
		settings, err := mailer.LocalSettings(".")
		if err != nil {
			return fail(err)
		}
		dev, err := site.ReadDevSettings(".")
		if err != nil {
			return fail(err)
		}
		if !dev.EmailPreview && settings.Validate() != nil {
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
			fmt.Fprintln(out)
			return 0
		}
		fmt.Fprintln(out, "FleaCMS "+buildinfo.Version)
		return 0
	case "pull":
		flags := flag.NewFlagSet("pull", flag.ContinueOnError)
		flags.SetOutput(errOut)
		flags.Usage = func() {
			fmt.Fprintln(flags.Output(), "Usage: flea pull [--snips-only | --theme-only]\n\nPull live snips and theme into your local site. Stop dev first.\nReview and commit changes afterwards. Configuration and sessions stay local.")
			fmt.Fprintln(flags.Output())
		}
		if len(args) == 2 && (args[1] == "--help" || args[1] == "-h") {
			flags.SetOutput(out)
		}
		snipsOnly := flags.Bool("snips-only", false, "pull only snips")
		themeOnly := flags.Bool("theme-only", false, "pull only theme")
		if e := flags.Parse(args[1:]); e != nil {
			if e == flag.ErrHelp {
				return 0
			}
			return 2
		}
		if flags.NArg() != 0 {
			return fail(fmt.Errorf("unexpected pull arguments"))
		}
		if e := setup.Pull(ctx, ".", *snipsOnly, *themeOnly, os.Stdin, out, errOut); e != nil {
			return fail(e)
		}
		return 0
	case "deploy", "https":
		var s setup.DeploySettings
		flags := flag.NewFlagSet(args[0], flag.ContinueOnError)
		flags.SetOutput(errOut)
		flags.Usage = func() {
			fmt.Fprintf(flags.Output(), "Usage: flea %s [options]\n\nConfigure the current Git branch's destination and publish it.\nUse -configure to edit saved settings in .flea.config.json.\nDeployment requires a clean, committed Git repository.\n\nOptions\n", args[0])
			flags.PrintDefaults()
			fmt.Fprintln(flags.Output())
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
		withSnips := flags.Bool("with-snips", false, "also publish reviewed snip changes (pull first)")
		withTheme := flags.Bool("with-theme", false, "also publish reviewed theme changes (pull first)")
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
			if f.Name != "configure" && f.Name != "with-snips" && f.Name != "with-theme" {
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
		s.WithSnips = *withSnips
		s.WithTheme = *withTheme
		if args[0] == "https" && (*withSnips || *withTheme) {
			return fail(fmt.Errorf("use these options with flea deploy"))
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
			fmt.Fprintln(flags.Output())
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
		starter := false
		var names []string
		for _, arg := range args[1:] {
			switch arg {
			case "--starter", "-s":
				starter = true
			case "--skip-git", "-skip-git":
				skipGit = true
			case "-h", "--help":
				fmt.Fprintln(out, "Usage: flea new [--starter|-s] <site-name> [--skip-git]\n\nCreates a clean site and starts a Git repository on main.\nUse --starter or -s for example pages and team content.\nUse --skip-git to leave Git setup to you.")
				fmt.Fprintln(out)
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
		if err := createSite(name, starter); err != nil {
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
		fmt.Fprintf(out, "\nStart developing:\n  cd %s\n  flea dev\n\n", name)
		return 0
	case "reset":
		if len(args) == 2 && (args[1] == "--help" || args[1] == "-h") {
			fmt.Fprintln(out, "Usage: flea reset\n\nReset this local site to a clean layout and homepage.\nReplaces content, uploads, customizations and settings after confirmation.\nKeeps Git history and a private sibling backup. Never changes the remote site.")
			fmt.Fprintln(out)
			return 0
		}
		if len(args) != 1 {
			return 2
		}
		root, err := os.Getwd()
		if err != nil {
			return fail(err)
		}
		if err = resetSite(root, os.Stdin, out); err != nil {
			return fail(err)
		}
		return 0
	case "status":
		if len(args) == 2 && (args[1] == "--help" || args[1] == "-h") {
			fmt.Fprintln(out, "Usage: flea status\n\nCheck the current site's configuration, templates, and content.\nRun this command from your site folder.")
			fmt.Fprintln(out)
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
		fmt.Fprintf(out, "Valid local project: %s\n%d pages, %d assets, %d snips\n", p.App["site_name"], len(p.Pages), len(p.Assets), len(p.Snips)+len(content))
		if len(p.Admin.Emails) > 0 {
			fmt.Fprintf(out, "Local sessions: %s (%d-day idle expiry)\n", filepath.Join(p.Root, "flea", "runtime", "sessions.json"), p.Admin.SessionIdleDays)
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
			fmt.Fprintln(flags.Output(), "Usage: flea dev [options]\n\nStart the local development server from your site folder.\nAsks for a site name once, then offers optional owner setup.\nNew sites preview email in your browser without sending.\nChanges reload automatically.\n\nOptions")
			flags.PrintDefaults()
			fmt.Fprintln(flags.Output())
		}
		if len(args) == 2 && (args[1] == "--help" || args[1] == "-h") {
			flags.SetOutput(out)
		}
		port := flags.Int("port", 3000, "local HTTP port (0 selects an available port)")
		var lan bool
		flags.BoolVar(&lan, "lan", false, "enable full development access on a trusted private IPv4 network (HTTP)")
		flags.BoolVar(&lan, "l", false, "same as --lan")
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
		unlock, err := site.LockDevelopment(".")
		if err != nil {
			return fail(err)
		}
		defer unlock()
		p, err := site.LoadEditing(ctx, ".")
		if err != nil {
			return fail(err)
		}
		changed, err := prepareDevOwner(ctx, p, os.Stdin, out, devTerminal(ctx, os.Stdin))
		if err != nil {
			return fail(err)
		}
		if changed {
			updated := p.SnipsUpdated
			p, err = site.LoadEditing(ctx, ".")
			if err != nil {
				return fail(err)
			}
			p.SnipsUpdated = p.SnipsUpdated || updated
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
		bindHost := "127.0.0.1"
		var lanAddresses []string
		if lan {
			lanAddresses, err = devLANAddresses()
			if err != nil {
				return fail(err)
			}
			bindHost = "0.0.0.0"
		}
		listener, err := net.Listen("tcp4", net.JoinHostPort(bindHost, fmt.Sprint(*port)))
		if err != nil {
			return fail(err)
		}
		logger := log.New(errOut, "", 0)
		var preview *mailpreview.Preview
		var previewSender mailer.Sender
		if p.Dev.EmailPreview {
			name := mailSettings.Name
			if name == "" {
				name = fmt.Sprint(p.App["site_name"])
			}
			previewOrigin := fmt.Sprintf("http://127.0.0.1:%d", listener.Addr().(*net.TCPAddr).Port)
			preview, err = mailpreview.New(previewOrigin, name, mailSettings.From, logger)
			if err != nil {
				listener.Close()
				return fail(err)
			}
			previewSender, sender = preview, preview
			logger.Print("Email preview is on. Messages open in your browser and are not sent.")
		}
		if p.SnipsUpdated {
			logger.Print("Snips updated in flea/snips.json.")
		}
		for _, warning := range p.Warnings {
			logger.Printf("Warning: %s", warning)
		}
		handler := site.NewServer(p, logger)
		var httpHandler http.Handler = handler
		{
			httpHandler = devAdmin(p.Root, fmt.Sprint(p.App["site_name"]), handler, logger, previewSender)
			if len(p.Admin.Emails) == 0 {
				logger.Print("Owner sign-in isn't set up. Run flea admin when you're ready.")
			}
			if len(p.Admin.Emails) > 0 && sender == nil {
				logger.Print("Owner sign-in needs email setup. Run flea email.")
			}
		}
		if p.Contact.To != "" {
			form := contact.New(p.Contact, p.Root, sender, &contact.Budget{}, true, logger)
			form.RenderEmail = handler.RenderEmail
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
		if preview != nil {
			next := httpHandler
			httpHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if strings.HasPrefix(r.URL.Path, mailpreview.Path) {
					preview.ServeHTTP(w, r)
				} else {
					next.ServeHTTP(w, r)
				}
			})
		}
		if lan {
			httpHandler = devaccess.WrapLAN(httpHandler, lanAddresses, listener.Addr().(*net.TCPAddr).Port)
		}
		server := &http.Server{Handler: httpHandler, ReadHeaderTimeout: 5 * time.Second, ReadTimeout: 30 * time.Second, WriteTimeout: 30 * time.Second, IdleTimeout: 30 * time.Second}
		watchCtx, cancel := context.WithCancel(ctx)
		defer cancel()
		watchDone := make(chan error, 1)
		go func() { watchDone <- handler.Watch(watchCtx, p.Root, logger) }()
		done := make(chan error, 1)
		go func() { done <- server.Serve(listener) }()
		localAddress := fmt.Sprintf("localhost:%d", listener.Addr().(*net.TCPAddr).Port)
		fmt.Fprintf(out, "\nWebsite  http://%s\nSign in  http://%s/admin\n", localAddress, localAddress)
		if lan {
			for _, address := range lanAddresses {
				base := fmt.Sprintf("http://%s:%d", address, listener.Addr().(*net.TCPAddr).Port)
				fmt.Fprintf(out, "\nNetwork  %s\nSign in  %s/admin\n", base, base)
			}
			fmt.Fprintln(out, "\nLAN mode uses unencrypted HTTP. Use only on a trusted network; do not forward this port.")
			if preview != nil {
				fmt.Fprintln(out, "Email previews open on this computer. Use the code on your other device.")
			}
		}
		fmt.Fprintln(out, "\nPress Ctrl-C to stop.")
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
	return createSite(target, false)
}

func createSite(target string, starter bool) error {
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
		if !starter && (name == "site/team.html" || name == "site/contact.html" || name == "site/thank-you.html" || name == "site/test.html") {
			return nil
		}
		body, err := templates.Files.ReadFile(name)
		if err != nil {
			return err
		}
		if !starter {
			switch name {
			case "site/index.html":
				body, err = templates.MinimalFiles.ReadFile("minimal/index.html")
			case "flea/snips.json":
				body, err = templates.MinimalFiles.ReadFile("minimal/snips.json")
			case "site/_navigation.html":
				body = []byte("<a href=\"/\">Home</a>\n")
			case ".flea.config.json":
				body = []byte("{\n  \"app\": {},\n  \"dev\": {\"email_preview\": true},\n  \"admin\": {\"emails\": [], \"session_idle_days\": 30}\n}\n")
			}
			if err != nil {
				return err
			}
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
	if err == nil {
		err = site.IgnoreSessionFiles(target)
	}
	if err == nil {
		err = os.Mkdir(filepath.Join(target, "flea", "media"), 0755)
	}
	return err
}
