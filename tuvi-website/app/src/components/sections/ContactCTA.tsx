import { siteContent } from "@/content/site";
import { getBookCallUrl, getCallInDisplay, getCallInTelHref, getContactEmail } from "@/lib/env";
import SectionShell from "@/components/ui/SectionShell";
import Reveal from "@/components/ui/Reveal";
import RequestCallbackForm from "@/components/RequestCallbackForm";

export default function ContactCTA() {
  const { contact } = siteContent;
  const callInDisplay = getCallInDisplay();
  const callInHref = getCallInTelHref();

  return (
    <SectionShell id={contact.id}>
      <Reveal>
        <div className="ink-panel relative overflow-hidden rounded-[2rem] p-5 text-center sm:rounded-[2.5rem] sm:p-8 md:p-16">
          <span className="relative inline-flex items-center gap-2.5 rounded-full border border-white/20 bg-white/10 px-3.5 py-1.5 text-xs font-semibold text-white">
            <span className="h-2 w-2 rounded-full bg-primary" aria-hidden />
            {contact.eyebrow}
          </span>
          <h2 className="relative mx-auto mt-5 max-w-2xl font-display text-3xl font-bold leading-[1.08] tracking-tight text-white md:text-4xl lg:text-5xl">
            {contact.title}
          </h2>
          <p className="relative mx-auto mt-4 max-w-xl text-base leading-relaxed text-white/70 md:text-lg">
            {contact.description}
          </p>

          <div className="relative mt-8 flex flex-col items-center justify-center gap-4 sm:flex-row">
            <a
              href={getBookCallUrl()}
              className="inline-flex w-full items-center justify-center gap-2 rounded-full bg-white px-6 py-3 text-sm font-semibold text-black shadow-lg transition-colors duration-200 hover:bg-white/90 sm:w-auto"
            >
              {contact.primaryCta} <span aria-hidden>→</span>
            </a>
            {callInHref && callInDisplay ? (
              <a
                href={callInHref}
                className="w-full rounded-full border border-white/25 px-5 py-3 text-sm font-semibold text-white transition hover:bg-white/10 sm:w-auto"
              >
                Call our AI · {callInDisplay}
              </a>
            ) : null}
            <a
              href={`mailto:${getContactEmail()}`}
              className="max-w-full break-all text-sm text-white/70 transition-colors hover:text-white sm:break-normal"
            >
              {getContactEmail()}
            </a>
          </div>

          <div className="relative mx-auto mt-9 max-w-md rounded-3xl border border-white/10 bg-bg-elevated p-5 text-left shadow-2xl sm:p-6">
            <RequestCallbackForm />
            <p className="mt-4 text-center text-[11px] leading-5 text-muted">
              By submitting a callback request, you agree that Tuvi may contact you about this request. See our{" "}
              <a href="/privacy" className="font-semibold text-primary underline underline-offset-2">
                Privacy Policy
              </a>
              .
            </p>
          </div>
        </div>
      </Reveal>
    </SectionShell>
  );
}
