/**
 * Restaurant demo site — loads restaurants[index] from data JSON.
 * URL: ?id=1  or  set window.RESTAURANT_CONFIG.index in config.js
 */

(function () {
  "use strict";

  const $ = (sel) => document.querySelector(sel);
  const $$ = (sel) => document.querySelectorAll(sel);

  const DAY_ORDER = [
    "monday", "tuesday", "wednesday", "thursday",
    "friday", "saturday", "sunday",
  ];

  const FOOD_CUISINE_KEYWORDS = [
    "restaurant", "food", "meal", "dining", "cafe", "bistro",
    "bar", "pub", "alcohol", "beer", "wine", "coffee", "cocktail",
    "vegan", "vegetarian", "gluten", "halal", "kosher",
  ];

  function getRestaurantIndex() {
    const params = new URLSearchParams(window.location.search);
    if (params.has("id")) {
      const n = parseInt(params.get("id"), 10);
      if (!Number.isNaN(n) && n >= 0) return n;
    }
    if (params.has("index")) {
      const n = parseInt(params.get("index"), 10);
      if (!Number.isNaN(n) && n >= 0) return n;
    }
    return window.RESTAURANT_CONFIG?.index ?? 0;
  }

  function escapeHtml(str) {
    return String(str ?? "")
      .replace(/&/g, "&amp;")
      .replace(/</g, "&lt;")
      .replace(/>/g, "&gt;")
      .replace(/"/g, "&quot;");
  }

  /** Open Gmail compose with restaurant email in To field. */
  function gmailComposeUrl(toEmail, subject) {
    const to = String(toEmail || "").trim();
    if (!to) return "";
    const params = new URLSearchParams({ view: "cm", fs: "1", to });
    if (subject) params.set("su", subject);
    return `https://mail.google.com/mail/?${params.toString()}`;
  }

  function titleCase(s) {
    return String(s)
      .toLowerCase()
      .replace(/\b\w/g, (c) => c.toUpperCase());
  }

  function pickImage(images) {
    if (!images?.length) return "";
    const first = images[0];
    if (typeof first === "string") return first;
    return first.url || first.thumbnail || "";
  }

  function isUsableImageUrl(url) {
    if (!url || typeof url !== "string") return false;
    const trimmed = url.trim();
    return /^https?:\/\//i.test(trimmed);
  }

  /** Wire load/error — handles cached images (complete before listener). */
  function wireImage(img, onOk, onFail) {
    const settle = () => {
      if (img.naturalWidth > 0) onOk?.();
      else onFail?.();
    };
    img.addEventListener("load", settle, { once: true });
    img.addEventListener("error", onFail, { once: true });
    if (img.complete) settle();
  }

  /** Menu card image — show by default, remove wrapper only on error. */
  function appendImage(parent, url, alt, wrapClass) {
    if (!isUsableImageUrl(url)) return;

    const wrap = document.createElement("div");
    wrap.className = wrapClass;

    const img = document.createElement("img");
    img.alt = alt || "";
    img.loading = "lazy";
    img.decoding = "async";
    img.referrerPolicy = "no-referrer";

    wireImage(
      img,
      null,
      () => wrap.remove()
    );

    wrap.appendChild(img);
    parent.prepend(wrap);
    img.src = url;
  }

  /** Single img — hide wrapper only when URL missing or load fails. */
  function setImage(imgEl, url, alt, hideTarget) {
    const target = hideTarget || imgEl;
    if (!imgEl || !isUsableImageUrl(url)) {
      if (target) target.hidden = true;
      return;
    }

    target.hidden = false;
    imgEl.alt = alt || "";
    imgEl.referrerPolicy = "no-referrer";
    wireImage(
      imgEl,
      null,
      () => {
        target.hidden = true;
        imgEl.removeAttribute("src");
      }
    );
    imgEl.src = url;
  }

  /** Gallery tile — visible by default; removed only if image fails. */
  function createGalleryItem(url, layoutClass, alt, index, onRemoved) {
    if (!isUsableImageUrl(url)) return null;

    const item = document.createElement("div");
    item.className = `gallery__item gallery__item--${layoutClass} fx-card`;
    item.style.setProperty("--reveal-i", String(index % 9));

    const img = document.createElement("img");
    img.alt = alt || `Gallery ${index + 1}`;
    img.loading = "lazy";
    img.decoding = "async";
    img.referrerPolicy = "no-referrer";

    wireImage(img, null, () => {
      item.remove();
      onRemoved?.();
    });

    item.appendChild(img);
    img.src = url;
    return item;
  }

  function heroImage(r) {
    const imgs = r.images || {};
    if (imgs.thumbnail) return imgs.thumbnail;
    const gal = imgs.gallery || [];
    if (gal[0]?.url) return gal[0].url;
    const menu = (r.menu_items || []).find((m) => pickImage(m.images));
    return menu ? pickImage(menu.images) : "";
  }

  function foodCuisines(list) {
    return (list || []).filter((c) => {
      const low = c.toLowerCase();
      return !FOOD_CUISINE_KEYWORDS.some((k) => low === k || low.includes(k + " "));
    }).slice(0, 6);
  }

  function primaryCuisine(r) {
    const cuisines = foodCuisines(r.cuisines);
    if (cuisines.length) return cuisines[0];
    const raw = (r.cuisines || []).find((c) => /restaurant|cuisine|food/i.test(c));
    return raw || "Fine Dining";
  }

  function aboutDescription(r) {
    const city = r.location?.city || "Australia";
    const cuisine = primaryCuisine(r);
    const rating = r.rating ? `Rated ${r.rating} stars by ${r.reviews_count || "our"} guests` : "";
    return (
      `Welcome to ${r.name} — a celebrated ${cuisine.toLowerCase()} destination in ${city}. ` +
      `We craft memorable dining experiences with seasonal ingredients, warm hospitality, ` +
      `and an atmosphere designed for every occasion. ${rating}.`
    ).trim();
  }

  function starsHtml(n) {
    const full = Math.round(Number(n) || 0);
    return "★".repeat(Math.min(5, full)) + "☆".repeat(Math.max(0, 5 - full));
  }

  function groupMenuByCategory(items) {
    const groups = new Map();
    for (const item of items || []) {
      const cat = (item.category || "Menu").trim();
      if (!groups.has(cat)) groups.set(cat, []);
      groups.get(cat).push(item);
    }
    return groups;
  }

  function normalizeCategoryLabel(cat) {
    return cat
      .replace(/^A LA CARTE\s*-\s*/i, "")
      .replace(/\s+/g, " ")
      .trim();
  }

  function showError(msg) {
    $("#loader")?.classList.add("hidden");
    $("#app")?.classList.add("hidden");
    const err = $("#error");
    err?.classList.remove("hidden");
    const el = $("#error-msg");
    if (el) el.textContent = msg;
  }

  function renderRestaurant(r) {
    const img = heroImage(r);
    const city = [r.location?.city, r.location?.state, r.location?.country]
      .filter(Boolean)
      .join(", ");

    document.title = `${r.name} — ${primaryCuisine(r)}`;

    // Nav & footer
    $("#nav-brand").textContent = r.name;
    $("#footer-name").textContent = r.name;

    // Hero
    if (img) $("#hero-bg").style.backgroundImage = `url('${img}')`;
    $("#hero-cuisine").textContent = primaryCuisine(r);
    $("#hero-title").textContent = r.name;
    $("#hero-location").textContent = city;
    $("#hero-rating").innerHTML = r.rating
      ? `<span>${starsHtml(r.rating)}</span> ${r.rating} · ${r.reviews_count || 0} reviews`
      : "";
    $("#hero-price").textContent = r.price_level || "";

    // About
    $("#about-title").textContent = `Dining at ${r.name}`;
    $("#about-desc").textContent = aboutDescription(r);
    const tags = $("#about-tags");
    tags.innerHTML = foodCuisines(r.cuisines)
      .concat((r.cuisines || []).slice(0, 4))
      .filter((v, i, a) => a.indexOf(v) === i)
      .slice(0, 8)
      .map((c) => `<li>${escapeHtml(c)}</li>`)
      .join("");
    const aboutImg = pickImage((r.menu_items || [])[0]?.images) || img;
    const aboutWrap = document.querySelector(".about__img-wrap");
    if (aboutImg && aboutWrap) {
      setImage($("#about-img"), aboutImg, r.name, aboutWrap);
    } else if (aboutWrap) {
      aboutWrap.hidden = true;
    }
    $("#about-card").innerHTML =
      r.rating
        ? `<strong>${r.rating}</strong><br><span style="font-size:0.85em;opacity:0.85">Guest rating</span>`
        : escapeHtml(r.name);

    // Menu
    const menuItems = r.menu_items || [];
    $("#menu-sub").textContent = menuItems.length
      ? `${menuItems.length} dishes curated from our kitchen`
      : "Menu coming soon";

    const groups = groupMenuByCategory(menuItems);
    const categories = [...groups.keys()];
    const filters = $("#menu-filters");
    const grid = $("#menu-grid");
    let activeCat = "all";

    function renderMenu(filter) {
      grid.innerHTML = "";
      const items =
        filter === "all"
          ? menuItems
          : groups.get(filter) || [];

      for (const [idx, item] of items.slice(0, 48).entries()) {
        const photo = pickImage(item.images);
        const popular = item.images?.[0]?.popular;
        const card = document.createElement("article");
        card.className = "menu-card fx-card";
        card.style.setProperty("--reveal-i", String(idx % 12));

        const body = document.createElement("div");
        body.className = "menu-card__body";
        body.innerHTML = `
            <p class="menu-card__cat">${escapeHtml(normalizeCategoryLabel(item.category || ""))}</p>
            <h3 class="menu-card__name">${escapeHtml(item.name)}${popular ? '<span class="menu-card__popular">Popular</span>' : ""}</h3>
            ${item.description ? `<p class="menu-card__desc">${escapeHtml(item.description)}</p>` : ""}
            ${item.price ? `<p class="menu-card__price">${escapeHtml(item.price)}</p>` : ""}`;

        card.appendChild(body);
        if (photo) {
          appendImage(card, photo, item.name, "menu-card__img");
        }
        grid.appendChild(card);
      }
      window.refreshCardTilt?.(grid);
    }

    if (categories.length > 1) {
      filters.innerHTML =
        `<button class="menu__filter active" data-cat="all">All</button>` +
        categories
          .slice(0, 8)
          .map(
            (c) =>
              `<button class="menu__filter" data-cat="${escapeHtml(c)}">${escapeHtml(normalizeCategoryLabel(c))}</button>`
          )
          .join("");
      filters.addEventListener("click", (e) => {
        const btn = e.target.closest(".menu__filter");
        if (!btn) return;
        $$(".menu__filter").forEach((b) => b.classList.remove("active"));
        btn.classList.add("active");
        activeCat = btn.dataset.cat;
        renderMenu(activeCat);
      });
    }
    renderMenu("all");

    // Gallery
    const galleryEl = $("#gallery-grid");
    const galleryUrls = [];
    for (const g of r.images?.gallery || []) {
      if (g.url) galleryUrls.push(g.url);
    }
    for (const item of menuItems) {
      const u = pickImage(item.images);
      if (u && !galleryUrls.includes(u)) galleryUrls.push(u);
    }
    const layouts = ["wide", "tall", "sq", "half", "sq", "half"];
    const gallerySection = document.querySelector("#gallery");
    gallerySection?.classList.remove("hidden");
    galleryEl.innerHTML = "";

    const syncGallerySection = () => {
      if (!galleryEl.children.length) {
        gallerySection?.classList.add("hidden");
      }
    };

    galleryUrls.slice(0, 9).forEach((url, i) => {
      const el = createGalleryItem(
        url,
        layouts[i % layouts.length],
        `Gallery ${i + 1}`,
        i,
        syncGallerySection
      );
      if (el) galleryEl.appendChild(el);
    });

    if (!galleryUrls.length) {
      gallerySection?.classList.add("hidden");
    }

    // Reviews
    const reviews = r.reviews || [];
    $("#reviews-aggregate").textContent = r.rating
      ? `${r.rating} average from ${r.reviews_count || reviews.length} reviews`
      : "";
    $("#reviews-grid").innerHTML = reviews
      .slice(0, 6)
      .map(
        (rev, i) => `
        <article class="review-card fx-card" style="--reveal-i: ${i}">
          <div class="review-card__stars">${starsHtml(rev.stars)}</div>
          <p class="review-card__text">"${escapeHtml(rev.review)}"</p>
          <p class="review-card__author">${escapeHtml(rev.reviewer || "Guest")}</p>
          <p class="review-card__date">${escapeHtml(rev.date || "")}</p>
        </article>`
      )
      .join("");
    if (!reviews.length) {
      document.querySelector("#reviews")?.classList.add("hidden");
    }

    // Visit
    const addr = r.location?.address || city;
    $("#visit-address").innerHTML = escapeHtml(addr).replace(/\n/g, "<br>");
    const contact = r.contact || {};
    const emailTo = (contact.email || "").trim();
    const gmailUrl = emailTo
      ? gmailComposeUrl(emailTo, `Table reservation — ${r.name}`)
      : "";

    $("#visit-contact").innerHTML = [
      contact.phone
        ? `<a href="tel:${contact.phone.replace(/\s/g, "")}">${escapeHtml(contact.phone)}</a>`
        : "",
      gmailUrl
        ? `<a href="${gmailUrl}" target="_blank" rel="noopener noreferrer">${escapeHtml(emailTo)}</a>`
        : "",
      contact.website
        ? `<a href="${escapeHtml(contact.website)}" target="_blank" rel="noopener">Website</a>`
        : "",
    ]
      .filter(Boolean)
      .join("");

    const hoursBody = $("#hours-body");
    const hours = r.hours || {};
    hoursBody.innerHTML = DAY_ORDER.map((day) => {
      const val = hours[day] || "—";
      return `<tr><td>${titleCase(day)}</td><td>${escapeHtml(val)}</td></tr>`;
    }).join("");

    const lat = r.location?.coordinates?.latitude;
    const lng = r.location?.coordinates?.longitude;
    const mapUrl =
      lat && lng
        ? `https://www.google.com/maps?q=${lat},${lng}`
        : `https://www.google.com/maps/search/?api=1&query=${encodeURIComponent(addr)}`;
    $("#visit-map").href = mapUrl;
    if (lat && lng) {
      $("#visit-map-frame").innerHTML = `
        <iframe
          loading="lazy"
          referrerpolicy="no-referrer-when-downgrade"
          src="https://maps.google.com/maps?q=${lat},${lng}&z=15&output=embed"
          title="Map"
        ></iframe>`;
    }

    // Reserve
    $("#reserve-text").textContent = `Join us at ${r.name} in ${r.location?.city || "town"}. Call or email to reserve your table.`;
    const reserveActions = $("#reserve-actions");
    reserveActions.innerHTML = [
      contact.phone
        ? `<a class="btn btn--primary" href="tel:${contact.phone.replace(/\s/g, "")}">Call ${escapeHtml(contact.phone)}</a>`
        : "",
      gmailUrl
        ? `<a class="btn btn--ghost" href="${gmailUrl}" target="_blank" rel="noopener noreferrer">Email Us</a>`
        : "",
      contact.website
        ? `<a class="btn btn--ghost" href="${escapeHtml(contact.website)}" target="_blank" rel="noopener">Official Site</a>`
        : "",
    ]
      .filter(Boolean)
      .join("");
  }

  function initNav() {
    const header = $("#header");
    const toggle = $("#nav-toggle");
    const links = $("#nav-links");

    window.addEventListener("scroll", () => {
      header?.classList.toggle("header--scrolled", window.scrollY > 40);
    });

    toggle?.addEventListener("click", () => {
      links?.classList.toggle("open");
    });

    links?.querySelectorAll("a").forEach((a) => {
      a.addEventListener("click", () => links?.classList.remove("open"));
    });
  }

  async function load() {
    const index = getRestaurantIndex();
    const dataPath = window.RESTAURANT_CONFIG?.dataPath || "../data/restaurants_data.json";

    try {
      const res = await fetch(dataPath);
      if (!res.ok) throw new Error(`HTTP ${res.status} loading ${dataPath}`);
      const data = await res.json();
      const restaurants = data.restaurants || [];

      if (!restaurants.length) throw new Error("No restaurants in data file");
      if (index >= restaurants.length) {
        throw new Error(
          `Restaurant index ${index} out of range (0–${restaurants.length - 1}). ` +
          `Use ?id=0 for "${restaurants[0].name}"`
        );
      }

      const restaurant = restaurants[index];
      renderRestaurant(restaurant);
      initNav();
      window.initScrollEffects?.();

      $("#loader")?.classList.add("hidden");
      $("#app")?.classList.remove("hidden");

      console.info(
        `[Tuvi Template] Loaded restaurants[${index}]: ${restaurant.name} ` +
        `(${restaurants.length} total — try ?id=${(index + 1) % restaurants.length})`
      );
    } catch (err) {
      showError(err.message);
      console.error(err);
    }
  }

  if (document.readyState === "loading") {
    document.addEventListener("DOMContentLoaded", load);
  } else {
    load();
  }
})();
