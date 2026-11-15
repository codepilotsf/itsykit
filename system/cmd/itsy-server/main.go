package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"itsy/internal/admin"
	"itsy/internal/buildinfo"
	"itsy/internal/host"
	"itsy/internal/release"
	"itsy/internal/syncstate"
)

func main() {
	validate := flag.Bool("validate", false, "validate all installed sites without starting the server")
	upgrade := flag.String("upgrade", "", "install verified server executable (root only)")
	upgradeSHA := flag.String("sha256", "", "expected server checksum")
	recoverUpgrade := flag.Bool("recover-upgrade", false, "recover interrupted executable upgrade (root only)")
	backupSite := flag.Bool("backup", false, "export site backup (root only)")
	restoreSite := flag.String("restore", "", "restore a site backup (root only)")
	expectedState := flag.String("expected-state", "", "reviewed live-state digest")
	history := flag.Bool("history", false, "list site release history (root only)")
	rollback := flag.String("rollback", "", "restore retained code, preserving live content (root only)")
	expectedRelease := flag.String("expected-release", "", "reviewed active release ID")
	export := flag.Bool("export-state", false, "export Blog content state (root only)")
	listSites := flag.Bool("list-sites", false, "list managed sites (read-only)")
	siteInfo := flag.String("site-info", "", "print reviewed state for one managed site (read-only)")
	setIndexing := flag.String("set-indexing", "", "set one registered hostname indexing policy (root only)")
	indexingMode := flag.String("indexing-mode", "", "noindex or allow")
	removeSite := flag.String("remove-site", "", "unregister a managed site, retaining its files (root only)")
	planPath := flag.String("sync-plan", "", "reviewed sync plan accompanying deployment")
	version := flag.Bool("version", false, "print ItsyWeb release version")
	config := flag.String("config", "/etc/flea/server.json", "server JSON configuration")
	check := flag.Bool("check", false, "check configuration and running server readiness")
	socketActivation := flag.Bool("socket-activation", false, "require HTTP and HTTPS listening sockets from systemd")
	protocol := flag.Bool("protocol", false, "print deployment protocol version")
	mailStatus := flag.Bool("mail-status", false, "check server email configuration (root only)")
	configureMail := flag.Bool("configure-mail", false, "save email settings from stdin (root only)")
	archive := flag.String("deploy", "", "activate a staged site archive (root only)")
	name := flag.String("name", "", "managed site name")
	domain := flag.String("domain", "", "comma-separated site hostnames")
	https := flag.Bool("https", false, "enable managed HTTPS for the hostnames")
	email := flag.String("email", "", "ACME contact email")
	agree := flag.Bool("agree-tos", false, "accept Let's Encrypt terms of service")
	recover := flag.Bool("recover", false, "restore an interrupted activation (root only)")
	dns := flag.Bool("check-dns", false, "verify DNS from this server before uploading")
	expected := flag.String("expected-ips", "", "expected DNS destination IPs (default: local interfaces)")
	branch := flag.String("branch", "", "source Git branch")
	commit := flag.String("commit", "", "source Git commit")
	canonical := flag.String("canonical", "", "primary domain for WWW redirect")
	flag.Parse()
	if *version {
		fmt.Println("ItsyWeb " + buildinfo.Version)
		return
	}
	logger := log.New(os.Stderr, "", 0)
	if flag.NArg() != 0 {
		logger.Print("unexpected arguments")
		os.Exit(2)
	}
	if *protocol {
		fmt.Println(release.Protocol)
		return
	}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	var err error
	if *validate {
		_, _, err = host.Load(*config, nil)
	} else if *upgrade != "" || *recoverUpgrade {
		if os.Geteuid() != 0 || *config != "/etc/flea/server.json" {
			logger.Print("Upgrade requires root and the installed config.")
			os.Exit(1)
		}
		if *recoverUpgrade {
			err = admin.SystemUpgrader().Recover(ctx)
		} else {
			err = admin.SystemUpgrader().Upgrade(ctx, *upgrade, *upgradeSHA)
		}
		if err == nil {
			fmt.Println("Server ready. Site data and configuration preserved.")
		}
	} else if *backupSite || *restoreSite != "" {
		if os.Geteuid() != 0 || *config != "/etc/flea/server.json" {
			logger.Print("Recovery requires root and the installed config.")
			os.Exit(1)
		}
		if *backupSite {
			err = admin.System().Backup(*name, *branch, os.Stdout)
		} else {
			var id string
			id, err = admin.System().Restore(ctx, *name, *branch, *restoreSite, *expectedState)
			if err == nil {
				fmt.Println("Restored site backup. Release " + id)
			}
		}
	} else if *history || *rollback != "" {
		if os.Geteuid() != 0 || *config != "/etc/flea/server.json" {
			logger.Print("Recovery requires root and the installed config.")
			os.Exit(1)
		}
		if *history {
			var records []admin.ReleaseRecord
			records, err = admin.System().History(*name, *branch)
			if err == nil {
				err = json.NewEncoder(os.Stdout).Encode(records)
			}
		} else {
			var id string
			id, err = admin.System().Rollback(ctx, *name, *branch, *rollback, *expectedRelease)
			if err == nil {
				fmt.Println("Rolled back code. Live Blog content preserved; CSS comes from the selected release. Release " + id)
			}
		}
	} else if *export {
		if os.Geteuid() != 0 || *config != "/etc/flea/server.json" {
			logger.Print("State export requires root and the installed config.")
			os.Exit(1)
		}
		var snapshot syncstate.Snapshot
		snapshot, err = admin.System().Snapshot(*name, *branch)
		if err == nil {
			err = json.NewEncoder(os.Stdout).Encode(snapshot)
		}
	} else if *mailStatus || *configureMail {
		if os.Geteuid() != 0 || *config != "/etc/flea/server.json" {
			logger.Print("Email configuration requires root and the installed server config.")
			os.Exit(1)
		}
		if *mailStatus {
			if admin.MailReady(*config) {
				fmt.Println("FLEA_MAIL=ready")
			} else {
				fmt.Println("FLEA_MAIL=missing")
			}
		} else {
			err = admin.System().ConfigureMail(os.Stdin)
		}
	} else if *dns {
		var ips []string
		if *expected != "" {
			ips = strings.Split(*expected, ",")
		}
		err = admin.CheckDNS(ctx, admin.Domains(*domain), ips, os.Stdout)
	} else if *archive != "" || *https || *recover {
		if os.Geteuid() != 0 || *config != "/etc/flea/server.json" {
			logger.Print("administration requires root and the installed server config")
			os.Exit(1)
		}
		m := admin.System()
		if *recover {
			err = m.Recover(ctx)
		} else {
			if *branch == "" || *commit == "" {
				logger.Print("deployment requires source branch and commit")
				os.Exit(1)
			}
			var revision string
			var plan *syncstate.Plan
			if *planPath != "" {
				f, e := os.Open(*planPath)
				if e != nil {
					logger.Print(e)
					os.Exit(1)
				}
				d := json.NewDecoder(io.LimitReader(f, 8<<20))
				d.DisallowUnknownFields()
				e = d.Decode(&plan)
				if e == nil && d.Decode(new(any)) != io.EOF {
					e = fmt.Errorf("invalid sync plan")
				}
				f.Close()
				if e != nil || plan == nil {
					logger.Print("Invalid sync plan")
					os.Exit(1)
				}
			}
			revision, err = m.Apply(ctx, admin.Options{Sync: plan, Archive: *archive, Name: *name, Domains: admin.Domains(*domain), HTTPS: *https, Email: *email, AgreeTOS: *agree, Canonical: *canonical, Branch: *branch, Commit: *commit})
			if err == nil {
				fmt.Printf("Activated release %s\n", revision)
			}
		}
	} else if *listSites || *siteInfo != "" {
		if *config != "/etc/flea/server.json" {
			logger.Print("Site administration uses the installed server config.")
			os.Exit(1)
		}
		if *listSites {
			var sites []admin.SiteSummary
			sites, err = admin.System().Sites()
			if err == nil {
				err = json.NewEncoder(os.Stdout).Encode(sites)
			}
		} else {
			var info admin.SiteSummary
			info, err = admin.System().SiteInfo(*siteInfo)
			if err == nil {
				err = json.NewEncoder(os.Stdout).Encode(info)
			}
		}
	} else if *setIndexing != "" || *indexingMode != "" {
		if os.Geteuid() != 0 || *config != "/etc/flea/server.json" {
			logger.Print("Indexing configuration requires root and the installed server config.")
			os.Exit(1)
		}
		err = admin.System().SetIndexing(ctx, *setIndexing, *indexingMode, *expectedState)
		if err == nil {
			fmt.Printf("Indexing policy for %s: %s\n", *setIndexing, *indexingMode)
		}
	} else if *removeSite != "" {
		if os.Geteuid() != 0 || *config != "/etc/flea/server.json" {
			logger.Print("Site removal requires root and the installed server config.")
			os.Exit(1)
		}
		err = admin.System().RemoveSite(ctx, *removeSite, *expectedState)
		if err == nil {
			fmt.Println("Removed site from routing. Releases, content, media, sessions and certificates retained.")
		}
	} else if *check {
		err = host.Check(ctx, *config)
	} else if *socketActivation {
		err = host.RunSocketActivated(ctx, *config, logger)
	} else {
		err = host.Run(ctx, *config, logger)
	}
	if err != nil {
		logger.Print(err)
		os.Exit(1)
	}
}
