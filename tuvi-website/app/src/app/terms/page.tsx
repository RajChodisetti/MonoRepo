import type { Metadata } from "next";
import {
  LegalList,
  LegalPage,
  LegalSection,
} from "@/components/legal/LegalPage";

const updated = "14 July 2026";

export const metadata: Metadata = {
  title: "Terms of Service | Tuvi Solutions",
  description:
    "Terms governing use of the Tuvi Solutions website, consultations, voice assistant, and connected services.",
  alternates: { canonical: "/terms" },
  openGraph: {
    title: "Terms of Service | Tuvi Solutions",
    description:
      "Terms governing use of the Tuvi Solutions website, consultations, voice assistant, and connected services.",
    type: "website",
    url: "/terms",
  },
  robots: { index: true, follow: true },
};

export default function TermsPage() {
  return (
    <LegalPage
      eyebrow="Legal"
      title="Terms of Service"
      summary="These terms govern your use of the Tuvi Solutions website and its consultation, callback, voice, and connected service features."
      updated={updated}
    >
      <LegalSection title="1. Acceptance of these terms">
        <p>
          By accessing or using the Tuvi Solutions website or its features, you agree to
          these Terms of Service (&quot;Terms&quot;). If you use the services for an organization,
          you confirm that you have authority to accept these Terms for that
          organization. If you do not agree, do not use the services.
        </p>
      </LegalSection>

      <LegalSection title="2. Website and services">
        <p>
          Tuvi Solutions provides information about software, web, application, AI, and
          automation services. The website may allow you to book a consultation,
          request a callback, interact with an AI voice assistant, or use connected
          communications features.
        </p>
        <p>
          Website information and an initial consultation are not a commitment to
          deliver a project. Any paid work, deliverables, fees, timelines, warranties,
          ownership terms, and support obligations will be governed by a separate
          written proposal or service agreement.
        </p>
      </LegalSection>

      <LegalSection title="3. Consultations, callbacks, and AI interactions">
        <LegalList>
          <li>Provide accurate contact and scheduling information.</li>
          <li>
            Do not submit passwords, payment-card data, health information, or other
            highly sensitive information through forms or the voice assistant.
          </li>
          <li>
            The voice assistant is an AI system and may misunderstand a request or
            produce an incomplete response. Confirm important details with a person.
          </li>
          <li>
            Callback availability, consultation times, and confirmations may depend on
            third-party communications and calendar services.
          </li>
        </LegalList>
      </LegalSection>

      <LegalSection title="4. Google-connected functionality">
        <p>
          Google Workspace mailbox connections may be used only by an authorized
          mailbox owner or administrator. A connected mailbox authorizes Tuvi to use
          the Gmail API to send approved messages; it does not give Tuvi permission to
          read the mailbox, contacts, or Google Drive files. The owner may revoke access
          through their Google Account at any time.
        </p>
        <p>
          Use of Google-connected functionality is also subject to applicable Google
          terms and policies. Our handling of Google user data is described in our
          Privacy Policy.
        </p>
        <p>
          Google-connected sending must not be used for spam, unsolicited bulk
          commercial email, or to circumvent Gmail limits, filters, or abuse controls.
          Where Google policy or applicable law requires recipient consent, the sender
          must obtain and retain that consent before sending.
        </p>
      </LegalSection>

      <LegalSection title="5. Communications and opt-out">
        <p>
          If you request a consultation or callback, you agree that we may contact you
          about that request. Business outreach recipients may unsubscribe using the
          link in an applicable email. Service, security, and transaction messages may
          still be sent when necessary and legally permitted.
        </p>
      </LegalSection>

      <LegalSection title="6. Acceptable use">
        <p>You must not:</p>
        <LegalList>
          <li>Use the services unlawfully, fraudulently, or to harm another person.</li>
          <li>
            Probe, disrupt, overload, bypass security, or attempt unauthorized access to
            the website, accounts, infrastructure, or data.
          </li>
          <li>
            Upload malicious code, impersonate another person, or submit information
            you are not authorized to provide.
          </li>
          <li>
            Copy, scrape, or exploit the services in a way that infringes rights or
            materially interferes with their operation.
          </li>
        </LegalList>
      </LegalSection>

      <LegalSection title="7. Intellectual property">
        <p>
          The website, branding, design, text, software, and other materials supplied by
          Tuvi are owned by Tuvi Solutions or its licensors and are protected by
          applicable intellectual-property laws. These Terms give you a limited,
          revocable, non-transferable right to use the public website for its intended
          business purpose. Rights in client deliverables are defined in the applicable
          service agreement.
        </p>
      </LegalSection>

      <LegalSection title="8. Third-party services and links">
        <p>
          The services may rely on or link to third-party platforms such as Google,
          communications providers, calendar services, or hosting providers. Their
          services are governed by their own terms and privacy practices. Tuvi is not
          responsible for third-party services outside our control.
        </p>
      </LegalSection>

      <LegalSection title="9. Availability and changes">
        <p>
          We may update, suspend, restrict, or discontinue all or part of the website or
          a feature for maintenance, security, legal, or business reasons. We do not
          promise uninterrupted or error-free availability.
        </p>
      </LegalSection>

      <LegalSection title="10. Disclaimers">
        <p>
          To the extent permitted by law, the public website and free consultation
          features are provided &quot;as is&quot; and &quot;as available&quot; without implied guarantees
          of fitness, merchantability, non-infringement, or a particular business
          outcome. Nothing in these Terms excludes a guarantee, warranty, or consumer
          right that cannot legally be excluded.
        </p>
      </LegalSection>

      <LegalSection title="11. Limitation of liability">
        <p>
          To the extent permitted by law, Tuvi Solutions will not be liable for
          indirect, incidental, special, consequential, or lost-profit damages arising
          from use of the public website or free features. Any liability connected to a
          paid project is governed by the applicable service agreement. Mandatory legal
          rights and liabilities remain unaffected.
        </p>
      </LegalSection>

      <LegalSection title="12. Suspension or termination">
        <p>
          We may restrict or end access where we reasonably believe these Terms have
          been breached or access creates a security, legal, or operational risk. You
          may stop using the public services at any time and may ask us to disconnect a
          connected Google mailbox.
        </p>
      </LegalSection>

      <LegalSection title="13. Applicable law">
        <p>
          Applicable law governs these Terms without limiting any mandatory rights you
          have in your place of residence. A separate service agreement may specify the
          governing law and dispute process for paid work.
        </p>
      </LegalSection>

      <LegalSection title="14. Changes to these terms">
        <p>
          We may update these Terms as the services or legal requirements change. The
          revised Terms will be posted here with a new update date. Continued use after
          an update means you accept the revised Terms where permitted by law.
        </p>
      </LegalSection>
    </LegalPage>
  );
}
