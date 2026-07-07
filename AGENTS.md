# Agent notes for this repo

## How to query the production DB on k3s

There is no postgres pod in the `linglow` namespace, and the `linglow` app
image does not bundle a `psql` client. The actual Postgres instance backing
`linglow_unified` lives in the `english` namespace, as `deploy/english-postgres`
(role `english`, default db `english`); `linglow_unified` is a sibling database
on that same instance.

Working command (read-only queries):

```bash
kubectl -n english exec -it deploy/english-postgres -- psql -U english -d linglow_unified -c "<SQL>"
```

Notes:
- `-U postgres` does NOT work — that role doesn't exist on this instance.
- Do not port-forward, spin up an ad-hoc `pg-client` pod, or look for
  Vault/ExternalSecret-based access — none of that is set up for this project;
  the `kubectl exec` into `english-postgres` above is the working path.

## Mascot name: always "Lumi"

The mascot is always written **`Lumi`** — in every language, including Russian UI
text. Never transliterate it (not «Луми», not «Люми»). It stays Latin-script "Lumi"
in glosses, prompts, help text and locale strings.

## Icons: Lucide only, no emoji

The project's icon pack is **[Lucide](https://lucide.dev)** (ISC license — free for
commercial use). Both icon components — `webapp/src/components/linglow/LgIcon.vue`
(the Linglow UI set) and `webapp/src/components/Icon.vue` (older/admin) — hold
hand-inlined Lucide glyphs (24×24 viewBox, stroke-width 2, round caps). When you need
a new icon, copy the SVG paths from lucide.dev into `LgIcon.vue` as a new
`name === '<lucide-name>'` branch (keep the Lucide name), then reference it via
`<LgIcon name="…">`. Don't pull in another icon library.

Never use emoji (colorful pictographs like 📍🎨💬) anywhere in the product UI —
neither in Vue templates nor in data/config that renders into the UI. Plain
typographic glyphs (arrows →↑↓, ✓, ×) are fine. This keeps the interface visually
consistent across themes and platforms.

## Spanish verb-forms binaries in the `linglow` image

Dockerfile ships `/app/import_spanish_verbs`, `/app/backfill_word_verb_links`,
`/app/sync_verb_training_json` (+ Jehle CSV and bundled `verb_forms` JSON).

- **One-time per DB** (`linglow_unified` after merge): `import_spanish_verbs` (Jehle + haber) + `backfill_word_verb_links`.
- **Every rollout**: initContainer runs `sync_verb_training_json` (GitOps: `devops-time-host/apps/linglow/base/deployment.yaml`).
- Operator runbook: `devops-time-host/apps/linglow/RELEASE_K3S.md` §2.7.

## OpenRouter proxy in k3s

OpenRouter-compatible requests can use a dedicated SOCKS5 proxy via
`OPENROUTER_SOCKS5_PROXY` (also accepted as `AI_SOCKS5_PROXY` for the main AI
client). This is intentionally scoped to OpenRouter HTTP clients instead of
global `HTTP_PROXY`/`HTTPS_PROXY`, so Postgres/Redis/Kubernetes-internal traffic
stays direct.

Runtime coverage:
- main AI/chat-completions client (`AI_URL`);
- training/backfill CLI commands that create the same AI client;
- TTS OpenRouter provider (`TTS_BASE_URL`) through `TTS_OPENROUTER_SOCKS5_PROXY`
  or fallback to `OPENROUTER_SOCKS5_PROXY`;
- Speaking evaluator (`SPEAKING_EVAL_BASE_URL`) through
  `SPEAKING_OPENROUTER_SOCKS5_PROXY` or fallback to `OPENROUTER_SOCKS5_PROXY`.

k3s ConfigMaps for `english`, `spanish`, and `linglow` set
`OPENROUTER_SOCKS5_PROXY: "51.254.98.124:1080"`.

## Imported Claude Cowork project instructions
