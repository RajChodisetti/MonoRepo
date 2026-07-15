import type { Metadata } from "next";
import Link from "next/link";
import Nav from "@/components/layout/Nav";
import Footer from "@/components/layout/Footer";
import { getContactEmail } from "@/lib/env";

export const metadata: Metadata = {
  title: "Tuvi Outreach | Google Workspace Integration",
  description:
    "How Tuvi Outreach uses the Gmail API to send operator-approved email from authorized Tuvi Solutions Workspace mailboxes.",
  alternates: { canonical: "/google-workspace" },
  openGraph: {
    title: "Tuvi Outreach | Google Workspace Integration",
    description:
      "How Tuvi Outreach uses the Gmail API to send operator-approved email from authorized Tuvi Solutions Workspace mailboxes.",
    type: "website",
    url: "/google-workspace",
  },
  robots: { index: true, follow: true },
};

const cards = [
  {
    title: "One narrow permission",
    body: "The integration requests only gmail.send so it can submit an approved outbound message to Gmail over HTTPS.",
  },
  {
    title: "No mailbox reading",
    body: "It does not request access to read inbox messages, contacts, attachments, message history, or Google Drive files.",
  },
  {
    title: "Controlled access",
    body: "Mailbox connections are provisioned only for authorized Tuvi Workspace owners or administrators. There is no public signup.",
  },
] as const;

export default function GoogleWorkspacePage() {
  const contactEmail = getContactEmail();

  return (
    <>
      <Nav />
      <main id="main-content" tabIndex={-1} className="relative min-h-screen overflow-hidden pt-24">
        <div className="pointer-events-none absolute inset-0 grid-bg opacity-35" />
        <div className="pointer-events-none absolute left-0 top-28 h-80 w-80 -translate-x-1/3 rounded-full bg-sage/70 blur-[110px]" />
        <div className="pointer-events-none absolute right-0 top-96 h-72 w-72 translate-x-1/3 rounded-full bg-parchment blur-[100px]" />

        <div className="relative mx-auto max-w-5xl px-5 py-14 md:px-8 md:py-24">
          <header className="max-w-3xl">
            <p className="text-xs font-bold uppercase tracking-[0.16em] text-primary">
              Google Workspace integration
            </p>
            <h1 className="mt-3 font-display text-4xl font-bold tracking-tight text-ink md:text-6xl">
              Tuvi Outreach
            </h1>
            <p className="mt-6 text-lg leading-8 text-muted">
              Tuvi Outreach is a Tuvi Solutions-operated tool for sending
              operator-approved email from authorized Tuvi Google Workspace mailboxes
              through the Gmail API.
            </p>
          </header>

          <section className="mt-12 grid gap-5 md:grid-cols-3" aria-label="Access summary">
            {cards.map((card) => (
              <article
                key={card.title}
                className="rounded-3xl border border-border bg-bg-elevated p-6 shadow-sm"
              >
                <h2 className="font-display text-lg font-bold text-ink">{card.title}</h2>
                <p className="mt-3 text-sm leading-6 text-muted">{card.body}</p>
              </article>
            ))}
          </section>

          <section className="mt-12 rounded-3xl border border-border bg-bg-elevated p-7 md:p-10">
            <h2 className="font-display text-2xl font-bold text-ink">How access is used</h2>
            <ol className="mt-5 grid gap-5 text-sm leading-7 text-muted md:grid-cols-3">
              <li>
                <span className="block font-display text-2xl font-bold text-primary">01</span>
                An authorized mailbox owner or administrator grants the Gmail send
                permission through Google OAuth.
              </li>
              <li>
                <span className="block font-display text-2xl font-bold text-primary">02</span>
                Tuvi exchanges the authorization for a short-lived access token and
                sends only the message prepared for delivery.
              </li>
              <li>
                <span className="block font-display text-2xl font-bold text-primary">03</span>
                Tuvi records delivery metadata for audit, security, and suppression
                purposes. Access can be revoked from the Google Account at any time.
              </li>
            </ol>
          </section>

          <section className="mt-8 rounded-3xl border border-primary/20 bg-primary/5 p-7 md:p-10">
            <h2 className="font-display text-2xl font-bold text-ink">Responsible use</h2>
            <p className="mt-3 max-w-3xl leading-7 text-muted">
              The integration is intended only for communications that Tuvi and the
              authorized sender are permitted to send. It must not be used for spam,
              unsolicited bulk commercial email, or to evade Google sending limits,
              filters, or abuse protections.
            </p>
          </section>

          <section className="mt-8 flex flex-col gap-5 rounded-3xl border border-border bg-bg-elevated p-7 shadow-sm md:flex-row md:items-center md:justify-between md:p-10">
            <div>
              <h2 className="font-display text-2xl font-bold text-ink">Data and support</h2>
              <p className="mt-2 max-w-2xl leading-7 text-muted">
                Review how Google user data is handled, how to revoke access, and how to
                request deletion. For support, email{" "}
                <a
                  href={`mailto:${contactEmail}`}
                  className="font-semibold text-primary transition hover:text-accent"
                >
                  {contactEmail}
                </a>
                .
              </p>
            </div>
            <div className="flex shrink-0 flex-wrap gap-3">
              <Link
                href="/privacy"
                className="rounded-full bg-ink px-5 py-3 text-sm font-semibold text-white transition hover:bg-primary"
              >
                Privacy Policy
              </Link>
              <Link
                href="/terms"
                className="rounded-full border border-border px-5 py-3 text-sm font-semibold text-ink transition hover:border-primary/40 hover:text-primary"
              >
                Terms of Service
              </Link>
            </div>
          </section>
        </div>
      </main>
      <Footer />
    </>
  );
}
