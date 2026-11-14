# Product

<!-- impeccable:product-schema 1 -->

## Platform

web

## Users

Developers building simple websites with HTML and CSS, including developers working with coding agents. The owner confirmed both audiences. They want familiar files and a simple end-to-end system rather than a collection of tools to assemble.

## Product Purpose

ItsyWeb adds server-rendered layouts, partials, variables, conditions, loops and clean URLs to ordinary HTML files. Its CLI includes local development, SSH server setup and deployment, automatic HTTPS, backups and recovery.

The owner describes the experience as very close to creating a static website in plain HTML, with useful superpowers. This is the intended simplicity and positioning, not a claim that ItsyWeb generates a static output directory. The home page must give both ordinary HTML and the complete install-to-HTTPS workflow substantial emphasis.

## Operating Context

Developers edit ordinary files in a text editor and preview them with `itsy dev`. The tutorial builds a small bookshop. Deployment uses a committed Git branch, SSH access, an Ubuntu server and a domain. HTTPS requires the documented DNS and network prerequisites.

Generated projects include AGENTS.md and portable Markdown skills for site authoring and common additions. Developers can use their coding agent in the project and review the resulting files and preview before publishing. ItsyWeb sets up its software on a supplied Ubuntu host; it does not provision a cloud virtual machine.

## Capabilities and Constraints

No frontend build step or Node runtime is required. The public test binaries support Apple Silicon macOS and x86-64 Linux. Contact, Blog and managed apps have their own documentation. The public website is an existing ItsyWeb site with plain HTML, CSS, JSON content and Prewind utilities. Ordinary page copy and metadata live in HTML; Blog retains its own admin editor, storage and cover uploads. Preserve URLs, Contact CSRF and saved deployment mappings. No deployment or release is authorized by this design task.

The integrated source targets the page-editing-removal candidate and protocol 30. The public 0.21.0 installer uses protocol 29 and does not supply this contract. A matching release, current native acceptance and explicit server-upgrade authorization remain prerequisites to rollout.

## Brand Commitments

Retain the approved ItsyWeb logo and mark. The existing navy and ice palette is the starting identity for this home-page composition exploration. The user authorized rethinking messaging and section order, then requested Impeccable and Mobbin in conjunction with the framework study.

## Evidence on Hand

Current site templates, theme.css, shared CSS and itsy/snips.json; the Getting Started bookshop tutorial; the framework study at .impeccable/research/homepage-study.html; approved vector brand assets in site/images/. External reference sites supply presentation evidence, not claims about ItsyWeb. Do not fabricate customers, testimonials, usage statistics or performance claims.
