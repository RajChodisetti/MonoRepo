"use client";

import { useState } from "react";
import type { ElysianContent } from "../lib/mapContent";

export default function ElysianFAQ({ items }: { items: ElysianContent["faq"] }) {
  const [openIndex, setOpenIndex] = useState<number | null>(null);

  if (!items.length) return null;

  return (
    <section className="faq section" id="faq">
      <div className="container faq-wrap">
        <div className="section-head reveal fade-up">
          <p className="eyebrow">Good to Know</p>
          <h2 className="section-title">
            Frequently Asked <span className="gold-text">Questions</span>
          </h2>
        </div>
        <div className="faq-list reveal fade-up">
          {items.map((item, i) => (
            <div key={item.question} className={`faq-item${openIndex === i ? " open" : ""}`}>
              <button
                type="button"
                className="faq-q"
                onClick={() => setOpenIndex(openIndex === i ? null : i)}
              >
                {item.question}
                <span className="faq-icon">+</span>
              </button>
              <div className="faq-a">
                <p>{item.answer}</p>
              </div>
            </div>
          ))}
        </div>
      </div>
    </section>
  );
}
