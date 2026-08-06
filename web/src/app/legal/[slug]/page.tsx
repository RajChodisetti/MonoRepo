import type { Metadata } from "next";
import Link from "next/link";
import { notFound } from "next/navigation";
import SiteFooter from "@/components/layout/SiteFooter";

type LegalPage = {
  slug: string;
  title: string;
  description: string;
  body: string[];
};

const legalPages: LegalPage[] = [
  {
    slug: "cookies",
    title: "Cookie Settings",
    description: "How Tuvi uses cookies on this website.",
    body: [
      "We use essential cookies to run the site and optional analytics cookies to understand what content helps restaurant operators.",
      "You can control non-essential cookies in your browser settings. Continuing to use the site with defaults means essential cookies stay on.",
    ],
  },
  {
    slug: "privacy",
    title: "Privacy",
    description: "How Tuvi handles personal information.",
    body: [
      "We collect details you share through forms (like demo requests) so we can respond and improve our services.",
      "We do not sell your personal information. Contact contact@tuvisolutions.com for privacy questions or access requests.",
    ],
  },
  {
    slug: "website-terms",
    title: "Website Terms",
    description: "Terms for using the Tuvi marketing website.",
    body: [
      "This site is provided for information about Tuvi products and services. Content may change without notice.",
      "Do not misuse the site, attempt unauthorized access, or scrape content in a way that harms our systems.",
    ],
  },
  {
    slug: "disclaimer",
    title: "Disclaimer",
    description: "Important limits on site content and examples.",
    body: [
      "Case studies and metrics on this site include fictional demo venues for illustration. Results vary by restaurant.",
      "Nothing on this website is legal, financial, or professional advice.",
    ],
  },
  {
    slug: "restaurant-agreements",
    title: "Restaurant Agreements",
    description: "Overview of customer agreements for restaurants on Tuvi.",
    body: [
      "Restaurants using Tuvi enter into a services agreement covering platform access, support, and billing.",
      "Ask your Tuvi contact or email contact@tuvisolutions.com for the current agreement package.",
    ],
  },
  {
    slug: "platform-terms",
    title: "Platform Terms",
    description: "High-level terms for using the Tuvi platform.",
    body: [
      "Platform access is for legitimate restaurant operations. You are responsible for your menu, pricing, and guest communications.",
      "We may update platform terms; material changes will be communicated to active customers.",
    ],
  },
  {
    slug: "accessibility",
    title: "Accessibility",
    description: "Our commitment to an accessible Tuvi website.",
    body: [
      "We aim to make this site usable with common assistive technologies and clear visual hierarchy.",
      "If you hit a barrier, email contact@tuvisolutions.com with the page URL and we will work to improve it.",
    ],
  },
];

type PageProps = {
  params: Promise<{ slug: string }>;
};

export function generateStaticParams() {
  return legalPages.map((page) => ({ slug: page.slug }));
}

export async function generateMetadata({ params }: PageProps): Promise<Metadata> {
  const { slug } = await params;
  const page = legalPages.find((p) => p.slug === slug);
  if (!page) return { title: "Legal | Tuvi" };
  return { title: `${page.title} | Tuvi`, description: page.description };
}

export default async function LegalPage({ params }: PageProps) {
  const { slug } = await params;
  const page = legalPages.find((p) => p.slug === slug);
  if (!page) notFound();

  return (
    <>
      <section className="hero-atmosphere relative px-4 pb-10 pt-14 sm:px-8 sm:pt-20 md:px-12">
        <div className="relative z-10 mx-auto max-w-[720px]">
          <Link href="/" className="text-[13px] font-semibold text-primary hover:text-primary-dim">
            ← Home
          </Link>
          <h1 className="mt-5 font-display text-[clamp(2rem,4vw,3rem)] font-semibold tracking-[-0.03em] text-ink">
            {page.title}
          </h1>
          <p className="mt-3 text-[16px] text-muted">{page.description}</p>
        </div>
      </section>
      <section className="px-4 pb-16 sm:px-8 md:px-12">
        <div className="mx-auto max-w-[720px] space-y-5">
          {page.body.map((para) => (
            <p key={para} className="text-[15px] leading-relaxed text-muted sm:text-[16px]">
              {para}
            </p>
          ))}
        </div>
      </section>
      <SiteFooter />
    </>
  );
}
