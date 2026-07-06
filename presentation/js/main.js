const nav = document.querySelector(".nav");
const reveals = document.querySelectorAll(".reveal");

const onScroll = () => {
  nav?.classList.toggle("scrolled", window.scrollY > 40);
};

const revealObserver = new IntersectionObserver(
  (entries) => {
    entries.forEach((entry) => {
      if (entry.isIntersecting) {
        entry.target.classList.add("visible");
      }
    });
  },
  { threshold: 0.12, rootMargin: "0px 0px -40px 0px" }
);

reveals.forEach((el) => revealObserver.observe(el));
window.addEventListener("scroll", onScroll, { passive: true });
onScroll();
