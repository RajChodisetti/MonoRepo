import "./styles.css";
import {
  ArrowRight,
  AudioLines,
  BadgeDollarSign,
  Bot,
  CalendarCheck,
  ChartNoAxesCombined,
  ChefHat,
  ClipboardCheck,
  Coins,
  Globe2,
  MailCheck,
  Megaphone,
  MousePointerClick,
  QrCode,
  RefreshCw,
  Send,
  ShieldCheck,
  Sparkles,
  Star,
  Store,
  UtensilsCrossed,
  WandSparkles,
  createIcons
} from "lucide";

const services = [
  {
    id: "demo-sites",
    icon: "globe-2",
    title: "Custom Premium Websites",
    summary: "Build a polished restaurant website around the venue, menu, and brand.",
    pain: "Many restaurant sites look dated, generic, or hard to use on mobile.",
    move: "Launch a custom site that feels premium and makes ordering, booking, and discovery simple.",
    impact: "Gives guests a stronger first impression and gives owners a real digital storefront.",
    accent: "gold"
  },
  {
    id: "qr-ordering",
    icon: "qr-code",
    title: "QR Ordering App",
    summary: "Let guests browse, customize, and order from the table.",
    pain: "Rush-hour service slows down when staff become order takers.",
    move: "Add a branded table-ordering flow that sends requests to staff clearly.",
    impact: "Improves speed, average order value, and table throughput.",
    accent: "teal"
  },
  {
    id: "rewards",
    icon: "star",
    title: "Rewards and Membership",
    summary: "Points, tiers, member perks, birthdays, and visit tracking.",
    pain: "Restaurants pay to win new guests but fail to bring them back.",
    move: "Package loyalty as a clear repeat-visit engine.",
    impact: "Creates a retention story beyond the website build.",
    accent: "violet"
  },
  {
    id: "voice-ai",
    icon: "audio-lines",
    title: "AI Voice Receptionist",
    summary: "A browser voice host that answers questions and captures bookings.",
    pain: "Missed calls during service hours turn into lost revenue.",
    move: "Answer common questions and capture booking requests when staff are busy.",
    impact: "Keeps more guest intent from being lost during service hours.",
    accent: "cyan"
  },
  {
    id: "reservations",
    icon: "calendar-check",
    title: "Smart Reservations",
    summary: "Capture reservation requests with pending status and follow-up.",
    pain: "Restaurants need demand capture without overpromising availability.",
    move: "Route requests through a safe pending-reservation workflow.",
    impact: "Gives owners a cleaner way to collect and review guest demand.",
    accent: "green"
  },
  {
    id: "campaigns",
    icon: "mail-check",
    title: "Customer Promotions",
    summary: "Email and SMS-ready campaigns for offers, events, rewards, and reactivation.",
    pain: "Restaurants often have no simple way to bring regulars back at the right moment.",
    move: "Turn menus, specials, and loyalty moments into polished promotional campaigns.",
    impact: "Creates repeat visits without asking owners to become marketers.",
    accent: "coral"
  },
  {
    id: "automation",
    icon: "wand-sparkles",
    title: "Menu and Content Automation",
    summary: "OCR menu data, organize photos, draft captions, scripts, and CTAs.",
    pain: "Owners lack time to turn menu and photo assets into campaigns.",
    move: "Convert menus and photos into organized, reviewed marketing assets.",
    impact: "Extends the product story into ongoing growth operations.",
    accent: "lime"
  },
  {
    id: "owner-dashboard",
    icon: "shield-check",
    title: "Owner Dashboard",
    summary: "A clean place to review orders, reservations, rewards, and performance.",
    pain: "Owners need visibility without another complicated operational tool.",
    move: "Give owners a focused dashboard for the systems Tuvi Solutions runs.",
    impact: "Makes the platform feel manageable after launch, not just impressive in a demo.",
    accent: "blue"
  }
];

const features = [
  "QR ordering gives guests a direct path from table scan to kitchen request.",
  "Rewards turn counter check-ins into repeat-visit moments.",
  "Staff see cleaner requests during busy service windows.",
  "Owners can connect ordering, loyalty, reservations, and campaigns from one platform."
];

const icons = {
  ArrowRight,
  AudioLines,
  BadgeDollarSign,
  Bot,
  CalendarCheck,
  ChartNoAxesCombined,
  ChefHat,
  ClipboardCheck,
  Coins,
  Globe2,
  MailCheck,
  Megaphone,
  MousePointerClick,
  QrCode,
  RefreshCw,
  Send,
  ShieldCheck,
  Sparkles,
  Star,
  Store,
  UtensilsCrossed,
  WandSparkles
};

const serviceGrid = document.querySelector("#service-grid");
const serviceDetail = document.querySelector("#service-detail");
const featureListEl = document.querySelector("#feature-list");

let activeService = services[0];

function serviceButtonTemplate(service, index) {
  return `
    <button class="service-card reveal" data-service-id="${service.id}" data-accent="${service.accent}" aria-pressed="${index === 0}" style="--stack-index: ${index};">
      <span class="service-index">${String(index + 1).padStart(2, "0")}</span>
      <span class="service-visual" aria-hidden="true">
        <i data-lucide="${service.icon}"></i>
        <span class="service-ring ring-a"></span>
        <span class="service-ring ring-b"></span>
        <span class="service-dot dot-a"></span>
        <span class="service-dot dot-b"></span>
      </span>
      <span class="service-copy">
        <strong>${service.title}</strong>
        <span>${service.summary}</span>
      </span>
    </button>
  `;
}

function detailTemplate(service) {
  return `
    <div class="detail-header" data-accent="${service.accent}">
      <i data-lucide="${service.icon}"></i>
      <div>
        <span>Selected service</span>
        <h3>${service.title}</h3>
      </div>
    </div>
    <dl class="detail-list">
      <div>
        <dt>Pain</dt>
        <dd>${service.pain}</dd>
      </div>
      <div>
        <dt>Tuvi Solutions approach</dt>
        <dd>${service.move}</dd>
      </div>
      <div>
        <dt>Business impact</dt>
        <dd>${service.impact}</dd>
      </div>
    </dl>
    <a class="inline-action" href="#contact">
      <span>Talk through this service</span>
      <i data-lucide="arrow-right"></i>
    </a>
  `;
}

function renderServices() {
  serviceGrid.innerHTML = services.map(serviceButtonTemplate).join("");
  serviceDetail.innerHTML = detailTemplate(activeService);

  serviceGrid.querySelectorAll("[data-service-id]").forEach((button) => {
    button.addEventListener("click", () => {
      setActiveService(button.dataset.serviceId);
    });
  });
}

function setActiveService(serviceId) {
  activeService = services.find((service) => service.id === serviceId) || services[0];
  serviceGrid
    .querySelectorAll("[data-service-id]")
    .forEach((node) => node.setAttribute("aria-pressed", String(node.dataset.serviceId === activeService.id)));
  serviceDetail.innerHTML = detailTemplate(activeService);
  createIcons({ icons });
}

function renderFeatureList() {
  if (!featureListEl) return;

  featureListEl.innerHTML = features
    .map(
      (feature) => `
        <div class="feature-item reveal">
          <i data-lucide="clipboard-check"></i>
          <span>${feature}</span>
        </div>
      `
    )
    .join("");
}

function initReveal() {
  const revealObserver = new IntersectionObserver(
    (entries) => {
      entries.forEach((entry) => {
        if (entry.isIntersecting) {
          entry.target.classList.add("is-visible");
        }
      });
    },
    { threshold: 0.16, rootMargin: "0px 0px -48px 0px" }
  );

  document.querySelectorAll(".reveal").forEach((node) => revealObserver.observe(node));
}

function initTopbar() {
  const topbar = document.querySelector("[data-topbar]");
  const update = () => {
    topbar?.classList.toggle("is-scrolled", window.scrollY > 24);
  };
  update();
  window.addEventListener("scroll", update, { passive: true });
}

function initHeroMotion() {
  const canvas = document.querySelector("[data-hero-motion]");
  const hero = document.querySelector(".hero-section");
  const reducedMotion = window.matchMedia("(prefers-reduced-motion: reduce)").matches;

  if (!(canvas instanceof HTMLCanvasElement) || !hero) return;

  const ctx = canvas.getContext("2d");
  if (!ctx) return;

  const pointer = { x: 0.68, y: 0.34 };
  const target = { x: 0.68, y: 0.34 };
  const nodes = Array.from({ length: 34 }, (_, index) => ({
    x: (index * 31) % 100,
    y: 12 + ((index * 47) % 78),
    radius: 1.2 + (index % 4) * 0.42,
    speed: 0.45 + (index % 6) * 0.08,
    phase: index * 0.73
  }));

  let width = 0;
  let height = 0;
  let animationFrame = 0;

  const resize = () => {
    const rect = canvas.getBoundingClientRect();
    const dpr = Math.min(window.devicePixelRatio || 1, 2);
    width = Math.max(1, Math.floor(rect.width));
    height = Math.max(1, Math.floor(rect.height));
    canvas.width = Math.floor(width * dpr);
    canvas.height = Math.floor(height * dpr);
    ctx.setTransform(dpr, 0, 0, dpr, 0, 0);
  };

  const updatePointer = (clientX, clientY) => {
    const rect = canvas.getBoundingClientRect();
    target.x = Math.max(0, Math.min(1, (clientX - rect.left) / rect.width));
    target.y = Math.max(0, Math.min(1, (clientY - rect.top) / rect.height));
  };

  const drawWave = (time, layer) => {
    const base = height * (0.18 + layer * 0.16);
    const amplitude = 20 + layer * 11;
    const offset = (pointer.y - 0.5) * 42 * (layer + 1);
    const hue = layer % 2 === 0 ? "66, 210, 200" : "216, 177, 106";

    ctx.beginPath();
    for (let x = -40; x <= width + 40; x += 16) {
      const drift = Math.sin(x * 0.009 + time * (0.7 + layer * 0.16) + layer) * amplitude;
      const pulse = Math.cos(x * 0.017 - time * 0.54 + layer) * (8 + layer * 4);
      const y = base + drift + pulse + offset;
      if (x === -40) ctx.moveTo(x, y);
      else ctx.lineTo(x, y);
    }
    ctx.strokeStyle = `rgba(${hue}, ${0.22 - layer * 0.025})`;
    ctx.lineWidth = 1.2 + layer * 0.32;
    ctx.stroke();
  };

  const drawNodes = (time) => {
    const resolved = nodes.map((node) => {
      const pullX = (pointer.x - 0.5) * 28 * node.speed;
      const pullY = (pointer.y - 0.5) * 20 * node.speed;
      return {
        x: (node.x / 100) * width + Math.sin(time * node.speed + node.phase) * 22 + pullX,
        y: (node.y / 100) * height + Math.cos(time * node.speed * 0.8 + node.phase) * 14 + pullY,
        radius: node.radius
      };
    });

    for (let i = 0; i < resolved.length; i += 1) {
      for (let j = i + 1; j < resolved.length; j += 1) {
        const dx = resolved[i].x - resolved[j].x;
        const dy = resolved[i].y - resolved[j].y;
        const distance = Math.hypot(dx, dy);
        if (distance > 145) continue;
        const alpha = (1 - distance / 145) * 0.16;
        ctx.strokeStyle = `rgba(247, 242, 232, ${alpha})`;
        ctx.lineWidth = 1;
        ctx.beginPath();
        ctx.moveTo(resolved[i].x, resolved[i].y);
        ctx.lineTo(resolved[j].x, resolved[j].y);
        ctx.stroke();
      }
    }

    resolved.forEach((node, index) => {
      ctx.beginPath();
      ctx.arc(node.x, node.y, node.radius, 0, Math.PI * 2);
      ctx.fillStyle = index % 3 === 0 ? "rgba(216, 177, 106, 0.78)" : "rgba(66, 210, 200, 0.68)";
      ctx.fill();
    });
  };

  const drawPointerHalo = () => {
    const x = pointer.x * width;
    const y = pointer.y * height;
    const gradient = ctx.createRadialGradient(x, y, 0, x, y, Math.max(width, height) * 0.26);
    gradient.addColorStop(0, "rgba(66, 210, 200, 0.24)");
    gradient.addColorStop(0.42, "rgba(216, 177, 106, 0.12)");
    gradient.addColorStop(1, "rgba(7, 8, 7, 0)");
    ctx.fillStyle = gradient;
    ctx.fillRect(0, 0, width, height);
  };

  const draw = (timestamp = 0) => {
    const time = timestamp * 0.001;
    pointer.x += (target.x - pointer.x) * 0.08;
    pointer.y += (target.y - pointer.y) * 0.08;

    ctx.clearRect(0, 0, width, height);
    drawPointerHalo();
    for (let layer = 0; layer < 4; layer += 1) drawWave(time, layer);
    drawNodes(time);

    if (!reducedMotion) {
      animationFrame = window.requestAnimationFrame(draw);
    }
  };

  resize();

  hero.addEventListener("pointermove", (event) => updatePointer(event.clientX, event.clientY), {
    passive: true
  });
  hero.addEventListener(
    "touchmove",
    (event) => {
      const touch = event.touches[0];
      if (touch) updatePointer(touch.clientX, touch.clientY);
    },
    { passive: true }
  );
  window.addEventListener("resize", () => {
    resize();
    if (reducedMotion) draw();
  });

  if (reducedMotion) {
    draw();
  } else {
    animationFrame = window.requestAnimationFrame(draw);
  }

  window.addEventListener("pagehide", () => window.cancelAnimationFrame(animationFrame));
}

function initVideoFallbacks() {
  document.querySelectorAll("video").forEach((video) => {
    video.addEventListener("error", () => {
      video.classList.add("video-missing");
    });
  });
}

function initManagedVideos() {
  const videos = Array.from(document.querySelectorAll("video[data-src]"));
  if (videos.length === 0) return;

  const visibleVideos = new Set();

  const loadVideo = (video) => {
    if (video.dataset.loaded === "true") return;
    video.src = video.dataset.src;
    video.dataset.loaded = "true";
    video.load();
  };

  const playVideo = (video) => {
    loadVideo(video);
    video.play().catch(() => {
      // Mobile browsers may defer playback until enough data is buffered.
    });
  };

  const pauseVideo = (video) => {
    video.pause();
  };

  const observer = new IntersectionObserver(
    (entries) => {
      entries.forEach((entry) => {
        const video = entry.target;
        if (!(video instanceof HTMLVideoElement)) return;

        if (entry.isIntersecting) {
          visibleVideos.add(video);
          playVideo(video);
        } else {
          visibleVideos.delete(video);
          pauseVideo(video);
        }
      });
    },
    { threshold: 0.42, rootMargin: "160px 0px" }
  );

  videos.forEach((video) => observer.observe(video));

  document.addEventListener("visibilitychange", () => {
    if (document.hidden) {
      videos.forEach(pauseVideo);
      return;
    }
    visibleVideos.forEach(playVideo);
  });
}

function initNavSpy() {
  const navLinks = Array.from(document.querySelectorAll(".nav-links a"));
  const sections = navLinks
    .map((link) => document.querySelector(link.getAttribute("href")))
    .filter(Boolean);

  const observer = new IntersectionObserver(
    (entries) => {
      entries.forEach((entry) => {
        if (!entry.isIntersecting) return;
        navLinks.forEach((link) => {
          link.classList.toggle("is-active", link.getAttribute("href") === `#${entry.target.id}`);
        });
      });
    },
    { threshold: 0.5 }
  );

  sections.forEach((section) => observer.observe(section));
}

renderServices();
renderFeatureList();
createIcons({ icons });
initReveal();
initTopbar();
initHeroMotion();
initVideoFallbacks();
initManagedVideos();
initNavSpy();
