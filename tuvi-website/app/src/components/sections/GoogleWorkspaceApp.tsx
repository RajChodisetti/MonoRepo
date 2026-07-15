import Link from "next/link";
import { siteContent } from "@/content/site";
import ServiceIcon from "@/components/ui/ServiceIcon";

export default function GoogleWorkspaceApp() {
  const app = siteContent.oauthApp;

  return (
    <section
      id="google-workspace-app"
      aria-labelledby="google-workspace-app-title"
      className="relative scroll-mt-28 px-5 py-10 md:px-8 md:py-12"
    >
      <div className="mx-auto max-w-6xl rounded-3xl border border-border bg-bg-elevated px-5 py-6 shadow-[0_18px_50px_-42px_rgba(15,39,31,0.45)] sm:px-7 sm:py-7 lg:flex lg:items-center lg:justify-between lg:gap-10">
        <div className="flex min-w-0 items-start gap-4 sm:gap-5">
          <span className="flex h-11 w-11 shrink-0 items-center justify-center rounded-2xl bg-sage/65 text-primary" aria-hidden="true">
            <ServiceIcon name="mail" className="h-5 w-5" />
          </span>
          <div className="min-w-0 max-w-3xl">
            <p className="text-[11px] font-bold uppercase tracking-[0.14em] text-primary">
              Google Workspace application · Operated by Tuvi Solutions
            </p>
            <h2
              id="google-workspace-app-title"
              className="mt-2 font-display text-xl font-semibold tracking-[-0.02em] text-ink sm:text-2xl"
            >
              <span className="italic">{app.name}</span> — private Workspace email delivery
            </h2>
            <p className="mt-2 text-sm leading-6 text-muted">
              Authorized Tuvi team members use {app.name} to send individually reviewed
              business email through the Gmail API over HTTPS. It uses only{" "}
              <code className="rounded bg-surface px-1.5 py-0.5 font-mono text-[0.8rem] text-ink">
                gmail.send
              </code>
              , cannot read inbox messages or other Google account content, and has no
              public signup.
            </p>
          </div>
        </div>

        <div className="mt-5 flex flex-wrap items-center gap-x-5 gap-y-3 border-t border-border pt-5 lg:mt-0 lg:shrink-0 lg:border-l lg:border-t-0 lg:pl-8 lg:pt-0">
          <Link
            href="/google-workspace"
            className="inline-flex items-center gap-2 text-sm font-semibold text-primary transition-colors duration-200 hover:text-accent"
          >
            App details <span aria-hidden="true">→</span>
          </Link>
          <Link
            href="/privacy"
            className="text-sm font-semibold text-muted transition-colors duration-200 hover:text-ink"
          >
            Privacy
          </Link>
          <Link
            href="/terms"
            className="text-sm font-semibold text-muted transition-colors duration-200 hover:text-ink"
          >
            Terms
          </Link>
        </div>
      </div>
    </section>
  );
}
