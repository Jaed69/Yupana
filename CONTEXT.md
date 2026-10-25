# Yupana — Project Context (handoff)

Read this file first. It captures decisions already made in the research phase so they are not re-litigated without new evidence.

## What Yupana is

A modular platform for Peruvian micro and small businesses (MYPEs) that turns scattered purchase invoices into clean, owned data:

1. **Multichannel intake** — invoices arrive as email XML, PDF, photos of paper (often with the SUNAT QR), via WhatsApp and Telegram.
2. **Validation** — each invoice is checked against SUNAT (Consulta Integrada de Comprobante de Pago).
3. **Deduplication** — the same invoice arriving through several channels is detected (unique key: issuer RUC + type + series + number).
4. **SIRE reconciliation** — received invoices are reconciled against SUNAT's RCE proposal (Registro de Compras Electrónico).
5. **AI expense categorization** — infer what each purchase is for and report spending by category.
6. **Natural-language questions** — e.g. "¿cuánto gasté en insumos de cocina en septiembre vs agosto?"
7. **Data ownership** — the customer can export or sync their data to their own database.

Modules should be valuable alone and better together (Odoo-like), activated per tenant.

- **Pilot customer:** Basílica Café (Cusco). Reported volume: **~200 received invoices per day (~6,000/month)** in every format (XML, PDF, paper with and without QR, WhatsApp). The per-format breakdown is not measured yet.
  - Throughput is trivial for any stack: ~200/day is about one every 7 minutes on average.
  - It matters for external limits: SUNAT validation calls (rate limits unknown), WhatsApp replies (1,000 free service messages/month per number — batch replies, do not answer each invoice individually), and manual review load.
  - At this volume the manual process is a real, measurable pain — record current hours spent as the baseline.
- **Differentiation:** a bot for the business **owner** (not only the accountant) that captures, validates, categorizes and answers questions. Accountants serving many RUCs are the realistic distribution channel.
- **Main competitor:** ContaUno (RAPIFAC) already does photo → extraction → SUNAT validation → SIRE, from S/ 59/month, aimed at accountants, with no AI categorization or conversational bot announced.

## Research

- `docs/research/informe.md` — final report (Spanish): market, architecture, AI layer, SUNAT/legal, costs, 6-phase MVP plan.
- `docs/research/notas/` — the four underlying research notes with sources:
  - `competencia_mercado.md` — competitors, pricing, market size.
  - `ingesta_sunat_legal.md` — WhatsApp/Telegram APIs and costs, SUNAT QR format, SUNAT/SIRE APIs, Peru data-protection law.
  - `ia_categorizacion_consultas.md` — categorization approaches, vector DB vs pgvector, RAG vs text-to-SQL vs typed functions, evals, LLM costs.
  - `arquitectura_modular.md` — modular monolith vs microservices, Odoo modularity, outbox/queues, multi-tenancy with RLS, data export.
  - `evaluacion_backend_go_python.md` — measured Go vs Python evaluation (QR accuracy, memory, library maturity); scripts in `docs/research/bench/`.

## Decisions taken

| Decision | Rationale (see report) |
|---|---|
| Separate **private** repo, deployed on the existing VPS + Traefik from `luciel-platform` | Product/IP separation; security design must not be public; independent CI. Move to its own server before storing multiple customers' SOL credentials. |
| **Modular monolith**, not microservices | 1–2 developers. Split into processes (`api`, `worker`, `bot`) sharing one Postgres, not into services with their own databases. |
| PostgreSQL shared schema + `tenant_id` + **Row Level Security** (`FORCE RLS`, `WITH CHECK`, `SET LOCAL` per request) | Simplest isolation that holds at small scale. |
| Queue/outbox **inside Postgres** (`SKIP LOCKED`), `LISTEN/NOTIFY` only as a wake-up signal | No Redis needed at start. |
| Ingestion priority: **XML > SUNAT QR > PDF text via LLM > photo via vision LLM** | Most exact and cheapest sources first. |
| Categorization in layers: **per-supplier memory → pgvector similarity on tenant history → cheap LLM** for unknowns; user corrections feed memory | Tenant history beats generic models; no dedicated vector DB. |
| **No RAG for numeric questions.** LLM picks among 10–20 **typed query functions** with tested SQL; `tenant_id` always from the session, never from the model | RAG fails at aggregations; free text-to-SQL is unreliable and an injection risk. RAG only for searching document text. |
| Start with **Telegram**, add official **WhatsApp Cloud API** later | Telegram is free; WhatsApp charges service replies since 2026-10-01 (1,000 free/month per number). Unofficial WhatsApp libraries are banned-account risk — do not use. |
| SIRE reconciliation first via **manual upload** of the RCE proposal; API with a secondary SOL user and per-tenant encrypted credentials later | SIRE API requires each customer's SOL credentials — the most sensitive asset in the system. |
| Export: ZIP/CSV first → Google Sheets → incremental sync to customer's Postgres (paid add-on) | Data ownership without heavy infra (Airbyte needs Kubernetes and 8 GB RAM). |
| Do **not** build on top of Odoo | Its free plan forbids third-party apps/custom Python; customization ties pricing to per-user plans MYPEs won't pay. |
| **Go** for `api`, `worker`, `bot`; QR via **zxing-cpp's official Go wrapper (cgo)**; Python only as dev tooling (evals) | Founder priority: efficiency on a shared 12 GB VPS is a product/profile feature. Measured: cgo zxing-cpp matches Python exactly (114/252 synthetic photos vs 81/252 gozxing, identical on arm64); ~30 MB vs ~145 MB RSS under decode load. Wrapper is experimental: pin tag, hide behind a `QRDecoder` interface, gozxing fallback. See `docs/research/notas/evaluacion_backend_go_python.md`. |

## Open decisions (evaluate first)

1. **Product brand.** "Yupana" (the Inca abacus) is the working name.
   - yupana.com/.app/.dev are taken; yupana.io appeared free; `.pe` is unchecked.
   - A German financial-management SaaS is also named Yupana — same industry. Check INDECOPI before using it commercially.

## Still to verify

- Exact WhatsApp per-message rates for Peru, and Meta business verification requirements.
- SUNAT API rate limits; whether one platform credential suffices for Consulta Integrada across many tenants.
- Whether any SUNAT API lets a buyer download the XML of received invoices.
- Official text of Ley 29733 and its regulation DS 016-2024-JUS (current notes rely on a secondary source): database registration, breach notification, international transfer to LLM providers.
- Exact SUNAT QR field format against current regulation (RS 113-2018 Annex A 6.4: `RUC|TIPO|SERIE|NUMERO|IGV|TOTAL|FECHA|TIPO DOC ADQ|NUM DOC ADQ|VALOR RESUMEN`).

## MVP phases (summary — details in the report)

- **Phase 0:** skeleton — modules + per-tenant activation, RLS, outbox worker, Docker Compose, backups.
- **Phase 1:** email XML + Telegram photo/QR intake, SUNAT validation, deduplication, ZIP/CSV export.
- **Phase 2:** layered categorization with corrections from the bot, plus an eval set built from the pilot's real invoices.
- **Phase 3:** official WhatsApp Cloud API, plus PDF/photo extraction with a vision LLM.
- **Phase 4:** SIRE reconciliation (manual upload first, then API).
- **Phase 5:** natural-language queries via typed functions, plus sync to Google Sheets or the customer's database.
