package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"flea/internal/admin"
	"flea/internal/buildinfo"
	"flea/internal/host"
	"flea/internal/release"
)

func main() {
	version := flag.Bool("version", false, "print Flea CMS release version")
	config := flag.String("config", "/etc/flea/server.json", "server JSON configuration")
	check := flag.Bool("check", false, "check configuration and running server readiness")
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
		fmt.Println("Flea CMS " + buildinfo.Version)
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
	if *mailStatus || *configureMail {
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
			revision, err = m.Apply(ctx, admin.Options{Archive: *archive, Name: *name, Domains: admin.Domains(*domain), HTTPS: *https, Email: *email, AgreeTOS: *agree, Canonical: *canonical, Branch: *branch, Commit: *commit})
			if err == nil {
				fmt.Printf("Activated release %s\n", revision)
			}
		}
	} else if *check {
		err = host.Check(ctx, *config)
	} else {
		err = host.Run(ctx, *config, logger)
	}
	if err != nil {
		logger.Print(err)
		os.Exit(1)
	}
}
