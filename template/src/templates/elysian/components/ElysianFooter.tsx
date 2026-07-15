"use client";

import { useState } from "react";
import type { ElysianContent } from "../lib/mapContent";
import ElysianImage from "./ElysianImage";

export default function ElysianFooter({
  name,
  nameAccent,
  footer,
  showInsta,
}: {
  name: string;
  nameAccent: string;
  footer: ElysianContent["footer"];
  showInsta: boolean;
}) {
  const [subscribed, setSubscribed] = useState(false);

  return (
    <footer className="footer">
      <div className="container footer-grid">
        <div className="footer-brand">
          <a href="#home" className="logo">
            {name}
            {nameAccent ? <span>{nameAccent}</span> : null}
          </a>
          <p>{footer.tagline}</p>
          <form
            className="newsletter"
            id="newsletterForm"
            onSubmit={(e) => {
              e.preventDefault();
              setSubscribed(true);
              window.setTimeout(() => setSubscribed(false), 2200);
            }}
          >
            <input type="email" placeholder="Your email address" required />
            <button type="submit" className="btn btn-gold ripple">
              {subscribed ? "Subscribed ✓" : "Subscribe"}
            </button>
          </form>
        </div>
        <div className="footer-links">
          <h4>Quick Links</h4>
          <a href="#about">About</a>
          <a href="#menu">Menu</a>
          <a href="#reservation">Reservations</a>
          <a href="#faq">FAQ</a>
        </div>
        <div className="footer-links">
          <h4>Contact</h4>
          <a href="#contact">Visit Us</a>
          <a href="#reservation">Reserve</a>
        </div>
        {showInsta ? (
          <div className="footer-insta">
            <h4>Gallery</h4>
            <div className="insta-grid">
              {footer.instaImages.map((url, i) => (
                <ElysianImage
                  key={url + i}
                  src={url}
                  alt={`${name} gallery ${i + 1}`}
                  width={300}
                  height={300}
                  sizes="150px"
                />
              ))}
            </div>
          </div>
        ) : null}
      </div>
      <div className="footer-bottom">
        <p>
          &copy; {new Date().getFullYear()} {name}
          {nameAccent ? ` ${nameAccent}` : ""}. All rights reserved.
        </p>
        <p>Crafted with devotion, served with care.</p>
      </div>
    </footer>
  );
}
