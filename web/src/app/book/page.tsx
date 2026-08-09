import type { Metadata } from "next";
import BookConsultationForm from "@/components/booking/BookConsultationForm";
import SiteFooter from "@/components/layout/SiteFooter";

export const metadata: Metadata = {
  title: "Book a Call | Tuvi Solutions",
  description: "Choose an available time for a free consultation with Tuvi Solutions.",
  alternates: { canonical: "/book" },
};

export default function BookPage() {
  return (
    <>
      <section className="hero-atmosphere relative overflow-hidden px-4 py-14 sm:px-8 sm:py-20 md:px-12 md:py-24">
        <div className="hero-grid pointer-events-none absolute inset-0 opacity-50" />
        <div className="relative mx-auto grid max-w-[1100px] items-start gap-10 lg:grid-cols-[minmax(0,0.8fr)_minmax(30rem,1.2fr)] lg:gap-14">
          <header className="pt-3 lg:sticky lg:top-28">
            <p className="text-xs font-semibold uppercase tracking-[0.16em] text-primary">Talk with Tuvi</p>
            <h1 className="mt-4 max-w-[12ch] font-display text-[clamp(2.5rem,6vw,4.8rem)] font-semibold leading-[0.98] tracking-[-0.045em] text-ink">
              Let&apos;s map your next growth move.
            </h1>
            <p className="mt-6 max-w-xl text-base leading-7 text-muted sm:text-lg">
              Pick an open time, tell us how to reach you, and Tuvi will reserve the consultation in its database immediately.
            </p>
            <ul className="mt-7 space-y-3 text-sm text-muted">
              <li className="flex gap-3"><span className="text-primary" aria-hidden="true">✓</span>Review your restaurant&apos;s current online sales flow</li>
              <li className="flex gap-3"><span className="text-primary" aria-hidden="true">✓</span>Identify the highest-leverage website, SEO, or retention opportunity</li>
              <li className="flex gap-3"><span className="text-primary" aria-hidden="true">✓</span>Leave with clear next steps—no obligation</li>
            </ul>
          </header>
          <BookConsultationForm />
        </div>
      </section>
      <SiteFooter />
    </>
  );
}
