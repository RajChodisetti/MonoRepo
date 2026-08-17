# Owner-Style Missing Features And Business Gap Analysis

Date: 2026-08-17 (reconciled from the 2026-07-19 assessment)

This is a strategic product-gap analysis, not the release-state source of truth. Use
`docs/SESSION_SUMMARY.md` and `docs/SESSION_DELIVERED.md` for current deployment
status and operational counts.

## Executive Assessment

Tuvi is not yet close to Owner.com as a production restaurant growth platform. Tuvi is closer to a sales-enablement platform that can generate personalized restaurant website demos, run reviewed outreach, capture reservation requests, and show an AI receptionist prototype.

The biggest missing business capability is first-party commerce. Without online ordering, payments, customer accounts, order history, loyalty, SMS/push marketing, delivery, and reporting from real transactions, Tuvi cannot yet claim the core Owner outcome: growing direct online orders and repeat customers.

## Current Strengths

- Personalized website demos backed by server-side payloads and token-gated access.
- Three active restaurant website templates with source-aware media handling.
- Places-first scraping, Apollo enrichment, manual media review, bounded AI website review, and human approval gates; the former OCR pipeline is retired.
- Durable, administrator-managed Gmail outreach with editable plain-text sequences, account health, pacing, sender failover, preview, provider-confirmed delivery state, and a unified reply inbox.
- Admin portal for scrape jobs, restaurant review, photos, demo/campaign management, outreach, and engagement evidence.
- Pending reservation request API and availability-slot calculation.
- Browser/phone voice agent that can act as a restaurant AI receptionist and submit pending reservation requests.
- Public demo engagement tracking including template, foreground time, and voice transcript turns.

## Current Weaknesses

- No live online ordering or checkout.
- No payments, refunds, tax, tips, fees, reconciliation, or dispute handling.
- No order lifecycle, kitchen display, prep-time management, busy mode, or staff order console.
- No customer CRM sourced from actual orders.
- No SMS marketing, push notifications, loyalty, rewards, promo codes, or abandoned-cart automation.
- No native or PWA guest mobile app with saved profile, reorder, rewards, and push.
- No delivery dispatch, delivery-zone logic, third-party driver integration, or live delivery tracking.
- No POS integration.
- No owner-facing production dashboard comparable to an Owner App.
- No review engine, listings management, or automated local SEO operations beyond site metadata/planning.
- No case-study or before/after measurement engine from real customers.

## Feature Gap Matrix

| Owner-Style Capability | Tuvi Current State | Business Gap | Priority |
| --- | --- | --- | --- |
| AI websites | Partial. Personalized demo templates, server-side payloads, reviewed media policy, reservation CTA, voice widget, engagement tracking. | Need production live-site lifecycle, custom domains, CMS/editor, domain redirects, site ops, SEO monitoring, and ongoing optimization. | Critical |
| Restaurant SEO | Early/partial. Template SEO code and a public digital-footprint review exist, but there is no full local SEO operating system. | Need sitemap/canonical/structured data guarantees, GBP/Yelp intake, local landing pages, SEO reports, keyword/ranking tracking, and human review workflow for claims. | Critical |
| Online menu | Partial. Menus/menu_items schema and public rendering exist. | Need owner menu editor, item availability, modifiers, option groups, item images at scale, allergens, taxability, prep times, import/export, and menu publishing workflow. | Critical |
| Online ordering | Missing. Reservation requests are not orders. | Need cart, checkout, pickup/delivery scheduling, payment, taxes, tips, receipts, order confirmation, failure handling, and order audit trail. | Critical |
| Smart upsells | Missing. | Need add-ons, recommended items, item bundles, cart-level suggestions, promo rules, and revenue attribution. | High |
| Payments | Missing. | Need provider adapter, PCI-safe design, Stripe or similar integration, webhooks, refunds, reconciliation, reporting, and failure/retry behavior. | Critical |
| Delivery | Missing. | Need delivery zones, fees, address validation, driver assignment, third-party dispatch adapter, in-house driver mode, status tracking, refunds, and support runbook. | High |
| Catering | Missing. | Need catering menus, lead times, order minimums, deposits, quote/request flow, production scheduling, and separate reporting. | Medium |
| Kitchen tablet / order console | Missing. | Need staff-facing order intake, accept/reject, prep time, complete/cancel, busy mode, item 86ing, sound/print alerts, and tablet layout. | Critical after ordering |
| POS integrations | Missing. | Need POS strategy ADR, one target POS first, menu sync, order injection, auth/token storage, error queue, and fallback mode. | High |
| Owner mobile app | Missing as a product. Sites are responsive; there is no guest app. | Need PWA/native app, saved profile, reordering, rewards, push consent, app publishing, support, and versioning. | High after ordering |
| Owner/staff app | Missing. Internal admin portal exists; restaurant owners do not have a production product dashboard. | Need restaurant owner login, restaurant-scoped dashboard, order/reservation/campaign controls, mobile-usable layout, alerts, and permissions. | Critical |
| Loyalty and rewards | Missing. | Need customer identity, points ledger, earn/redeem rules, reward catalog, promo abuse prevention, and reporting. | High |
| Email marketing to customers | Partial but wrong audience. Tuvi supports prospect outreach, not restaurant customer lifecycle marketing. | Need opt-in customer lists from orders/app/reservations, segments, lifecycle automations, manual sends, revenue attribution, and unsubscribe controls per restaurant. | Critical after ordering |
| SMS marketing | Missing. | Need SMS provider, opt-in capture, quiet hours, templates, segmentation, compliance, opt-out, and delivery reporting. | High |
| Push notifications | Missing. | Need PWA/native app, push provider, consent, segmentation, campaign composer, delivery/open/click attribution. | High after app |
| Marketing automation | Partial for Tuvi prospect outreach. Missing for restaurant customers. | Need order-triggered automations: welcome, reorder, abandoned cart, win-back, birthday, holiday, review request, loyalty nudges. | Critical |
| Customer CRM | Missing. | Need customer profiles, order history, preferences, consent, segments, LTV, repeat rate, and export controls. | Critical after ordering |
| Reviews engine | Missing. | Need review request after orders, review monitoring/import, reply drafting, owner approval, escalation, and review impact reporting. | Medium |
| Listings management | Missing. Scraping reads public data; it does not manage listings. | Need source inventory, profile completeness, proposed updates, manual approval, API/provider publishing where allowed, and audit history. | Medium |
| Reporting and analytics | Partial. Demo/email engagement exists; no sales/order/customer analytics. | Need sales, order source, campaign revenue, repeat rate, loyalty, SEO, reviews, top sellers, refunds, delivery problems, and operational dashboards. | Critical |
| Restaurant AI receptionist | Partial. Voice agent can answer simple questions and submit pending reservation requests, but the main Go receptionist packages remain placeholders. | Need production provider strategy, inbound call mapping, call logs in Postgres, evals, safety review, handoff, hours/menu accuracy, and restaurant owner visibility. | High |
| AI phone ordering | Missing. The current AI receptionist takes reservation requests, not food orders. | Need menu/cart tool flow, payment safety, order confirmation, kitchen handoff, and strict failure behavior. | Later |
| Content automation | Planned/reserved. | Need content_jobs schema/API/worker/UI, prompt versioning, restaurant context, factuality checks, approval/export, and history. | Medium |
| Integrations marketplace | Missing. | Need integration inventory, provider adapters, OAuth/credential storage, health checks, tenant isolation, support tooling. | Later |
| Subscription/billing for Tuvi customers | Missing. | Need customer accounts, plans, invoices, payment collection, failed-payment lifecycle, entitlements, and cancellation/export. | Medium |
| Onboarding and launch operations | Partial/manual. | Need onboarding checklist, domain/GBP/Yelp/POS/payment intake, asset requests, launch status, customer success notes, and support queue. | Critical for selling services |
| Proof/case studies | Missing. | Need baseline capture and before/after metrics: sales, orders, traffic, reviews, direct-vs-third-party savings, campaign ROI. | Critical for business credibility |

## What We Can Claim Now

Use these claims carefully:

- We can generate personalized restaurant website demos from real restaurant data.
- We can show multiple polished website designs for a restaurant lead.
- We can send human-reviewed outreach with provider-confirmed delivery state when explicitly approved and enabled.
- We can capture pending reservation requests.
- We can demonstrate an AI receptionist that discloses it is AI and can help with basic questions and reservation requests.
- We can show internal admin evidence for demo engagement and outreach state.

## What We Should Not Claim Yet

Avoid these until built and verified:

- "Commission-free online ordering."
- "We replace Owner."
- "We grow direct orders automatically."
- "Restaurants get their own mobile app."
- "SMS and push marketing are live."
- "Loyalty and rewards are live."
- "POS integrations are live."
- "Kitchen tablet/order management is live."
- "Delivery dispatch is live."
- "Confirmed reservations are supported."
- "We handle payments."
- "We manage Google/Yelp listings automatically."
- "We have before/after sales numbers."

## Business Roadblocks To Owner Parity

### Data Flywheel

Owner's value depends on first-party order/customer data. Tuvi currently has prospect and demo-engagement data, not customer transaction data. Until Tuvi processes orders, retention automation will be mostly synthetic.

### Operational Trust

Restaurants will tolerate a demo bug more than an ordering bug. Once Tuvi accepts orders, payments, or delivery, uptime, support, refund handling, and staff workflow become core obligations.

### Provider Complexity

Payments, SMS, push, POS, delivery, and listings each introduce provider contracts, credentials, webhook security, cost, rate limits, compliance, and incident response.

### Product Focus

Building website, ordering, loyalty, delivery, mobile app, POS, SEO, review, listings, voice, and AI content at once will dilute engineering. The rational build order is website wedge first, ordering second, retention third, integrations fourth.

## Recommended Business Packaging

### Package 1: AI Website Demo And Sales Conversion

Sellable soon after Stage 0 hardening.

Includes:

- personalized restaurant website demo;
- reviewed media/profile setup;
- reservation-request CTA;
- AI receptionist demo;
- tracked owner engagement;
- optional human-reviewed outreach campaign.

This is not Owner parity, but it is a strong prospecting and agency-style product.

### Package 2: Managed Restaurant Website Plus Reservation Capture

Sellable after Stage 1.

Includes:

- live restaurant website;
- basic SEO setup;
- CMS/editor or managed edits;
- pending reservations/callback requests;
- analytics summary;
- AI receptionist add-on.

This can compete with website agencies and some restaurant website builders, not Owner's full commerce suite.

### Package 3: Direct Ordering And Customer Growth

Sellable after Stages 2-4.

Includes:

- first-party online ordering;
- pickup/delivery scheduling;
- payments;
- customer CRM;
- email/SMS automation;
- loyalty/rewards;
- reporting.

This is the real Owner-comparison package.

### Package 4: Operations Platform

Sellable after Stages 5-7.

Includes:

- app/PWA;
- kitchen console/tablet;
- delivery dispatch;
- POS integration;
- review/listings management;
- detailed analytics and case-study proof.

## Near-Term Product Decisions Needed

- Is Tuvi trying to sell managed websites/AI receptionist first, or a direct-ordering SaaS?
- Which first POS matters most for the target restaurants?
- Which payments provider should be approved first?
- Is the mobile strategy PWA-first or native-app-first?
- Does Tuvi want to support delivery directly, via third-party dispatch, or not in the first commerce version?
- Are we targeting takeout-heavy restaurants, fine dining, or both? Owner's strongest fit is direct online ordering and repeat orders, especially for takeout-heavy operators.
- What is the acceptable support/on-call commitment once orders and payments go live?

## Sources

- Owner demo page: https://www.owner.com/demo
- Owner product catalog/home: https://www.owner.com/
- Owner website/SEO positioning: https://www.owner.com/restaurant-website-ai and https://www.owner.com/restaurant-seo
- Owner online ordering: https://www.owner.com/online-ordering
- Owner email/SMS marketing: https://www.owner.com/email-sms-marketing
- Owner loyalty/rewards: https://www.owner.com/loyalty-rewards
- Owner push notifications: https://www.owner.com/push-notifications
- Owner delivery: https://www.owner.com/delivery
- Owner owner app/mobile operations: https://www.owner.com/mobile
- Owner kitchen tablet: https://www.owner.com/kitchen-tablet
- Owner POS integrations: https://www.owner.com/pos-integrations
- Owner reporting/analytics: https://www.owner.com/reporting-analytics
- Current Tuvi implementation sources: `backend/internal/http/router.go`, `backend/internal/seoreport/`, `backend/migrations/`, `apps/web/src/app/(admin)/`, `template/src/`, `voice-sales-agent/`, `docs/SESSION_DELIVERED.md`

Owner product pages and the Tuvi implementation inventory were rechecked on
2026-08-17.
