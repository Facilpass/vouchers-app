# vouchers-app

CDN multi-tenant de vouchers FacilPass. Upload admin UI via path livre `/<tenant>/<evento>/<categoria>/<arquivo>.webp`. Embed em email/landing via `<img src>`.

## Stack

- Go 1.23 stdlib + `golang.org/x/crypto/bcrypt` + CGO `github.com/chai2010/webp`
- Alpine 3.20 multi-stage Docker (~18 MB)
- Traefik path-based routing: `/admin/*` auth + demais público

## Deploy

- Staging: `vouchers.facilpass.prod.local` (Dokploy homelab)
- Produção: `vouchers.facilpass.com.br` (Dokploy Hostinger 187.77.224.44)

## Features

- Upload-only admin UI single-user (bcrypt + session HMAC)
- Path livre multi-tenant `<tenant>/<evento>/<categoria>/[...]` até 10 níveis
- Auto-sanitize filename (Unicode NFD, whitelist, collapse) + path (lowercase/hífen)
- Auto-convert JPEG/PNG → WebP lossy q=85 (WebP passthrough)
- Timestamp versioning: nunca sobrescreve, renomeia com `_{YYYYMMDD-HHMM}`
- URL pública limpa + tag `<img>` pronta pra copiar

## Segurança

- HTTPS obrigatório (mkcert staging / LE prod) + HSTS
- CSP, X-Frame, X-Content-Type-Options nosniff
- Rate limit público 60 req/s + admin 10 req/s (Traefik)
- Rate limit login 5 tentativas / 15 min (app)
- Path traversal blindado, MIME sniff, CSRF HMAC
- Container non-root, read-only FS, cap_drop ALL, mem/cpu/pids limits
- Volume external dedicado, zero compartilhamento

## Docs

Spec: `docs/superpowers/specs/2026-04-22-vouchers-facilpass-design.md` (repo `homelab`).
Plano: `docs/superpowers/plans/2026-04-22-vouchers-facilpass-execution.md` (repo `homelab`).
