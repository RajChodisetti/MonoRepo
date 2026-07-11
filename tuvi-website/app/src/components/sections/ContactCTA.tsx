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
        <div className="ink-panel relative overflow-hidden rounded-[2rem] p-10 text-center md:p-16">
          <span className="relative inline-flex items-center gap-2.5 rounded-full border border-white/20 bg-white/10 px-3.5 py-1.5 text-xs font-semibold text-white">
            <span className="h-2 w-2 rounded-full bg-primary" aria-hidden />
            {contact.eyebrow}
          </span>
          <h2 className="relative mx-auto mt-5 max-w-2xl font-display text-3xl font-bold leading-[1.08] tracking-tight text-white md:text-4xl lg:text-5xl">
            {contact.title}
          </h2>
          <p className="relative mx-auto mt-4 max-w-xl text-base leading-relaxed text-zinc-300 md:text-lg">
            {contact.description}
          </p>

          <div className="relative mt-8 flex flex-col items-center justify-center gap-4 sm:flex-row">
            <a
              href={getBookCallUrl()}
              className="inline-flex items-center justify-center gap-2 rounded-full bg-white px-6 py-3 text-sm font-semibold text-ink shadow-lg transition duration-300 hover:-translate-y-0.5 hover:shadow-xl"
            >
              {contact.primaryCta} <span aria-hidden>→</span>
            </a>
            {callInHref && callInDisplay ? (
              <a
                href={callInHref}
                className="rounded-full border border-white/25 px-5 py-3 text-sm font-semibold text-white transition hover:bg-white/10"
              >
                Call our AI · {callInDisplay}
              </a>
            ) : null}
            <a
              href={`mailto:${getContactEmail()}`}
              className="text-sm text-zinc-300 transition hover:text-white"
            >
              {getContactEmail()}
            </a>
          </div>

          <div className="relative mx-auto mt-9 max-w-md rounded-2xl bg-white p-6 text-left shadow-2xl">
            <RequestCallbackForm />
          </div>
        </div>
      </Reveal>
    </SectionShell>
  );
}
