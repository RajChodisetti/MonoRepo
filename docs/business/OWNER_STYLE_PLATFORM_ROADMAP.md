# Owner-Style Restaurant Growth Platform Roadmap

Date: 2026-08-17 (reconciled from the 2026-07-19 roadmap)

This is a directional business roadmap, not a release plan or deployment record.
Use `docs/SESSION_SUMMARY.md` and `docs/SESSION_DELIVERED.md` for current delivery
state and operational evidence.

## Purpose

This roadmap compares Tuvi's current restaurant platform against the broad service promise shown in Owner.com's public demo flow and product catalog: AI websites and SEO, marketing automations, mobile app, integrations, commission-free delivery, online ordering, loyalty, reporting, kitchen operations, and POS connectivity.

The honest business read: Tuvi is currently a strong sales-demo and lead-outreach system, not yet a full restaurant operating platform. We can credibly show parts of an Owner-style pitch, especially personalized websites, reviewed media, outreach, reservation requests, and an AI receptionist demo. We cannot yet credibly sell a replacement for Owner's live ordering, customer retention, mobile app, delivery, POS, kitchen, or reporting stack.

## Business Position Today

Tuvi's strongest wedge is pre-sale personalization:

- generate restaurant-specific demo sites from scraped/enriched public data;
- use manual media review and fail-closed public boundaries, with a bounded AI website review for owner-facing evidence;
- send reviewed outreach through Gmail with pacing, sender-health, operator controls, and provider-confirmed delivery state;
- capture replies, demo engagement, and pending reservation requests;
- demonstrate a browser/phone AI receptionist that can answer simple questions and create pending reservation requests;
- operate through an internal admin portal.

That is useful, but it is mostly a sales acquisition machine. Owner's public positioning is a live growth system for restaurants after they become customers: online ordering, direct customer data, repeat-order marketing, loyalty, mobile apps, delivery, POS integration, kitchen tablet, owner app, and analytics.

## Distance Estimate

| Goal | Current Distance |
| --- | --- |
| Phase 1 sales MVP | Late-stage but not finished. Most of the internal lead-to-demo-to-outreach loop exists, with remaining work around productized restaurant onboarding, owner-facing reservation/voice visibility, content automation, and dashboard depth. |
| Owner-style sales conversation | Partially credible. We can discuss AI websites, reservations, outreach automation, and AI receptionist. We should not imply live online ordering, loyalty, SMS, mobile app, delivery, POS, or kitchen operations are built. |
| Owner-style production platform | Far. The largest missing layer is first-party commerce: cart, checkout, payments, orders, fulfillment, customer identity, loyalty, and retention automation. |
| Competitive replacement for Owner | A multi-quarter effort. The original directional estimate was 9-18 months for a serious first version with a small team, assuming disciplined scope and no major provider/compliance surprises; it is not a delivery commitment. Native mobile apps, deep POS integrations, or delivery operations would expand it. |

## Strategic Recommendation

Do not try to clone Owner end to end first. Owner's core power is not just a website; it is the loop:

```text
Google discovery -> online order -> customer profile -> app/rewards signup -> email/SMS/push remarketing -> repeat order -> reporting proof
```

Tuvi does not yet have that loop because we do not yet process first-party orders or own customer transaction data. The fastest credible path is:

1. Win sales calls with personalized AI website demos plus AI receptionist.
2. Convert a few restaurants on website/reservation/voice/outreach services.
3. Add first-party ordering only after the sales-demo wedge proves demand.
4. Build retention automation from real orders, not from generic scraped leads.

## Roadmap Overview

### Stage 0: Stabilize The Sales Wedge

Target window: 0-4 weeks

Objective: Make Tuvi safe and credible for owner-facing demos without overclaiming live commerce.

Deliverables:

- Preserve the current name, valid business-email, lifecycle, inferred-business-consent, sequence-approval, and operator-send gates so eligibility remains auditable and fail closed.
- Keep the retired OCR pipeline absent; use manual media review and the bounded AI website review, and monitor their review queues and failure modes.
- Configure durable owner/licensed media storage so restaurants can upload stable assets rather than relying only on live Google previews.
- Tighten the three active website templates for SEO metadata, JSON-LD, performance, mobile conversion, accessibility, and clean empty states.
- Add or finish an owner-facing reservation dashboard, not only public reservation submission.
- Persist AI receptionist call logs into the main Go/Postgres product surface, not only the separate voice agent's SQLite/local flow.
- Build the content automation MVP that already exists in Phase 1 docs as reserved packages.
- Add a clear internal analytics summary for leads, demos generated, demo views, email sends, clicks, reservations, and AI receptionist engagement.
- Align sales copy so it says "reservation requests" and "AI receptionist demo" rather than implying confirmed table management or live ordering.

Exit criteria:

- A salesperson can open an admin lead, generate a polished demo, review media/profile data, send or stage outreach, and see engagement evidence.
- A restaurant owner can view a demo on mobile, submit a pending reservation request, and test the AI receptionist flow.
- No public page exposes raw lead enrichment, internal notes, unreviewed assets, or unsafe media.

### Stage 1: AI Website And Local SEO Product

Target window: 1-3 months

Objective: Turn personalized demo sites into a small sellable website/SEO product.

Deliverables:

- Custom-domain provisioning plan with TLS, domain verification, redirects, sitemap, robots, canonical URLs, and rollback.
- Restaurant CMS basics: hours, contact, menu, featured photos, offers, reservation policy, SEO title/description, and brand voice.
- Structured data for Restaurant, Menu, LocalBusiness, opening hours, address, phone, and review snippets where legally and factually supported.
- Google Business Profile and Yelp metadata intake during onboarding, with clear manual-update workflow before API automation.
- Local SEO pages where appropriate: location, cuisine, catering, specials, and neighborhood terms.
- "Website health" or "AI report" lead magnet for sales, similar in business function to Owner's restaurant grader.
- Live-site analytics for traffic, demo/source attribution, CTA clicks, reservation requests, and phone/voice interactions.

Exit criteria:

- Tuvi can sell "AI website plus local SEO setup" honestly, with human-operated fulfillment.
- At least 3 real restaurant websites or production-quality pilots are live with before/after metrics tracked.

### Stage 2: Direct Online Ordering MVP

Target window: 3-6 months

Objective: Build the first real commerce loop. This is the main gap versus Owner.

Deliverables:

- Menu commerce model: categories, items, variants, modifiers, required option groups, taxability, allergens, availability, prep times, photos, and sort order.
- Cart and checkout: pickup/delivery mode, scheduled ASAP/later ordering, customer contact, notes, tips, service fees, taxes, promo code hooks, and idempotent submission.
- Payments provider behind an adapter, likely Stripe first, with refunds, voids, webhooks, reconciliation, and no secrets in logs.
- Order lifecycle: pending, accepted, preparing, ready, completed, cancelled, refunded, failed.
- Restaurant order console: accept/reject, prep time, busy mode, item 86ing, customer contact, and order history.
- Customer receipts and operational notifications through email first, SMS later.
- Basic upsells: popular add-ons, item-level modifiers, cart recommendations, and reorder prompts.
- First-party customer records created from orders with explicit consent flags.

Exit criteria:

- A guest can place a real first-party pickup order on a live restaurant site.
- Staff can manage that order from a dashboard without direct database access.
- Payments, refunds, taxes, fees, and receipts have an auditable record.

### Stage 3: Customer CRM And Marketing Automations

Target window: 5-8 months

Objective: Move from prospect outreach to restaurant customer retention.

Deliverables:

- Customer table sourced from orders, reservations, loyalty signup, and opt-in forms.
- Consent model for email, SMS, and push channels.
- Segments: first-time customer, repeat customer, lapsed customer, high spender, birthday, abandoned checkout, catering prospect.
- Automated campaigns: welcome, reorder reminder, win-back, holiday/special event, abandoned cart, post-order review request, birthday/reward.
- Manual campaign composer for restaurant owners with guardrails and preview.
- SMS provider adapter and quiet-hours/compliance controls.
- Campaign attribution tied to orders, not just clicks.
- A/B testing later, after basic automations work.

Exit criteria:

- Tuvi can honestly claim marketing automation that grows repeat orders from existing customers.
- A restaurant owner can see how many customers and orders came from each campaign.

### Stage 4: Loyalty, Rewards, And Promotions

Target window: 6-9 months

Objective: Add the repeat-order mechanics Owner emphasizes through mobile app and rewards.

Deliverables:

- Loyalty accounts keyed by phone/email/customer ID.
- Points accrual from orders.
- Reward redemption rules and expiration.
- Promo codes and offers with channel/source attribution.
- Store credit or coupons with abuse prevention.
- App/web signup flow that does not require passwords for basic rewards enrollment.
- Reporting for reward signups, redemptions, retained customers, and incremental revenue.

Exit criteria:

- Guests can join a restaurant's rewards program and earn/redeem rewards on web orders.
- Marketing campaigns can target loyalty status.

### Stage 5: Mobile Experience

Target window: 7-11 months

Objective: Choose a pragmatic mobile route before committing to native app operations.

Recommended order:

1. PWA with install prompt, saved customer profile, reorder, rewards, and web push where supported.
2. White-label/native shell only after real order volume proves the maintenance burden is justified.
3. App Store and Play Store publishing operations, screenshots, review response, app updates, and support process.

Deliverables:

- Guest account/session model.
- Saved favorites and reorder.
- Push notification consent and provider setup.
- App deep links to menu, cart, offers, and rewards.
- Owner/staff mobile dashboard can remain web/PWA until native is justified.

Exit criteria:

- Tuvi can sell a mobile ordering and loyalty experience honestly, even if it begins as a PWA.

### Stage 6: Delivery, Catering, POS, And Kitchen Operations

Target window: 9-15 months

Objective: Build the operational platform layer that makes Tuvi comparable to Owner for takeout-heavy restaurants.

Deliverables:

- Delivery zones, fees, minimums, prep windows, address validation, driver notes, and delivery status.
- In-house driver mode first; third-party dispatch adapter later.
- Catering order mode with lead time, deposits, order minimums, and quote/request workflow.
- Kitchen display/tablet web app: new order alerting, accept/reject, prep time changes, complete/cancel, item 86ing, and busy mode.
- POS integration strategy and ADR: direct Toast/Square/Clover/LightSpeed, middleware, or tablet-first fallback.
- Menu/order sync adapters with retry, error visibility, and support runbooks.

Exit criteria:

- Tuvi can run a takeout/pickup restaurant's direct order flow without operational chaos.
- POS integration exists for at least one target POS, or the tablet fallback is good enough for early customers.

### Stage 7: Review, Listings, Reporting, And Proof Engine

Target window: 10-18 months

Objective: Build the measurement and proof system required to compete in restaurant growth software.

Deliverables:

- Review request flows after orders.
- Review monitoring and owner reply drafting, with human approval.
- Listings inventory across Google, Yelp, Apple Maps, Bing, Facebook, Instagram, TripAdvisor, and delivery marketplaces.
- Listings-change workflow with manual approval before publishing to external profiles.
- Reporting dashboard: sales, orders, source attribution, campaign revenue, repeat rate, loyalty, SEO visibility, reviews, refunds, delivery issues, and top items.
- Case study tooling: before/after snapshots, baseline capture, ROI calculations, exportable sales material.

Exit criteria:

- Tuvi can show before/after results instead of only demos.
- Sales can say "we grow direct online orders" with evidence from Tuvi-owned data.

## Risks And Gates

- Payments, SMS, push notifications, external listings updates, delivery dispatch, POS integrations, and production prompt/model changes all require explicit approval and provider/security review.
- A full Owner clone will create scope overload. The first decision should be whether Tuvi wants to be a website/AI receptionist agency-product hybrid or a full direct-ordering SaaS.
- First-party ordering changes the business risk profile: payment disputes, tax calculation, refunds, customer PII, restaurant support burden, uptime, and on-call become product obligations.
- Native mobile apps add operational drag through store review, device testing, app updates, push credentials, screenshots, and support.
- POS integrations can consume months if approached too broadly. Start with one target POS or a tablet-first fallback.

## Sources

- Owner demo page: https://www.owner.com/demo
- Owner home/product catalog: https://www.owner.com/
- Owner AI website page: https://www.owner.com/restaurant-website-ai
- Owner online ordering page: https://www.owner.com/online-ordering
- Owner email and SMS marketing page: https://www.owner.com/email-sms-marketing
- Owner loyalty and rewards page: https://www.owner.com/loyalty-rewards
- Owner push notifications page: https://www.owner.com/push-notifications
- Owner delivery page: https://www.owner.com/delivery
- Owner mobile app page: https://www.owner.com/mobile
- Owner kitchen tablet page: https://www.owner.com/kitchen-tablet
- Owner POS integrations page: https://www.owner.com/pos-integrations
- Owner reporting and analytics page: https://www.owner.com/reporting-analytics
- Tuvi current Phase 1 guide: `docs/phase1/PHASE1_IMPLEMENTATION_GUIDE.md`
- Tuvi current Phase 1 backlog: `docs/phase1/PHASE1_TECHNICAL_BACKLOG.md`
- Tuvi current delivery state: `docs/SESSION_DELIVERED.md`
- Tuvi current implementation: `backend/internal/seoreport/`, `apps/web/src/app/(admin)/`, `template/src/`, and `voice-sales-agent/`

Owner product pages and the Tuvi implementation inventory were rechecked on
2026-08-17.
