/**
 * Parallax scroll, reveal animations, and card hover effects.
 */
(function () {
  "use strict";

  const prefersReducedMotion = window.matchMedia(
    "(prefers-reduced-motion: reduce)"
  ).matches;

  function initParallax() {
    if (prefersReducedMotion) return;

    const heroBg = document.getElementById("hero-bg");
    const heroContent = document.querySelector(".hero__content");
    const aboutImg = document.querySelector(".about__img-wrap");
    const aboutCard = document.getElementById("about-card");

    let ticking = false;

    function onScroll() {
      if (ticking) return;
      ticking = true;
      requestAnimationFrame(() => {
        const y = window.scrollY;
        const vh = window.innerHeight;

        if (heroBg) {
          heroBg.style.transform = `translate3d(0, ${y * 0.42}px, 0) scale(1.08)`;
        }

        if (heroContent && y < vh) {
          const p = Math.min(1, y / (vh * 0.65));
          heroContent.style.opacity = String(1 - p * 0.85);
          heroContent.style.transform = `translate3d(0, ${y * 0.18}px, 0)`;
        }

        if (aboutImg) {
          const rect = aboutImg.getBoundingClientRect();
          const center = rect.top + rect.height / 2 - vh / 2;
          aboutImg.style.transform = `translate3d(0, ${center * -0.06}px, 0)`;
        }

        if (aboutCard) {
          const rect = aboutCard.getBoundingClientRect();
          const center = rect.top + rect.height / 2 - vh / 2;
          aboutCard.style.transform = `translate3d(0, ${center * -0.04}px, 0)`;
        }

        document.querySelectorAll("[data-parallax]").forEach((el) => {
          const speed = parseFloat(el.dataset.parallax) || 0.1;
          const rect = el.getBoundingClientRect();
          const offset = (rect.top + rect.height / 2 - vh / 2) * speed;
          el.style.transform = `translate3d(0, ${offset}px, 0)`;
        });

        ticking = false;
      });
    }

    window.addEventListener("scroll", onScroll, { passive: true });
    onScroll();
  }

  function revealInViewport(el) {
    const rect = el.getBoundingClientRect();
    const vh = window.innerHeight || document.documentElement.clientHeight;
    return rect.top < vh * 0.92 && rect.bottom > vh * 0.08;
  }

  function initScrollReveal() {
    const staggerTargets = document.querySelectorAll(".reveal-stagger");
    const revealTargets = document.querySelectorAll(".reveal");

    if (prefersReducedMotion) {
      staggerTargets.forEach((el) => el.classList.add("is-visible"));
      revealTargets.forEach((el) => el.classList.add("is-visible"));
      return;
    }

    const markVisible = (el, observer) => {
      el.classList.add("is-visible");
      observer.unobserve(el);
    };

    const staggerObserver = new IntersectionObserver(
      (entries) => {
        entries.forEach((entry) => {
          if (entry.isIntersecting) markVisible(entry.target, staggerObserver);
        });
      },
      { threshold: 0, rootMargin: "0px 0px 0px 0px" }
    );

    const revealObserver = new IntersectionObserver(
      (entries) => {
        entries.forEach((entry) => {
          if (entry.isIntersecting) markVisible(entry.target, revealObserver);
        });
      },
      { threshold: 0.12, rootMargin: "0px 0px -40px 0px" }
    );

    staggerTargets.forEach((el) => {
      if (revealInViewport(el)) {
        el.classList.add("is-visible");
        return;
      }
      staggerObserver.observe(el);
    });

    revealTargets.forEach((el) => {
      if (revealInViewport(el)) {
        el.classList.add("is-visible");
        return;
      }
      revealObserver.observe(el);
    });
  }

  function initCardTilt(root) {
    if (prefersReducedMotion || window.matchMedia("(max-width: 768px)").matches) {
      return;
    }

    const scope = root || document;
    const cards = scope.querySelectorAll(
      ".fx-card:not([data-tilt-bound])"
    );

    cards.forEach((card) => {
      card.dataset.tiltBound = "1";

      card.addEventListener("mousemove", (e) => {
        const rect = card.getBoundingClientRect();
        const x = (e.clientX - rect.left) / rect.width - 0.5;
        const y = (e.clientY - rect.top) / rect.height - 0.5;
        const isGallery = card.classList.contains("gallery__item");
        const lift = isGallery ? 0 : -6;
        const tilt = isGallery ? 5 : 7;

        card.style.transform =
          `perspective(900px) rotateY(${x * tilt}deg) rotateX(${-y * tilt}deg) translateY(${lift}px)`;
      });

      card.addEventListener("mouseleave", () => {
        card.style.transform = "";
      });
    });
  }

  function markRevealTargets() {
    document.querySelectorAll(".section-head").forEach((el) => {
      el.classList.add("reveal");
    });

    document.querySelectorAll(".about__text").forEach((el) => {
      el.classList.add("reveal");
    });
    document.querySelector(".about__visual")?.classList.add("reveal");

    ["#menu-filters", "#menu-grid", "#gallery-grid", "#reviews-grid"].forEach(
      (sel) => {
        const el = document.querySelector(sel);
        if (el) el.classList.add("reveal-stagger");
      }
    );

    document.querySelector(".visit__info")?.classList.add("reveal");
    document.querySelector(".visit__map-frame")?.classList.add("reveal");
    document.querySelector(".reserve__inner")?.classList.add("reveal");
  }

  window.initScrollEffects = function () {
    markRevealTargets();
    initParallax();
    initScrollReveal();
    initCardTilt();
  };

  window.refreshCardTilt = initCardTilt;
})();
