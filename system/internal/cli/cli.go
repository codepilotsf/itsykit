package cli

import (
	"context"
	"errors"
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
	"strings"
	"time"

	"itsy/internal/buildinfo"
	"itsy/internal/contact"
	"itsy/internal/devaccess"
	"itsy/internal/mailer"
	"itsy/internal/mailpreview"
	"itsy/internal/selfupdate"
	"itsy/internal/setup"
	"itsy/internal/site"
	"itsy/internal/sitefiles"
	"itsy/templates"
)

const usage = `_________
 ItsyWeb
‾‾‾‾‾‾‾‾‾
Small websites. Plain HTML. Editable content.

Usage
  itsy <command>

Build your site
  new <directory>  Create a new site
  reset            Reset this local site to a clean start
  dev              Start the local development server
  status           Show local and remote site status
  admin [email]    Add an admin who can edit content
  email            Set up email delivery

Publish your site
  pull             Bring live snips and media into your local site
  setup            Install ItsyWeb on your server
  deploy           Publish the current Git branch
  rollback         Restore previous code, keeping live content
  backup           Download a complete site backup
  restore          Restore a site backup after confirmation
  update-server    Update the server without resetting sites
  https            Configure HTTPS for your site
  server           List or remove sites on a server over SSH

Other commands
  update           Update this CLI and its bundled servers
  version          Show the installed version
  help             Show this help

Get started
  itsy new my-site
  cd my-site
  itsy dev

Command details
  itsy <command> --help

`

const serverUsage = `Usage: itsy server list [ssh-target]
       itsy server remove [ssh-target] [site-name]
       itsy server indexing [ssh-target] <hostname> <noindex|allow>

Inspect or configure sites installed on an ItsyWeb server over SSH. No local site,
Git repository or saved deployment settings are required. Omit the SSH target to
be asked for the server login.

  list     List sites with their HTTPS and HTTP domains and redirects.
  remove   Unregister one site's hostnames. Keeps every release, content file,
           media file, session and certificate on disk. Omit the site name to
           choose from the managed sites installed on the server. Requires typing
           the exact site name. Removal restarts the shared service, so other
           sites are briefly unavailable; when no HTTPS hostnames remain it
           disables automatic HTTPS while cached certificates stay on disk.
           Retained content is not restored automatically by redeploying the
           same name; back up or pull the site first if a later deployment must
           keep it.

  indexing Set one registered hostname to noindex, nofollow or add no restriction.
           Keeps website files unchanged. Requires typing the exact hostname and
           restarts the shared service. Policies survive deploy, sync and rollback.

All commands reject servers running an older ItsyWeb protocol; update the
server first. Nothing is deleted or pruned.

`

// Run returns 0 on success, 1 on project/runtime failure, 2 on usage errors.
func Run(ctx context.Context, args []string, out, errOut io.Writer) int {
	if len(args) == 0 {
		fmt.Fprint(out, usage)
		return 0
	}
	fail := func(err error) int { fmt.Fprintln(errOut, "itsy:", err); return 1 }
	switch args[0] {
	case "update":
		if len(args) == 2 && (args[1] == "--help" || args[1] == "-h") {
			fmt.Fprint(out, "Usage: itsy update\n\nDownload and verify the latest CLI from fleacms.com.\nUpdates this executable and its bundled servers, without changing remote servers.\nNo site folder is required.\n\n")
			return 0
		}
		if len(args) != 1 {
			fmt.Fprintln(errOut, "Usage: itsy update")
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
			fmt.Fprint(out, "Usage: itsy update-server [--recover]\n\nUpdate this branch's server using the bundled executable. All its sites\nrestart briefly. Configuration, credentials and site data are preserved.\nNo automatic format migration or service-unit changes.\n\n")
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
		terminal := false
		if file, ok := out.(*os.File); ok {
			terminal = devTerminal(ctx, file)
		}
		if err := setup.Upgrade(ctx, ".", *recover, os.Stdin, out, errOut, terminal); err != nil {
			return fail(err)
		}
		return 0
	case "backup", "restore":
		if len(args) == 2 && (args[1] == "--help" || args[1] == "-h") {
			fmt.Fprintf(out, "Usage: itsy %s <backup-file>\n\nUse this branch's saved destination. Backups include code, snips and theme,\nnot credentials, certificates or sessions. Restore replaces live site data\nafter confirmation. Backup refuses to overwrite an existing file.\n\n", args[0])
			return 0
		}
		if len(args) != 2 || strings.HasPrefix(args[1], "-") {
			fmt.Fprintf(errOut, "Usage: itsy %s <backup-file>\n", args[0])
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
			fmt.Fprint(out, "Usage: itsy rollback [--list] [release-id]\n\nRestore previous code for this branch. Keep live snips. Styles come from the selected release.\nWithout an ID, use the preceding release. Confirmation is required.\n\n")
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
				fmt.Fprintln(errOut, "Usage: itsy email preview <sign-in|contact>")
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
			fmt.Fprintln(out, "Usage: itsy email\n       itsy email preview <sign-in|contact>\n\nSet up Resend delivery, or print a sample HTML email without sending it.\nTemplates live in mailers/. Setup saves credentials in .env.\nContact settings need a dev restart.")
			fmt.Fprintln(out)
			return 0
		}
		if len(args) != 1 {
			fmt.Fprintln(errOut, "Usage: itsy email")
			return 2
		}
		if err := setup.Email(ctx, ".", os.Stdin, out); err != nil {
			return fail(err)
		}
		return 0
	case "admin":
		if len(args) == 2 && (args[1] == "--help" || args[1] == "-h") {
			fmt.Fprintln(out, "Usage: itsy admin [email]\n\nAdd an admin and set up email if needed. Local access updates automatically.\nView or remove admins in admin.emails in itsy.config.json.\nCommit and deploy to apply live changes. No invitation is sent.")
			fmt.Fprintln(out)
			return 0
		}
		if len(args) > 2 || (len(args) == 2 && strings.HasPrefix(args[1], "-")) {
			fmt.Fprintln(errOut, "Usage: itsy admin [email]")
			return 2
		}
		email := ""
		if len(args) == 2 {
			email = args[1]
		} else {
			var err error
			email, err = setup.Prompt(ctx, os.Stdin, out, "Admin email address")
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
			fmt.Fprintf(out, "\nAdded %s to %s.\n", email, sitefiles.ResolveAt(".", sitefiles.ConfigName))
		} else {
			fmt.Fprintf(out, "\n%s already has admin access in this site's configuration. Nothing changed.\n", email)
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
		fmt.Fprintln(out, "\nSign in at /admin. Local access updates automatically.\nFor the live site, commit and run itsy deploy.")
		return 0
	case "version", "--version":
		if len(args) == 2 && (args[1] == "--help" || args[1] == "-h") {
			fmt.Fprintln(out, "Usage: itsy version\n\nShow the installed version.")
			fmt.Fprintln(out)
			return 0
		}
		fmt.Fprintln(out, "ItsyWeb "+buildinfo.Version)
		return 0
	case "pull":
		flags := flag.NewFlagSet("pull", flag.ContinueOnError)
		flags.SetOutput(errOut)
		flags.Usage = func() {
			fmt.Fprintln(flags.Output(), "Usage: itsy pull\n\nPull live snips and media into your local site. Stop dev first.\nReview and commit changes afterwards. Configuration and sessions stay local.")
			fmt.Fprintln(flags.Output())
		}
		if len(args) == 2 && (args[1] == "--help" || args[1] == "-h") {
			flags.SetOutput(out)
		}
		if e := flags.Parse(args[1:]); e != nil {
			if e == flag.ErrHelp {
				return 0
			}
			return 2
		}
		if flags.NArg() != 0 {
			return fail(fmt.Errorf("unexpected pull arguments"))
		}
		if e := setup.Pull(ctx, ".", os.Stdin, out, errOut); e != nil {
			return fail(e)
		}
		return 0
	case "deploy", "https":
		var s setup.DeploySettings
		flags := flag.NewFlagSet(args[0], flag.ContinueOnError)
		flags.SetOutput(errOut)
		flags.Usage = func() {
			fmt.Fprintf(flags.Output(), "Usage: itsy %s [options]\n\nConfigure the current Git branch's destination and publish it.\nUse -configure to edit saved settings in itsy.config.json.\nDeployment requires a clean, committed Git repository.\n\nOptions\n", args[0])
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
			if f.Name != "configure" && f.Name != "with-snips" {
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
		if args[0] == "https" && *withSnips {
			return fail(fmt.Errorf("use these options with itsy deploy"))
		}
		if err := setup.Deploy(ctx, ".", s, args[0] == "https", os.Stdin, out, errOut); err != nil {
			return fail(err)
		}
		return 0
	case "setup":
		flags := flag.NewFlagSet("setup", flag.ContinueOnError)
		flags.SetOutput(errOut)
		flags.Usage = func() {
			fmt.Fprintln(flags.Output(), "Usage: itsy setup\n\nInstall ItsyWeb on your server over SSH.\nYou will be asked for the server address.")
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
	case "server":
		return runServer(ctx, args[1:], os.Stdin, out, errOut)
	case "help", "-h", "--help":
		fmt.Fprint(out, usage)
		return 0
	case "new":
		skipGit := false
		starter := false
		var names []string
		pathsOnly := false
		for _, arg := range args[1:] {
			if pathsOnly {
				names = append(names, arg)
				continue
			}
			switch arg {
			case "--":
				pathsOnly = true
			case "--starter", "-s":
				starter = true
			case "--skip-git", "-skip-git":
				skipGit = true
			case "-h", "--help":
				fmt.Fprintln(out, "Usage: itsy new [--starter|-s] <directory> [--skip-git]\n\nCreates a clean site with an initial Git commit on main.\nUse . for the current directory. Existing directories must be empty.\nUse --starter or -s for example pages and team content.\nUse --skip-git to leave Git setup to you.")
				fmt.Fprintln(out)
				return 0
			default:
				if strings.HasPrefix(arg, "-") {
					fmt.Fprintf(errOut, "Unknown option: %s\nRun itsy new --help for usage.\n", arg)
					return 2
				}
				names = append(names, arg)
			}
		}
		if len(names) != 1 || strings.TrimSpace(names[0]) == "" {
			fmt.Fprint(errOut, usage)
			return 2
		}
		name := filepath.Clean(names[0])
		if err := createSite(name, starter); err != nil {
			return fail(err)
		}
		fmt.Fprintf(out, "\nCreated your site: %s\n", name)
		if !skipGit {
			gitPath, err := exec.LookPath("git")
			if err != nil {
				fmt.Fprintln(out, "\nGit isn't installed. You can develop your site now.\nInstall Git before deploying, then run git init -b main in your site folder.")
			} else {
				runGit := func(args ...string) ([]byte, error) {
					cmd := exec.CommandContext(ctx, gitPath, args...)
					cmd.Dir = name
					cmd.Stdin = os.Stdin
					// Do not let inherited Git location overrides touch another repo.
					for _, entry := range os.Environ() {
						key, _, _ := strings.Cut(entry, "=")
						if key != "GIT_DIR" && key != "GIT_WORK_TREE" && key != "GIT_COMMON_DIR" && key != "GIT_INDEX_FILE" {
							cmd.Env = append(cmd.Env, entry)
						}
					}
					return cmd.CombinedOutput()
				}
				if _, err := runGit("init", "--quiet", "-b", "main"); err != nil {
					return fail(fmt.Errorf("Your site was created, but Git setup failed.\n\nOpen the %s folder and run:\n  git init -b main", name))
				}
				gitOutput, gitErr := runGit("add", "--", ".")
				if gitErr == nil {
					gitOutput, gitErr = runGit("commit", "--quiet", "-m", "Initial site")
				}
				if gitErr != nil {
					fmt.Fprintln(out, "\nYour site is ready, but its initial Git commit could not be created.")
					if detail := strings.TrimSpace(string(gitOutput)); detail != "" {
						fmt.Fprintln(out, "\n"+detail)
					}
					fmt.Fprintf(out, "\nResolve the Git issue above, then run inside %s:\n  git add .\n  git commit -m \"Initial site\"\n", name)
				} else {
					fmt.Fprintln(out, "\nGit is ready on main with an initial commit.\nTo connect a remote repository, run inside your site:\n  git remote add origin <repository-url>\n  git push -u origin main")
				}
			}
		}
		if name == "." {
			fmt.Fprint(out, "\nStart developing:\n  itsy dev\n\n")
		} else {
			// Quote paths containing spaces or shell metacharacters for copying.
			cdPath := name
			if strings.ContainsAny(cdPath, " \t\n\r'\"`$;&|<>(){}[]*?!~#\\") || strings.HasPrefix(cdPath, "-") {
				if strings.HasPrefix(cdPath, "-") {
					cdPath = "./" + cdPath
				}
				cdPath = "'" + strings.ReplaceAll(cdPath, "'", "'\"'\"'") + "'"
			}
			fmt.Fprintf(out, "\nStart developing:\n  cd %s\n  itsy dev\n\n", cdPath)
		}
		return 0
	case "reset":
		if len(args) == 2 && (args[1] == "--help" || args[1] == "-h") {
			fmt.Fprintln(out, "Usage: itsy reset\n\nReset this local site to a clean layout and homepage.\nReplaces content, uploads, customizations and settings after confirmation.\nKeeps Git history and a private sibling backup. Never changes the remote site.")
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
			fmt.Fprintln(out, "Usage: itsy status\n\nShow local validation and Git state, then check this branch's saved remote.\nCompares deployed code, editable content and theme over SSH.\nRun this command from your site folder.")
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
		terminal := false
		if file, ok := out.(*os.File); ok {
			terminal = devTerminal(ctx, file)
		}
		if err = setup.Status(ctx, p, os.Stdin, out, errOut, terminal); err != nil {
			return fail(err)
		}
		return 0
	case "dev":
		flags := flag.NewFlagSet("dev", flag.ContinueOnError)
		flags.SetOutput(errOut)
		flags.Usage = func() {
			fmt.Fprintln(flags.Output(), "Usage: itsy dev [options]\n\nStart the local development server from your site folder.\nAsks for a site name once, then offers optional admin setup.\nNew sites preview email in your browser without sending.\nChanges reload automatically.\nStarts at port 3000 and tries the next port if it is occupied.\n\nOptions")
			flags.PrintDefaults()
			fmt.Fprintln(flags.Output())
		}
		if len(args) == 2 && (args[1] == "--help" || args[1] == "-h") {
			flags.SetOutput(out)
		}
		port := flags.Int("port", 3000, "fixed HTTP port (0 selects an available port)")
		var lan bool
		flags.BoolVar(&lan, "lan", false, "enable full development access on a trusted private/Tailscale IPv4 network (HTTP)")
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
		automaticPort := true
		flags.Visit(func(f *flag.Flag) {
			if f.Name == "port" {
				automaticPort = false
			}
		})
		ctx, unlock, err := beginDevelopment(ctx, ".", out)
		if err != nil {
			return fail(err)
		}
		defer unlock()
		p, err := site.LoadEditing(ctx, ".")
		if err != nil {
			return fail(err)
		}
		changed, offered, err := prepareDevOwner(ctx, p, os.Stdin, out, devTerminal(ctx, os.Stdin))
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
		var lanAddresses, lanHostnames []string
		if lan {
			lanAddresses, err = devLANAddresses()
			if err != nil {
				return fail(err)
			}
			lanHostnames, err = devLANHostnames()
			if err != nil {
				return fail(err)
			}
			bindHost = "0.0.0.0"
		}
		listener, err := listenDev(ctx, bindHost, *port, automaticPort, out, (&net.ListenConfig{}).Listen)
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
			logger.Printf("Snips updated in %s.", sitefiles.ResolveAt(p.Root, "itsy/snips.json"))
		}
		for _, warning := range p.Warnings {
			logger.Printf("Warning: %s", warning)
		}
		handler := site.NewServer(p, logger)
		stopApps := handler.StartApps(ctx, sender, &contact.Budget{})
		defer stopApps()
		var httpHandler http.Handler = handler
		{
			httpHandler = devAdmin(p.Root, fmt.Sprint(p.App["site_name"]), handler, logger, previewSender)
			if len(p.Admin.Emails) == 0 && !offered {
				logger.Print("Admin sign-in isn't set up. Run itsy admin when you're ready.")
			}
			if len(p.Admin.Emails) > 0 && sender == nil {
				logger.Print("Admin sign-in needs email setup. Run itsy email.")
			}
		}
		if _, contactEnabled := p.Apps["contact"]; contactEnabled && sender == nil {
			logger.Print("Contact forms need email setup. Run itsy email, then restart dev.")
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
			httpHandler = devaccess.WrapLAN(httpHandler, lanAddresses, lanHostnames, listener.Addr().(*net.TCPAddr).Port)
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
			for _, name := range lanHostnames {
				fmt.Fprintf(out, "\nHostname http://%s:%d\n", name, listener.Addr().(*net.TCPAddr).Port)
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
			if errors.Is(context.Cause(ctx), errDevReplaced) {
				// Do not release the site lock while an old request could still
				// be saving content. The replacement caller has its own timeout.
				shutdownCtx = context.Background()
			}
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

// runServer handles server-wide site administration over an SSH target. The
// target and site name may be given explicitly; otherwise they are prompted for.
func runServer(ctx context.Context, args []string, in io.Reader, out, errOut io.Writer) int {
	if len(args) == 0 || args[0] == "--help" || args[0] == "-h" {
		fmt.Fprint(out, serverUsage)
		return 0
	}
	switch args[0] {
	case "indexing":
		if len(args) == 2 && (args[1] == "--help" || args[1] == "-h") {
			fmt.Fprint(out, serverUsage)
			return 0
		}
		if len(args) != 3 && len(args) != 4 {
			fmt.Fprintln(errOut, "Usage: itsy server indexing [ssh-target] <hostname> <noindex|allow>")
			return 2
		}
		target := ""
		if len(args) == 4 {
			target = args[1]
		}
		if err := setup.ServerIndexing(ctx, target, args[len(args)-2], args[len(args)-1], in, out, errOut); err != nil {
			fmt.Fprintln(errOut, "itsy:", err)
			return 1
		}
		return 0
	case "list":
		if len(args) == 2 && (args[1] == "--help" || args[1] == "-h") {
			fmt.Fprint(out, serverUsage)
			return 0
		}
		if len(args) > 2 || (len(args) == 2 && strings.HasPrefix(args[1], "-")) {
			fmt.Fprintln(errOut, "Usage: itsy server list [ssh-target]")
			return 2
		}
		target := ""
		if len(args) == 2 {
			target = args[1]
		}
		if err := setup.ServerList(ctx, target, in, out, errOut); err != nil {
			fmt.Fprintln(errOut, "itsy:", err)
			return 1
		}
		return 0
	case "remove":
		if len(args) >= 2 && (args[1] == "--help" || args[1] == "-h") {
			fmt.Fprint(out, serverUsage)
			return 0
		}
		if len(args) == 3 && (args[2] == "--help" || args[2] == "-h") {
			fmt.Fprint(out, serverUsage)
			return 0
		}
		if len(args) > 3 || (len(args) >= 2 && strings.HasPrefix(args[1], "-")) || (len(args) == 3 && strings.HasPrefix(args[2], "-")) {
			fmt.Fprintln(errOut, "Usage: itsy server remove [ssh-target] [site-name]")
			return 2
		}
		target, name := "", ""
		if len(args) >= 2 {
			target = args[1]
		}
		if len(args) == 3 {
			name = args[2]
		}
		if err := setup.ServerRemove(ctx, target, name, in, out, errOut); err != nil {
			fmt.Fprintln(errOut, "itsy:", err)
			return 1
		}
		return 0
	default:
		fmt.Fprint(errOut, serverUsage)
		return 2
	}
}

// Create makes a site in a new or empty directory without overwriting files.
func Create(target string) error {
	return createSite(target, false)
}

func createSite(target string, starter bool) error {
	target = filepath.Clean(target)
	if err := os.Mkdir(target, 0755); err != nil && !errors.Is(err, fs.ErrExist) {
		return err
	}
	info, err := os.Lstat(target)
	if err != nil {
		return err
	}
	if !info.IsDir() {
		return fmt.Errorf("%s is not a directory. Choose an empty directory", target)
	}
	boundary, err := os.OpenRoot(target)
	if err != nil {
		return err
	}
	defer boundary.Close()
	dir, err := boundary.Open(".")
	if err != nil {
		return err
	}
	entries, readErr := dir.Readdirnames(1)
	closeErr := dir.Close()
	if readErr != nil && readErr != io.EOF {
		return readErr
	}
	if closeErr != nil {
		return closeErr
	}
	if len(entries) != 0 {
		return fmt.Errorf("Directory %s isn't empty. Choose an empty directory", target)
	}
	err = fs.WalkDir(templates.Files, ".", func(name string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() {
			return nil
		}
		if !starter && (name == "site/team.html" || strings.HasPrefix(name, "site/contact/") || name == "site/test.html") {
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
			case "itsy/snips.json":
				body, err = templates.MinimalFiles.ReadFile("minimal/snips.json")
			case "site/_navigation.html":
				body = []byte("<a href=\"/\">Home</a>\n")
			case "itsy.config.json":
				body = []byte("{\n  \"app\": {},\n  \"redirects\": {},\n  \"dev\": {\"email_preview\": true},\n  \"admin\": {\"emails\": [], \"session_idle_days\": 30}\n}\n")
			}
			if err != nil {
				return err
			}
		}
		dest := filepath.FromSlash(name)
		if err := boundary.MkdirAll(filepath.Dir(dest), 0755); err != nil {
			return err
		}
		f, err := boundary.OpenFile(dest, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0644)
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
		err = boundary.Mkdir(filepath.Join("itsy", "media"), 0755)
	}
	return err
}
