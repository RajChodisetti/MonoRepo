import Image from "next/image";
import Link from "next/link";

const guides = {
  featured: {
    title: "Interview: How To Build a $7M/year Restaurant Business",
    image: "/guides/interview.jpg",
    href: "/resources/case-studies",
  },
  secondary: [
    {
      title: "Restaurant SEO is Easier Than You Think (3 Big Wins)",
      image: "/guides/seo.jpg",
      href: "/resources/seo-guide",
    },
    {
      title: "Buyer's Guide: The Best Restaurant Website Builders",
      image: "/guides/builders.jpg",
      href: "/resources/website-builders",
    },
  ],
} as const;

function PlayIcon({ className }: { className?: string }) {
  return (
    <svg viewBox="0 0 24 24" className={className} aria-hidden="true">
      <path d="M9 7.2v9.6l8.2-4.8L9 7.2Z" fill="currentColor" />
    </svg>
  );
}

function ArrowIcon() {
  return (
    <svg viewBox="0 0 20 20" className="h-4 w-4" aria-hidden="true">
      <path
        d="M7 4.5 12.5 10 7 15.5"
        fill="none"
        stroke="currentColor"
        strokeWidth="1.8"
        strokeLinecap="round"
        strokeLinejoin="round"
      />
    </svg>
  );
}

function GuideCard({
  title,
  image,
  href,
  size,
}: {
  title: string;
  image: string;
  href: string;
  size: "lg" | "sm";
}) {
  return (
    <Link
      href={href}
      className={`group relative block overflow-hidden rounded-[22px] sm:rounded-[26px] ${
        size === "lg" ? "h-full min-h-[340px]" : "h-full min-h-[160px]"
      }`}
    >
      <Image
        src={image}
        alt=""
        fill
        className="object-cover transition-transform duration-500 group-hover:scale-[1.03]"
        sizes={size === "lg" ? "(max-width: 1024px) 100vw, 60vw" : "(max-width: 1024px) 50vw, 22vw"}
      />
      <div
        className="absolute inset-0 bg-gradient-to-t from-black/70 via-black/20 to-transparent"
        aria-hidden="true"
      />

      <div className="absolute inset-x-0 bottom-0 flex items-end justify-between gap-3 p-4 sm:p-5">
        <p
          className={`max-w-[90%] font-semibold leading-snug tracking-[-0.02em] text-white ${
            size === "lg"
              ? "text-[15px] sm:text-[17px] md:text-[18px]"
              : "text-[13px] sm:text-[14px]"
          }`}
        >
          {title}
        </p>
        {size === "lg" ? (
          <span className="flex h-9 w-9 shrink-0 items-center justify-center rounded-full bg-black/45 text-white backdrop-blur-sm">
            <PlayIcon className="h-4 w-4" />
          </span>
        ) : null}
      </div>
    </Link>
  );
}

export default function GuidesResources() {
  return (
    <section className="bg-parchment px-4 py-14 sm:px-8 sm:py-16 md:px-12 md:py-18">
      <div className="mx-auto max-w-[1100px]">
        <h2 className="max-w-[18ch] font-display text-[clamp(1.65rem,3.2vw,2.45rem)] font-semibold leading-[1.15] tracking-[-0.03em] text-ink">
          See our free guides on growing your restaurant
        </h2>

        <div className="mt-7 grid gap-4 sm:mt-8 sm:gap-5 lg:grid-cols-[1.35fr_1fr]">
          {/* Featured */}
          <GuideCard
            title={guides.featured.title}
            image={guides.featured.image}
            href={guides.featured.href}
            size="lg"
          />

          {/* Right column */}
          <div className="grid gap-4 sm:gap-5">
            <div className="grid grid-cols-2 gap-4 sm:gap-5">
              {guides.secondary.map((guide) => (
                <GuideCard
                  key={guide.title}
                  title={guide.title}
                  image={guide.image}
                  href={guide.href}
                  size="sm"
                />
              ))}
            </div>

            <Link
              href="/resources"
              className="group flex items-center justify-between gap-4 rounded-[22px] bg-sage px-5 py-4 sm:rounded-[26px] sm:px-6 sm:py-5"
            >
              <div className="min-w-0">
                <p className="text-[14px] font-semibold tracking-[-0.015em] text-ink sm:text-[15px]">
                  Learn with Tuvi
                </p>
                <p className="mt-0.5 text-[12px] font-medium leading-snug text-muted sm:text-[13px]">
                  We create free videos for restaurant operators looking to grow
                </p>
              </div>
              <span className="flex h-9 w-9 shrink-0 items-center justify-center rounded-full bg-bg text-ink shadow-[0_2px_8px_rgba(15,39,31,0.08)] transition-transform group-hover:translate-x-0.5">
                <ArrowIcon />
              </span>
            </Link>
          </div>
        </div>
      </div>
    </section>
  );
}
