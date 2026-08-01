import type { FoodieContent } from "../lib/foodieContent";
import { telHref } from "@/lib/reservation";

function LogoMark() {
  return (
    <span className="foodie-logo-mark" aria-hidden="true">
      <svg viewBox="0 0 24 24" fill="none">
        <circle cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="1.6" />
        <path
          d="M8 12c0-2.2 1.8-4 4-4s4 1.8 4 4"
          stroke="currentColor"
          strokeWidth="1.6"
          strokeLinecap="round"
        />
        <path
          d="M12 5.5v3M9.5 6.2l.9 2.6M14.5 6.2l-.9 2.6"
          stroke="currentColor"
          strokeWidth="1.4"
          strokeLinecap="round"
        />
      </svg>
    </span>
  );
}

export default function FoodieFooter({
  brand,
  footer,
  contact,
}: {
  brand: FoodieContent["brand"];
  footer: FoodieContent["footer"];
  contact: FoodieContent["contact"];
}) {
  const year = new Date().getFullYear();

  return (
    <footer className="foodie-footer">
      {/* eslint-disable-next-line @next/next/no-img-element */}
      <img className="foodie-footer-garnish foodie-footer-garnish-basil-l" src="/foodie/basil.png" alt="" aria-hidden="true" />
      {/* eslint-disable-next-line @next/next/no-img-element */}
      <img className="foodie-footer-garnish foodie-footer-garnish-chili" src="/foodie/chilies.png" alt="" aria-hidden="true" />
      {/* eslint-disable-next-line @next/next/no-img-element */}
      <img className="foodie-footer-garnish foodie-footer-garnish-basil-r" src="/foodie/menu-basil.png" alt="" aria-hidden="true" />

      <div className="foodie-container foodie-footer-grid">
        <div className="foodie-footer-brand">
          <a href="#home" className="foodie-logo" aria-label={`${brand.name} home`}>
            <LogoMark />
            <span className="foodie-logo-text">{brand.name}</span>
          </a>
          <p>{footer.tagline}</p>
        </div>

        <div className="foodie-footer-col">
          <h4>Quick Links</h4>
          <nav className="foodie-footer-links" aria-label="Footer">
            {footer.links.map((link) => (
              <a key={link.href} href={link.href}>
                {link.label}
              </a>
            ))}
          </nav>
        </div>

        <div className="foodie-footer-col">
          <h4>Contact</h4>
          <p className="foodie-footer-meta">{contact.address}</p>
          <a className="foodie-footer-meta-link" href={telHref(contact.phone)}>
            {contact.phone}
          </a>
          <a className="foodie-footer-meta-link" href={`mailto:${contact.email}`}>
            {contact.email}
          </a>
        </div>
      </div>

      <div className="foodie-footer-bottom">
        <div className="foodie-container">
          <p>
            © {year} {brand.name} · Powered by Tuvi
          </p>
        </div>
      </div>
    </footer>
  );
}
