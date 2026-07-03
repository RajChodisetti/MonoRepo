import type { Metadata } from "next";
import Nav from "@/components/layout/Nav";
import Footer from "@/components/layout/Footer";
import BookConsultationForm from "@/components/BookConsultationForm";

export const metadata: Metadata = {
  title: "Book a Call | Tuvi Solutions",
  description: "Schedule a free consultation with Tuvi Solutions. Pick a slot and we'll add it to Google Calendar.",
};

export default function BookPage() {
  return (
    <>
      <Nav />
      <main className="relative min-h-screen pt-[72px]">
        <div className="pointer-events-none absolute inset-0 grid-bg opacity-40" />
        <div className="pointer-events-none absolute left-0 top-32 h-72 w-72 -translate-x-1/3 rounded-full bg-gold/10 blur-[100px]" />
        <div className="pointer-events-none absolute bottom-24 right-0 h-64 w-64 translate-x-1/4 rounded-full bg-cyan/10 blur-[90px]" />

        <div className="relative mx-auto max-w-xl px-5 py-12 md:px-8 md:py-16">
          <BookConsultationForm />
        </div>
      </main>
      <Footer />
    </>
  );
}
