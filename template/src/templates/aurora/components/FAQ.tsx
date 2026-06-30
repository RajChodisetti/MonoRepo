"use client";

import { useState } from "react";
import { motion, AnimatePresence } from "framer-motion";
import BlurReveal from "./ui/BlurReveal";
import type { AuroraContent } from "../lib/mapContent";

export default function FAQ({ items }: { items: AuroraContent["faq"] }) {
  const [open, setOpen] = useState<number | null>(0);

  return (
    <section className="aurora-section">
      <div className="aurora-container max-w-3xl">
        <BlurReveal className="text-center">
          <h2 className="aurora-heading text-4xl font-bold text-white">FAQ</h2>
        </BlurReveal>

        <div className="mt-12 space-y-3">
          {items.map((item, i) => (
            <BlurReveal key={item.question} delay={i * 0.05}>
              <div className="overflow-hidden rounded-xl border border-white/10 bg-white/5 backdrop-blur-md">
                <button
                  type="button"
                  className="flex w-full items-center justify-between p-5 text-left"
                  onClick={() => setOpen(open === i ? null : i)}
                >
                  <span className="font-medium text-white">{item.question}</span>
                  <span className="text-purple-400">{open === i ? "−" : "+"}</span>
                </button>
                <AnimatePresence>
                  {open === i && (
                    <motion.div
                      initial={{ height: 0, opacity: 0 }}
                      animate={{ height: "auto", opacity: 1 }}
                      exit={{ height: 0, opacity: 0 }}
                      className="overflow-hidden"
                    >
                      <p className="border-t border-white/10 px-5 py-4 text-sm text-white/60">
                        {item.answer}
                      </p>
                    </motion.div>
                  )}
                </AnimatePresence>
              </div>
            </BlurReveal>
          ))}
        </div>
      </div>
    </section>
  );
}
