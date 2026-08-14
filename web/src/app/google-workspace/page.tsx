import type { Metadata } from "next";
import Link from "next/link";
import SiteFooter from "@/components/layout/SiteFooter";
import { getContactEmail } from "@/lib/env";

const app = {
  name: "tuvi",
  description:
    "tuvi is a private, company-operated application from Tuvi Solutions. Authorized Tuvi team members use it to send individually reviewed business email from their Google Workspace mailboxes through the Gmail API over HTTPS.",
  dataUse:
    "Sending mailboxes request gmail.send. A separate inbound mailbox may use gmail.readonly only to capture replies to outreach plus-addresses; sending mailboxes cannot read inbox mail.",
  access:
    "Mailbox access is limited to authorized Tuvi Solutions owners and administrators. There is no public signup.",
  metadataDescription:
    "tuvi uses Google OAuth and the Gmail API to send individually reviewed email and capture replies through a separate inbound mailbox.",
} as const;

export const metadata: Metadata = {
  title: "tuvi | Google Workspace App by Tuvi Solutions",
  description: app.metadataDescription,
  alternates: { canonical: "/google-workspace" },
  openGraph: {
    title: "tuvi | Google Workspace App by Tuvi Solutions",
    description: app.metadataDescription,
    type: "website",
    url: "/google-workspace",
  },
  robots: { index: true, follow: true },
};

const cards = [
  {
    title: "Narrow send permission",
    body: "Sending mailboxes request gmail.send so tuvi can submit an individually reviewed outbound message to Gmail over HTTPS.",
  },
  {
    title: "Inbound replies stay separate",
    body: "A dedicated inbound mailbox may use gmail.readonly only to capture outreach replies. Sending mailboxes do not read inbox mail, contacts, attachments, message history, or Google Drive files.",
  },
  {
    title: "Controlled access",
    body: "Mailbox connections are provisioned only for authorized Tuvi Solutions Workspace owners or administrators. There is no public signup.",
  },
] as const;

export default function GoogleWorkspacePage() {
  const contactEmail = getContactEmail();

  return (
    <>
      <article className="hero-atmosphere relative overflow-hidden px-4 py-14 sm:px-8 sm:py-20 md:px-12 md:py-24">
        <div className="hero-grid pointer-events-none absolute inset-0 opacity-40" />
        <div className="relative mx-auto max-w-5xl">
          <header className="max-w-3xl">
            <p className="text-xs font-semibold uppercase tracking-[0.16em] text-primary">Google Workspace integration</p>
            <h1 className="mt-3 font-display text-[clamp(3rem,7vw,5.5rem)] font-semibold leading-none tracking-[-0.045em] text-ink">{app.name}</h1>
            <p className="mt-6 text-lg leading-8 text-muted">{app.description}</p>
            <p className="mt-4 max-w-2xl text-sm leading-6 text-muted">{app.dataUse} {app.access}</p>
          </header>

          <section className="mt-12 grid gap-5 md:grid-cols-3" aria-label="Access summary">
            {cards.map((card) => (
              <article key={card.title} className="rounded-3xl border border-border bg-bg p-6 shadow-[0_16px_50px_rgba(15,39,31,0.07)]">
                <h2 className="font-display text-xl font-semibold text-ink">{card.title}</h2>
                <p className="mt-3 text-sm leading-6 text-muted">{card.body}</p>
              </article>
            ))}
          </section>

          <section className="mt-12 rounded-3xl border border-border bg-bg p-7 sm:p-10">
            <h2 className="font-display text-3xl font-semibold tracking-[-0.025em] text-ink">How access is used</h2>
            <ol className="mt-6 grid gap-6 text-sm leading-7 text-muted md:grid-cols-3">
              <li><span className="block font-display text-3xl font-semibold text-primary">01</span>An authorized owner or administrator grants Gmail send access to a sending mailbox. A separate inbound mailbox may grant read-only access solely for outreach replies.</li>
              <li><span className="block font-display text-3xl font-semibold text-primary">02</span>tuvi exchanges the authorization for a short-lived access token and sends only the message prepared for delivery.</li>
              <li><span className="block font-display text-3xl font-semibold text-primary">03</span>Tuvi records delivery metadata for audit, security, and suppression purposes. Access can be revoked from the Google Account at any time.</li>
            </ol>
          </section>

          <section className="mt-8 rounded-3xl border border-primary/20 bg-sage p-7 sm:p-10">
            <h2 className="font-display text-3xl font-semibold tracking-[-0.025em] text-ink">Responsible use</h2>
            <p className="mt-3 max-w-3xl leading-7 text-muted">
              tuvi is intended only for communications that Tuvi Solutions and the
              authorized sender are permitted to send. It must not be used for spam,
              unsolicited bulk commercial email, or to evade Google sending limits,
              filters, or abuse protections.
            </p>
          </section>

          <section className="mt-8 flex flex-col gap-5 rounded-3xl border border-border bg-bg p-7 shadow-[0_16px_50px_rgba(15,39,31,0.07)] sm:p-10 md:flex-row md:items-center md:justify-between">
            <div>
              <h2 className="font-display text-3xl font-semibold tracking-[-0.025em] text-ink">Data and support</h2>
              <p className="mt-2 max-w-2xl leading-7 text-muted">
                Review how Google user data is handled, how to revoke access, and how to
                request deletion. For support, email{" "}
                <a href={`mailto:${contactEmail}`} className="font-semibold text-primary transition hover:text-accent">{contactEmail}</a>.
              </p>
            </div>
            <div className="flex shrink-0 flex-wrap gap-3">
              <Link href="/privacy" className="rounded-full bg-primary px-5 py-3 text-sm font-semibold text-bg transition hover:bg-primary-dim">Privacy Policy</Link>
              <Link href="/terms" className="rounded-full border border-border px-5 py-3 text-sm font-semibold text-ink transition hover:border-primary hover:text-primary">Terms of Service</Link>
            </div>
          </section>
        </div>
      </article>
      <SiteFooter />
    </>
  );
}
