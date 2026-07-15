import Link from "next/link";
import { siteContent } from "@/content/site";
import ServiceIcon from "@/components/ui/ServiceIcon";

const appFacts = [
  { label: "Application name", value: "tuvi", code: false },
  { label: "Purpose", value: "Send individually reviewed business email", code: false },
  { label: "Google permission", value: "gmail.send only", code: true },
  {
    label: "Availability",
    value: "Authorized Tuvi Solutions Workspace accounts",
    code: false,
  },
] as const;

export default function GoogleWorkspaceApp() {
  const app = siteContent.oauthApp;

  return (
    <section
      id="google-workspace-app"
      aria-labelledby="google-workspace-app-title"
      className="ink-panel relative scroll-mt-28 overflow-hidden border-y border-white/10 px-5 py-16 text-[#fffef8] md:px-8 md:py-20"
    >
      <div className="pointer-events-none absolute -right-28 -top-32 h-96 w-96 rounded-full border border-white/10" />
      <div className="pointer-events-none absolute -bottom-36 right-20 h-80 w-80 rounded-full border border-white/10" />

      <div className="relative mx-auto grid max-w-6xl gap-10 lg:grid-cols-[1.08fr_0.92fr] lg:items-center lg:gap-14">
        <div className="max-w-2xl">
          <p className="inline-flex items-center gap-2.5 rounded-full border border-white/15 bg-white/[0.06] px-3.5 py-2 text-xs font-semibold uppercase tracking-[0.16em] text-white/75">
            <span className="h-1.5 w-1.5 rounded-full bg-sage" aria-hidden="true" />
            {app.eyebrow}
          </p>

          <div className="mt-6 flex items-center gap-4">
            <span className="flex h-12 w-12 shrink-0 items-center justify-center rounded-2xl bg-white/10 text-sage" aria-hidden="true">
              <ServiceIcon name="mail" />
            </span>
            <p className="text-xs font-semibold uppercase tracking-[0.18em] text-white/55">
              Operated by Tuvi Solutions
            </p>
          </div>

          <h2
            id="google-workspace-app-title"
            className="mt-5 font-display text-4xl font-semibold leading-[1.02] tracking-[-0.035em] text-[#fffef8] sm:text-5xl lg:text-6xl"
          >
            <span className="italic">{app.name}</span> — {app.title}
          </h2>
          <p className="mt-6 max-w-xl text-base leading-7 text-white/75 md:text-lg md:leading-8">
            {app.description}
          </p>
          <p className="mt-4 max-w-xl text-sm leading-6 text-white/65 md:text-base md:leading-7">
            {app.dataUse} {app.access}
          </p>

          <div className="mt-8 flex flex-col gap-3 sm:flex-row sm:flex-wrap">
            <Link
              href="/google-workspace"
              className="inline-flex w-full items-center justify-center gap-2 rounded-full bg-[#fffef8] px-5 py-3 text-sm font-semibold text-ink transition-colors duration-200 hover:bg-sage sm:w-auto"
            >
              How {app.name} uses Google Workspace <span aria-hidden="true">→</span>
            </Link>
            <Link
              href="/privacy"
              className="inline-flex w-full items-center justify-center rounded-full border border-white/20 px-5 py-3 text-sm font-semibold text-[#fffef8] transition-colors duration-200 hover:bg-white/10 sm:w-auto"
            >
              Privacy Policy
            </Link>
            <Link
              href="/terms"
              className="inline-flex w-full items-center justify-center px-4 py-3 text-sm font-semibold text-white/70 transition-colors duration-200 hover:text-white sm:w-auto"
            >
              Terms of Service
            </Link>
          </div>
        </div>

        <div className="rounded-[2rem] border border-white/15 bg-white/[0.07] p-5 shadow-[0_32px_80px_-38px_rgba(0,0,0,0.55)] backdrop-blur-sm sm:p-7">
          <div className="flex items-start justify-between gap-5 border-b border-white/10 pb-5">
            <div>
              <p className="text-xs font-semibold uppercase tracking-[0.16em] text-white/50">
                Application identity
              </p>
              <p className="mt-2 font-display text-3xl font-semibold italic text-[#fffef8]">
                {app.name}
              </p>
            </div>
            <span className="rounded-full border border-sage/30 bg-sage/10 px-3 py-1.5 text-[11px] font-semibold uppercase tracking-[0.1em] text-sage">
              Private access
            </span>
          </div>

          <dl className="divide-y divide-white/10">
            {appFacts.map((fact) => (
              <div key={fact.label} className="grid gap-2 py-4 sm:grid-cols-[9rem_1fr] sm:gap-5">
                <dt className="text-xs font-semibold uppercase tracking-[0.12em] text-white/45">
                  {fact.label}
                </dt>
                <dd className="text-sm font-medium leading-6 text-white/85 sm:text-right">
                  {fact.code ? (
                    <code className="rounded-md bg-white/10 px-2 py-1 font-mono text-[0.8rem] text-sage">
                      {fact.value}
                    </code>
                  ) : (
                    fact.value
                  )}
                </dd>
              </div>
            ))}
          </dl>

          <p className="mt-2 border-t border-white/10 pt-5 text-xs leading-5 text-white/55">
            Users can revoke access at any time through their Google Account. Support is available at{" "}
            <a
              href={`mailto:${siteContent.brand.email}`}
              className="font-semibold text-sage underline decoration-sage/30 underline-offset-4 transition-colors hover:text-white"
            >
              {siteContent.brand.email}
            </a>
            .
          </p>
        </div>
      </div>
    </section>
  );
}
