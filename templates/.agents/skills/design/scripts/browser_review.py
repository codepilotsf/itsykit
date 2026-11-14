#!/usr/bin/env python3
"""Package and serve intentional static reviews without changing access policy."""
import argparse
import functools
from html.parser import HTMLParser
import http.client
from http.server import SimpleHTTPRequestHandler, ThreadingHTTPServer
import ipaddress
import json
import os
from pathlib import Path
import re
import secrets
import shlex
import socket
import subprocess
import sys
import threading
import time
from urllib.parse import unquote, urlsplit
import zipfile


class ReviewError(ValueError):
    pass


def export_root(project, export):
    project = Path(project).resolve(strict=True)
    path = Path(export)
    if not path.is_absolute():
        path = project / path
    if any(p.is_symlink() for p in (path, *path.parents)):
        raise ReviewError("Export paths must not contain symlinks")
    root = path.resolve(strict=True)
    relative = root.relative_to(project)
    if relative.parts[:2] != (".impeccable", "exports") or len(relative.parts) < 3:
        raise ReviewError("Export must be a dedicated .impeccable/exports/<review> directory")
    if not root.is_dir() or not (root / "index.html").is_file():
        raise ReviewError("Export needs an index.html")
    return project, root


class References(HTMLParser):
    def __init__(self):
        super().__init__()
        self.urls = []

    def handle_starttag(self, tag, attrs):
        for name, value in attrs:
            if value is None:
                continue
            if tag == "base" or name == "srcdoc":
                raise ReviewError("Remove base/srcdoc assumptions from the portable export")
            if name in {"href", "xlink:href", "src", "poster", "action", "formaction", "data"}:
                self.urls.append(value)
            if name == "srcset":
                # Embedded data images belong in src, avoiding ambiguous data commas.
                self.urls.extend(item.strip().split()[0] for item in value.split(",") if item.strip())


def check_url(value, source, root):
    value = unquote(value.strip())
    if value.startswith("data:") or value.startswith("#") or not value:
        return
    parsed = urlsplit(value)
    if parsed.scheme or parsed.netloc or value.startswith(("/", "\\")) or "\\" in value:
        raise ReviewError(f"Nonportable URL in {source.name}: {value[:100]}")
    if not parsed.path:
        return
    target = (source.parent / parsed.path).resolve()
    if not target.is_relative_to(root):
        raise ReviewError(f"URL escapes export in {source.name}: {value[:100]}")
    if target.is_dir():
        target /= "index.html"
    if not target.is_file() and target != root / "review.zip":
        raise ReviewError(f"Missing asset in {source.name}: {value[:100]}")


def inventory(root):
    files = []
    for path in sorted(root.rglob("*")):
        relative = path.relative_to(root)
        if path.is_symlink() or any(p.startswith(".") for p in relative.parts):
            raise ReviewError(f"Hidden files and symlinks are not review artifacts: {relative}")
        if path.is_dir():
            continue
        if not path.is_file():
            raise ReviewError(f"Unsupported review artifact: {relative}")
        if relative == Path("review.zip"):
            continue
        if path.suffix.lower() in {".html", ".htm", ".css", ".js", ".svg"}:
            body = path.read_text(encoding="utf-8")
            if re.search(r"(?:localhost|127\.0\.0\.1|\[::1\]|file://|/home/|/Users/|[A-Za-z]:\\)", body, re.I):
                raise ReviewError(f"Loopback or private filesystem reference in {relative}")
            urls = re.findall(r"url\(\s*['\"]?([^)'\"]+)", body, re.I)
            urls += re.findall(r"@import\s+['\"]([^'\"]+)", body, re.I)
            if path.suffix.lower() in {".html", ".htm", ".svg"}:
                parser = References()
                parser.feed(body)
                urls += parser.urls
            for url in urls:
                check_url(url, path, root)
        files.append(path)
    return files


def pack(root):
    files = inventory(root)
    archive = root / "review.zip"
    if archive.is_symlink():
        raise ReviewError("Archive must not be a symlink")
    # Stage outside the served tree so a reader never downloads a partial ZIP.
    temporary = root.parent / f".{root.name}.{secrets.token_hex(8)}.zip"
    try:
        with zipfile.ZipFile(temporary, "w", zipfile.ZIP_DEFLATED) as bundle:
            for path in files:
                bundle.write(path, path.relative_to(root))
        temporary.replace(archive)
    finally:
        temporary.unlink(missing_ok=True)
    return files + [archive]


def narrow_ip(value):
    address = ipaddress.ip_address(value)
    networks = ("127.0.0.0/8", "10.0.0.0/8", "172.16.0.0/12", "192.168.0.0/16",
                "100.64.0.0/10", "::1/128", "fc00::/7")
    if not any(address in ipaddress.ip_network(n) for n in networks):
        raise ReviewError("Bind must be a specific loopback or existing private-network IP")
    return str(address)


def route(args):
    if args.port is not None and not 0 <= args.port <= 65535:
        raise ReviewError("Port must be between 0 and 65535")
    if args.route and (args.bind or args.owner_base_url):
        raise ReviewError("Use --route OR an explicit --bind and --owner-base-url")
    if args.route == "local":
        return "127.0.0.1", "127.0.0.1", args.port or 0
    if args.route == "tailscale":
        try:
            result = subprocess.run(["tailscale", "status", "--json"], check=True,
                                    capture_output=True, text=True, timeout=10)
            status = json.loads(result.stdout)
        except (OSError, subprocess.SubprocessError, ValueError) as exc:
            raise ReviewError("Cannot discover existing Tailscale route; prepare ZIP and resolve transport") from exc
        node = status.get("Self", {})
        addresses = node.get("TailscaleIPs", [])
        host = node.get("DNSName", "").rstrip(".")
        if status.get("BackendState") != "Running" or not addresses or not host:
            raise ReviewError("No running Tailscale route; prepare ZIP and resolve transport")
        bind = next((a for a in addresses if ":" not in a), addresses[0])
        return narrow_ip(bind), host, args.port or 0
    if not args.bind or not args.owner_base_url:
        raise ReviewError("Choose --route local/tailscale or supply --bind and --owner-base-url")
    bind = narrow_ip(args.bind)
    parsed = urlsplit(args.owner_base_url)
    if (parsed.scheme != "http" or not parsed.hostname or parsed.username or parsed.password
            or parsed.path not in {"", "/"} or parsed.query or parsed.fragment):
        raise ReviewError("Owner base URL must be a direct HTTP origin without credentials or a path")
    port = parsed.port or 80
    if args.port is not None and args.port != port:
        raise ReviewError("Explicit port must match the owner base URL")
    try:
        owner_loopback = ipaddress.ip_address(parsed.hostname).is_loopback
    except ValueError:
        owner_loopback = parsed.hostname.rstrip(".").lower() == "localhost"
    if not ipaddress.ip_address(bind).is_loopback and owner_loopback:
        raise ReviewError("A remote private bind cannot use an owner loopback URL")
    return bind, parsed.hostname, port


def origin(host, port):
    return f"http://{'[' + host + ']' if ':' in host else host}:{port}/"


def read_state(path):
    return json.loads(path.read_text())


def write_state(path, state):
    temporary = path.with_name(path.name + ".tmp")
    descriptor = os.open(temporary, os.O_WRONLY | os.O_CREAT | os.O_EXCL, 0o600)
    try:
        with os.fdopen(descriptor, "w") as output:
            output.write(json.dumps(state, indent=2) + "\n")
        temporary.replace(path)
    finally:
        temporary.unlink(missing_ok=True)


def control(state, method="GET"):
    connection = http.client.HTTPConnection(state["bind"], state["port"], timeout=2)
    try:
        connection.request(method, "/__review_control", headers={"Authorization": "Bearer " + state["token"]})
        response = connection.getresponse()
        data = json.loads(response.read())
        if response.status != 200 or data.get("token") != state["token"]:
            raise ReviewError("Server ownership check failed; no process was stopped")
        return data
    finally:
        connection.close()


def run_server(state_path):
    state = read_state(state_path)
    root = Path(state["served_root"])
    allowed = {str(p.relative_to(root)) for p in pack(root)}

    class Handler(SimpleHTTPRequestHandler):
        def log_message(self, *args):
            pass

        def authenticated(self):
            return secrets.compare_digest(self.headers.get("Authorization", ""), "Bearer " + state["token"])

        def do_GET(self):
            if urlsplit(self.path).path == "/__review_control":
                if not self.authenticated():
                    self.send_error(403)
                    return
                body = json.dumps({"token": state["token"], "pid": os.getpid()}).encode()
                self.send_response(200)
                self.send_header("Content-Type", "application/json")
                self.send_header("Content-Length", str(len(body)))
                self.end_headers()
                self.wfile.write(body)
                return
            super().do_GET()

        def do_POST(self):
            if self.path != "/__review_control" or not self.authenticated():
                self.send_error(403)
                return
            self.do_GET()
            threading.Thread(target=self.server.shutdown, daemon=True).start()

        def send_head(self):
            requested = Path(self.translate_path(self.path))
            if requested.is_dir():
                requested /= "index.html"
            # Recheck real paths on every request, including symlinks introduced later.
            if (not requested.resolve().is_relative_to(root)
                    or any(p.is_symlink() for p in (requested, *requested.parents))
                    or str(requested.relative_to(root)) not in allowed):
                self.send_error(404)
                return None
            return super().send_head()

        def list_directory(self, path):
            self.send_error(404)
            return None

        def end_headers(self):
            self.send_header("Cache-Control", "no-store")
            self.send_header("X-Content-Type-Options", "nosniff")
            super().end_headers()

    class Server(ThreadingHTTPServer):
        address_family = socket.AF_INET6 if ":" in state["bind"] else socket.AF_INET
        daemon_threads = True

    with Server((state["bind"], state["port"]), functools.partial(Handler, directory=str(root))) as server:
        state.update(pid=os.getpid(), port=server.server_port, status="running")
        state["internal_preview_url"] = origin(state["bind"], server.server_port)
        state["owner_review_url"] = origin(state["owner_host"], server.server_port)
        state["download_url"] = state["owner_review_url"] + "review.zip"
        write_state(state_path, state)
        server.serve_forever(poll_interval=0.1)
    state["status"] = "stopped"
    write_state(state_path, state)


def serve(args, project, root):
    bind, host, port = route(args)
    state_path = root.with_name(root.name + ".server.json")
    owner = args.owner or str(root.relative_to(project))
    expected = {"served_root": str(root), "project_root": str(project), "owner": owner,
                "bind": bind, "owner_host": host, "requested_port": port}
    if state_path.exists():
        if state_path.is_symlink():
            raise ReviewError("Server record must not be a symlink")
        previous = read_state(state_path)
        if previous.get("status") != "stopped":
            if any(previous.get(k) != v for k, v in expected.items()):
                raise ReviewError("Existing server record belongs to another owner/configuration; inspect it first")
            control(previous)
            return previous
    pack(root)
    log_path = root.with_name(root.name + ".server.log")
    if log_path.is_symlink():
        raise ReviewError("Server log must not be a symlink")
    state = {**expected, "port": port, "status": "starting", "token": secrets.token_urlsafe(32),
             "stop_command": shlex.join([sys.executable, str(Path(__file__).resolve()), "stop", "--state", str(state_path)])}
    write_state(state_path, state)
    descriptor = os.open(log_path, os.O_WRONLY | os.O_CREAT | os.O_APPEND, 0o600)
    with os.fdopen(descriptor, "a") as log:
        process = subprocess.Popen([sys.executable, str(Path(__file__).resolve()), "_run", "--state", str(state_path)],
                                   stdin=subprocess.DEVNULL, stdout=log, stderr=log, start_new_session=True)
    for _ in range(100):
        if process.poll() is not None:
            state.update(status="stopped", startup_error=True)
            write_state(state_path, state)
            raise ReviewError(f"Review server failed; inspect {log_path}")
        current = read_state(state_path)
        if current.get("status") == "running":
            control(current)
            return current
        time.sleep(0.05)
    process.terminate()  # This Popen object is the child just started, never a recorded PID.
    process.wait(timeout=5)
    state.update(status="stopped", startup_error=True)
    write_state(state_path, state)
    raise ReviewError(f"Review server startup timed out; inspect {log_path}")


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    sub = parser.add_subparsers(dest="command", required=True)
    for name in ("pack", "serve"):
        command = sub.add_parser(name)
        command.add_argument("--project-root", required=True)
        command.add_argument("--export-dir", required=True)
        if name == "serve":
            command.add_argument("--route", choices=("local", "tailscale"))
            command.add_argument("--bind")
            command.add_argument("--owner-base-url")
            command.add_argument("--port", type=int)
            command.add_argument("--owner")
    for name in ("stop", "_run"):
        command = sub.add_parser(name)
        command.add_argument("--state", type=Path, required=True)
    args = parser.parse_args()
    try:
        if args.command == "_run":
            run_server(args.state)
            return
        if args.command == "stop":
            state = read_state(args.state)
            if state.get("status") != "stopped":
                control(state, "POST")
            print(json.dumps({"status": "stop-requested", "owner": state["owner"]}))
            return
        project, root = export_root(args.project_root, args.export_dir)
        if args.command == "pack":
            pack(root)
            print(json.dumps({"archive": str(root / "review.zip"), "entry": "index.html"}))
        else:
            state = serve(args, project, root)
            print(json.dumps({k: v for k, v in state.items() if k != "token"}, indent=2))
    except (ReviewError, OSError, ValueError, http.client.HTTPException) as exc:
        parser.exit(1, f"browser-review: {exc}\n")


if __name__ == "__main__":
    main()
