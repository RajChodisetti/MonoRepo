# Owner.com Product Parity Roadmap (Tuvi)

Reference: Owner.com product suite (discovery, sales, retention, ops).  
Purpose: For each product — current status, goal, crucial tasks, and where it fits in the org.

**Status legend**
- `Marketing page` — Tuvi landing page exists under `web/`
- `Partial product` — some real template/backend capability
- `Not built` — no real product yet (marketing only or missing)

---

## Summary table

| # | Product | Tuvi marketing | Real product today | Suggested MVP |
|---|---------|----------------|--------------------|---------------|
| 1 | Restaurant Website | Yes (`/restaurant-website-ai`) | Partial (templates + meta/JSON-LD) | Harden templates + CMS |
| 2 | Google / Restaurant SEO | Yes (`/product/seo`) | Partial (basic meta/JSON-LD only) | On-site SEO → listings → reviews → dashboard |
| 3 | Online Menu | Yes (`/product/menu`) | Partial (menu UI on templates) | Conversion menu + modifiers |
| 4 | Reviews Engine | Yes (`/product/reviews`) | Not built | Review ask + GBP replies |
| 5 | Listings Management | Yes (`/product/listings`) | Not built | NAP + GBP sync |
| 6 | Online Ordering | Yes (`/product/ordering`) | Not built | Cart → pay → kitchen |
| 7 | Smart Upsells | Yes (`/product/upsells`) | Not built | Checkout suggestions |
| 8 | Delivery | Yes (`/product/delivery`) | Not built | Flat-rate / hybrid delivery |
| 9 | Catering | Yes (`/product/catering`) | Not built | Catering menu + booking |
| 10 | AI Phone Ordering | Yes (`/product/ai-phone`) | Partial (voice agent exists separately) | Phone → order → kitchen |
| 11 | Branded Restaurant App | Link only | Not built | PWA/native guest app |
| 12 | Marketing Campaigns | Link only | Not built | Automated campaigns |
| 13 | Email & SMS Marketing | Link only | Not built | Lifecycle messages |
| 14 | Push Notifications | Link only | Not built | App push |
| 15 | Loyalty & Rewards | Link only | Partial (mentioned in restaurant services) | Points + rewards |
| 16 | Owner / Tuvi App | Link only | Not built | Operator mobile app |
| 17 | Reporting & Analytics | Link only | Not built | Sales / SEO / orders dashboard |
| 18 | Kitchen Tablet | Link only | Not built | Order ticket UI |
| 19 | POS Integrations | Link only | Not built | Square/Clover/Otter-style |

---

## 1. Restaurant Website

**Status:** Marketing page implemented + **partial product** (MonoRepo restaurant templates).

**Goal:** Every restaurant gets a conversion-focused site (brand, menu, SEO-ready, ordering CTAs) without custom redesign every time.

**Crucial tasks**
- Template parity (all themes: meta, JSON-LD, performance)
- Content CMS for menu/photos/hours
- Wire CTAs into real ordering when ready
- Deploy pipeline per restaurant domain

**Fits in org:** Core **Discovery** surface. Marketing sells it; Template/Web eng owns it; Success onboards content.

---

## 2. Google / Restaurant SEO

**Status:** Marketing page only (`/product/seo`) + basic template meta/JSON-LD. **Not a full SEO product.**

**Goal:** Restaurants rank better in local Google search; listings stay accurate; reviews grow; owners see visibility progress without doing SEO themselves.

**Crucial tasks**
1. Consistent schema + sitemap/robots/canonicals on all templates  
2. NAP source of truth + GBP sync  
3. Post-visit Google review prompts + reply path  
4. Simple visibility dashboard (30/60/90 day trends)  
5. Playbooks for Google algorithm updates + expert ops  

**Fits in org:** **Grow online discovery** pillar. Platform eng (APIs/jobs) + template eng (on-site) + ops (GBP onboarding) + marketing (honest claims). Sits beside Website, Listings, Reviews.

---

## 3. Online Menu

**Status:** Marketing page implemented; templates show menus. **Conversion menu product not built.**

**Goal:** Menu that turns browsers into buyers (layout, photos, bestsellers, continuous optimization).

**Crucial tasks**
- Structured menu data (categories, modifiers, prices, photos)
- Mobile-first menu UX on site
- Bestseller / featured logic
- Hook into cart when ordering ships

**Fits in org:** Discovery + Sales bridge. Shared data model with Ordering and SEO (Menu schema).

---

## 4. Reviews Engine

**Status:** Marketing page only.

**Goal:** Steady stream of Google reviews + easy replies → stronger rating and local SEO.

**Crucial tasks**
- Trigger review request after order/visit
- Deep link to Google review URL
- Request history + basic reply inbox
- Tie to SEO dashboard (rating/count)

**Fits in org:** Discovery + Success. Depends on Ordering (or manual triggers) and GBP (SEO).

---

## 5. Listings Management

**Status:** Marketing page only.

**Goal:** Name/address/phone/hours/website consistent everywhere (especially Google) for local SEO.

**Crucial tasks**
- NAP admin as single source of truth
- GBP field sync
- Expand directories later (Yelp, Facebook, etc.)
- Mismatch alerts

**Fits in org:** Discovery / Platform. Feeds SEO product; ops helps claim/verify listings.

---

## 6. Online Ordering

**Status:** Marketing page only; no full cart/checkout product.

**Goal:** Direct orders on the restaurant’s site/app — higher conversion, lower fees than marketplaces, owned customer list.

**Crucial tasks**
- Menu → modifiers → cart → checkout → payment
- Pickup/delivery rules + fees
- Order storage APIs + kitchen/POS handoff
- Guest identity / customer list
- Conversion experiments over time

**Fits in org:** **Grow online sales** core. Platform + payments + kitchen/POS. Unlocks Upsells, Delivery, Reviews, Loyalty.

---

## 7. Smart Upsells

**Status:** Marketing page only.

**Goal:** Raise average ticket with data-driven add-ons at checkout.

**Crucial tasks**
- Suggestion rules / ML later
- Checkout “goes well with” UI
- Attach-rate analytics
- Owner controls (include/exclude items)

**Fits in org:** Sales optimization layer on Ordering. Product + data eng.

---

## 8. Delivery

**Status:** Marketing page only.

**Goal:** Profitable delivery with top-rated drivers, flat fees, guest can be contacted; hybrid in-house + third-party.

**Crucial tasks**
- Delivery quoting / flat-rate rules
- Driver partner integration or dispatch
- Live tracking UX (map)
- Refund/support policy tooling

**Fits in org:** Sales + Ops. Depends on Ordering; partners for drivers.

---

## 9. Catering

**Status:** Marketing page only.

**Goal:** Commission-free catering bookings on the restaurant website (minimums, lead times, large trays).

**Crucial tasks**
- Catering menu + servings/pricing
- Request/booking form or checkout
- Minimums, notice windows, fees
- Lead notifications to restaurant

**Fits in org:** Sales / high-margin. Shares menu + site; Success for onboarding catering menus.

---

## 10. AI Phone Ordering

**Status:** Marketing page + **partial** (voice agent work exists in MonoRepo separately).

**Goal:** AI answers every missed call, takes order, sends to kitchen, grows customer list/loyalty.

**Crucial tasks**
- Telephony + AI host conversation
- Order extraction → kitchen ticket
- Fallback to human
- Loyalty / CRM capture
- Waitlist → GA launch

**Fits in org:** Sales + Voice/AI team. Integrates Ordering + Loyalty + Kitchen.

---

## 11. Branded Restaurant App

**Status:** Nav link only (no product page content config yet).

**Goal:** Guest app that drives ~repeat orders (Owner claims high repeat lift).

**Crucial tasks**
- Auth, menu, order history, push
- Reuse ordering backend
- App store / PWA strategy

**Fits in org:** Retention. Depends on Ordering; feeds Push + Loyalty.

---

## 12–15. Marketing Campaigns / Email & SMS / Push / Loyalty

**Status:** Links only (not built as Tuvi products).

**Goal:** Bring guests back automatically (campaigns, email/SMS, push, points).

**Crucial tasks**
- Guest profiles from orders
- Campaign builder + templates
- ESP/SMS provider
- Points earn/burn rules
- Push via app

**Fits in org:** **Grow repeat orders**. Marketing + Platform. Depends on Ordering customer list.

---

## 16–19. Run your restaurant (Owner App, Analytics, Kitchen Tablet, POS)

**Status:** Links only.

**Goal:** Operators run day-to-day — see orders/sales/SEO, fire kitchen tickets, sync POS.

**Crucial tasks**
- Kitchen ticket UI + sound/ack
- Metrics: sales, orders, fees saved, SEO health
- POS menu/order sync (Square/Clover/Otter-class)
- Operator mobile app

**Fits in org:** Ops pillar. Kitchen/POS unlock Ordering; Analytics sits across SEO + Sales + Retention.

---

## Recommended build order (org-wide)

1. **Restaurant Website harden** (already partial)  
2. **On-site SEO (Phase 1)** + **Online Menu data model**  
3. **Online Ordering MVP** (unlocks everything else)  
4. **Listings/GBP + Reviews** (SEO story becomes real)  
5. **Upsells / Catering / Delivery** (sales expand)  
6. **AI Phone** (productize voice)  
7. **Loyalty / Email / App / Kitchen / POS / Analytics**

---

## Ownership (minimal)

| Area | Owns |
|------|------|
| Product | Scope, SKUs, honest marketing vs ship state |
| Template / Web eng | Website, on-site SEO, menu UX |
| Platform eng | Orders, listings, GBP, CRM, jobs |
| Voice / AI | AI phone |
| Ops / Success | GBP claim, menu onboarding, review coaching |
| Marketing | Landing pages already largely done — update claims as features ship |

---

*Generated for Tuvi vs Owner.com parity planning.*
