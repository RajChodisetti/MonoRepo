"use client";

import { useId, useState } from "react";
import type { ProductFaqItem } from "@/content/products/types";

type ProductFaqProps = {
  items: ProductFaqItem[];
};

export default function ProductFaq({ items }: ProductFaqProps) {
  const baseId = useId();
  const [openIndex, setOpenIndex] = useState<number | null>(null);
  const [hoveredIndex, setHoveredIndex] = useState<number | null>(null);

  return (
    <section className="bg-bg px-5 pb-20 pt-16 sm:px-8 sm:pb-28 sm:pt-20">
      <div className="mx-auto w-full max-w-[780px]">
        <h2 className="text-center font-display text-[2.25rem] font-semibold tracking-tight text-ink sm:text-[2.75rem]">
          FAQ
        </h2>

        <div className="mt-10 space-y-1 sm:mt-14">
          {items.map((item, index) => {
            const isOpen = openIndex === index;
            const isHighlighted = isOpen || hoveredIndex === index;
            const panelId = `${baseId}-panel-${index}`;
            const buttonId = `${baseId}-button-${index}`;

            return (
              <div
                key={item.question}
                onMouseEnter={() => setHoveredIndex(index)}
                onMouseLeave={() => setHoveredIndex(null)}
                className={`rounded-[28px] px-6 transition-[background-color,box-shadow] duration-300 ease-out sm:rounded-[36px] sm:px-8 ${
                  isHighlighted ? "bg-parchment" : "bg-transparent"
                }`}
              >
                <button
                  id={buttonId}
                  type="button"
                  aria-expanded={isOpen}
                  aria-controls={panelId}
                  onClick={() => setOpenIndex(isOpen ? null : index)}
                  className="flex w-full items-start justify-between gap-6 py-6 text-left sm:items-center sm:py-7"
                >
                  <span className="text-[1.25rem] font-semibold leading-snug tracking-tight text-ink sm:text-[1.55rem]">
                    {item.question}
                  </span>
                  <span
                    className="relative mt-1.5 flex h-6 w-6 shrink-0 items-center justify-center text-ink sm:mt-0"
                    aria-hidden="true"
                  >
                    <span className="absolute h-[2px] w-4 rounded-full bg-current" />
                    <span
                      className={`absolute h-4 w-[2px] rounded-full bg-current transition-transform duration-300 ease-out ${
                        isOpen ? "scale-y-0" : "scale-y-100"
                      }`}
                    />
                  </span>
                </button>

                <div
                  id={panelId}
                  role="region"
                  aria-labelledby={buttonId}
                  className={`grid transition-[grid-template-rows] duration-300 ease-out ${
                    isOpen ? "grid-rows-[1fr]" : "grid-rows-[0fr]"
                  }`}
                >
                  <div className="overflow-hidden">
                    <p
                      className={`pb-6 text-[1.05rem] leading-relaxed text-muted transition-opacity duration-300 sm:pb-7 sm:text-[1.15rem] ${
                        isOpen ? "opacity-100" : "opacity-0"
                      }`}
                    >
                      {item.answer}
                    </p>
                  </div>
                </div>
              </div>
            );
          })}
        </div>
      </div>
    </section>
  );
}
