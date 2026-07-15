import type { Metadata } from "next";
import {
  LegalList,
  LegalPage,
  LegalSection,
} from "@/components/legal/LegalPage";

const updated = "14 July 2026";

export const metadata: Metadata = {
  title: "Privacy Policy | Tuvi Solutions",
  description:
    "How Tuvi Solutions collects, uses, protects, and shares information, including data processed through Google APIs.",
  alternates: { canonical: "/privacy" },
  openGraph: {
    title: "Privacy Policy | Tuvi Solutions",
    description:
      "How Tuvi Solutions collects, uses, protects, and shares information, including data processed through Google APIs.",
    type: "website",
    url: "/privacy",
  },
  robots: { index: true, follow: true },
};

export default function PrivacyPage() {
  return (
    <LegalPage
      eyebrow="Legal"
      title="Privacy Policy"
      summary="This policy explains what information Tuvi Solutions handles, why we handle it, and the choices available to you."
      updated={updated}
    >
      <LegalSection title="1. Who we are and what this policy covers">
        <p>
          Tuvi Solutions (&quot;Tuvi&quot;, &quot;we&quot;, &quot;us&quot;, or
          &quot;our&quot;) builds websites,
          applications, AI assistants, and business automation. This policy applies to
          our public website, consultation and callback services, voice assistant,
          business outreach, and Google-connected email functionality.
        </p>
      </LegalSection>

      <LegalSection title="2. Information we collect">
        <LegalList>
          <li>
            <strong className="text-ink">Consultation information:</strong> your name,
            email address, phone number, selected appointment time, confirmation details,
            and communications about the appointment.
          </li>
          <li>
            <strong className="text-ink">Voice and callback information:</strong> your
            name and phone number, microphone audio while a voice session is active,
            conversation transcripts or interaction records, and details needed to
            complete a booking or callback.
          </li>
          <li>
            <strong className="text-ink">Business outreach information:</strong>
            business contact details, restaurant or company information, outreach
            status, delivery events, responses, and opt-out preferences. This
            information may come from public sources, service providers, or directly
            from the business contact.
          </li>
          <li>
            <strong className="text-ink">Technical information:</strong> standard
            hosting and security logs may include an IP address, browser or device
            details, requested pages, timestamps, and error information.
          </li>
        </LegalList>
        <p>
          We do not currently use advertising cookies on this website. Outreach emails
          may use time-limited open or link tracking and always provide an unsubscribe
          option where required.
        </p>
      </LegalSection>

      <LegalSection title="3. Google API data and Gmail access">
        <p>
          Our outreach system uses the Google Gmail API only after an authorized Google
          Workspace mailbox owner or administrator grants access. The application
          requests the narrow Gmail send permission (
          <code className="rounded bg-surface px-1.5 py-0.5 text-sm text-ink">
            gmail.send
          </code>
          ) so it can send approved messages from that mailbox over Google&apos;s HTTPS API.
        </p>
        <LegalList>
          <li>
            We process the authorized mailbox address, recipient address, message
            subject and body, Google&apos;s delivery identifier, and OAuth credentials
            needed to perform the send.
          </li>
          <li>
            The application does not request permission to read or download inbox
            messages, contacts, Google Drive files, or other Google account content.
          </li>
          <li>
            Refresh tokens are kept in protected server configuration. Short-lived
            access tokens are used only to call Google&apos;s token and Gmail endpoints and
            are cached in server memory for their valid lifetime.
          </li>
          <li>
            Google user data is not sold, used for advertising, or used to train
            generalized AI or machine-learning models.
          </li>
        </LegalList>
        <p>
          Tuvi&apos;s use and transfer of information received from Google APIs adheres to
          the{" "}
          <a
            href="https://developers.google.com/terms/api-services-user-data-policy"
            target="_blank"
            rel="noopener noreferrer"
            className="font-semibold text-primary underline decoration-primary/30 underline-offset-4 transition hover:text-accent"
          >
            Google API Services User Data Policy
          </a>
          , including its Limited Use requirements.
        </p>
      </LegalSection>

      <LegalSection title="4. How we use information">
        <LegalList>
          <li>Provide, operate, secure, and improve our website and services.</li>
          <li>Schedule consultations, place requested callbacks, and send confirmations.</li>
          <li>
            Prepare and send reviewed business communications and maintain suppression
            and unsubscribe records.
          </li>
          <li>Diagnose errors, prevent abuse, and protect our systems and users.</li>
          <li>Meet legal, accounting, compliance, and audit obligations.</li>
        </LegalList>
      </LegalSection>

      <LegalSection title="5. When we share information">
        <p>
          We share information only as needed to provide and protect the services, such
          as with hosting, database, communications, calendar, voice, and email
          providers acting for us. Google receives the data required to authenticate
          the authorized mailbox and send messages through Gmail. We may also disclose
          information when required by law, to protect rights or safety, or as part of a
          business transaction subject to appropriate safeguards.
        </p>
        <p>
          We do not sell personal information or Google user data. Access by personnel
          is limited to people who need it to operate, support, secure, or comply with
          legal obligations for the service, or where the user has given permission.
        </p>
      </LegalSection>

      <LegalSection title="6. Retention and deletion">
        <p>
          We keep consultation, outreach, delivery, security, and opt-out records only
          for as long as reasonably needed for the purposes described in this policy,
          including providing the service, honoring suppression requests, resolving
          disputes, maintaining audit records, and meeting legal obligations. We then
          delete or de-identify the information where practicable.
        </p>
        <p>
          OAuth credentials remain configured only while a mailbox integration is
          active. An authorized mailbox owner or administrator may revoke access in
          their Google Account and may ask us to disconnect the mailbox and remove the
          stored credential. Revocation prevents future Gmail API access, although
          limited delivery and audit records may be retained for the purposes above.
        </p>
      </LegalSection>

      <LegalSection title="7. Your choices and rights">
        <LegalList>
          <li>
            Use the unsubscribe link in an outreach email to stop future marketing
            messages to that address.
          </li>
          <li>
            Revoke Tuvi&apos;s Google access from the security settings in your Google
            Account.
          </li>
          <li>
            Contact us to request access, correction, deletion, or restriction of
            personal information. Applicable law may provide additional rights.
          </li>
        </LegalList>
      </LegalSection>

      <LegalSection title="8. Security">
        <p>
          We use technical and organizational safeguards designed to protect
          information, including access controls, server-side credential storage,
          encrypted HTTPS connections, and restricted operational access. No system is
          completely secure, so we cannot guarantee absolute security.
        </p>
      </LegalSection>

      <LegalSection title="9. International processing">
        <p>
          Our service providers may process information in countries other than the one
          where you live. Where required, we use appropriate safeguards for those
          transfers and continue to protect the information as described here.
        </p>
      </LegalSection>

      <LegalSection title="10. Children">
        <p>
          Our services are intended for businesses and adults. We do not knowingly
          collect personal information from children under 13 through these services.
        </p>
      </LegalSection>

      <LegalSection title="11. Changes to this policy">
        <p>
          We may update this policy as our services or legal obligations change. We
          will post the revised version here and update the date above. Material changes
          will be communicated when required by law.
        </p>
      </LegalSection>
    </LegalPage>
  );
}
