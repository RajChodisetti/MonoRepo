import Link from "next/link";
import BrandLogo from "@/components/layout/BrandLogo";
import {
  footerLegalLinks,
  footerProductColumns,
  footerSecondaryColumns,
  type FooterColumn,
} from "@/components/layout/footer.config";

function FooterColumnBlock({ column }: { column: FooterColumn }) {
  return (
    <div>
      <p className="text-[12px] font-medium text-secondary sm:text-[13px]">{column.title}</p>
      <ul className="mt-3 flex flex-col gap-2 sm:mt-3.5 sm:gap-2.5">
        {column.links.map((link) => (
          <li key={link.href + link.label}>
            <Link
              href={link.href}
              className="text-[14px] font-medium text-ink transition-colors hover:text-primary sm:text-[15px]"
            >
              {link.label}
            </Link>
          </li>
        ))}
      </ul>
    </div>
  );
}

export default function SiteFooter() {
  return (
    <footer className="relative z-10 -mt-5 bg-bg sm:-mt-6 md:-mt-7">
      <div className="w-full rounded-t-[28px] border-t border-border bg-bg px-4 pb-10 pt-10 sm:rounded-t-[36px] sm:px-8 sm:pb-12 sm:pt-12 md:rounded-t-[44px] md:px-12 md:pt-14">
        <div className="mx-auto max-w-[1100px]">
          {/* Top: logo + CTAs */}
          <div className="flex flex-col gap-5 sm:flex-row sm:items-center sm:justify-between">
            <BrandLogo priority={false} />

            <div className="flex flex-wrap items-center gap-2.5">
              <Link
                href="#"
                className="inline-flex items-center justify-center rounded-full bg-primary px-5 py-2.5 text-[13px] font-semibold text-bg transition-colors hover:bg-primary-dim sm:px-6 sm:text-[14px]"
              >
                Get a free demo
              </Link>
              <Link
                href="#"
                className="inline-flex items-center justify-center rounded-full bg-surface px-5 py-2.5 text-[13px] font-semibold text-ink transition-colors hover:bg-parchment sm:px-6 sm:text-[14px]"
              >
                See how it works
              </Link>
            </div>
          </div>

          {/* Product columns */}
          <div className="mt-12 grid grid-cols-2 gap-x-6 gap-y-10 sm:mt-14 sm:gap-x-8 md:grid-cols-4 md:gap-y-0">
            {footerProductColumns.map((column) => (
              <FooterColumnBlock key={column.title} column={column} />
            ))}
          </div>

          {/* Secondary columns */}
          <div className="mt-10 grid grid-cols-2 gap-x-6 gap-y-10 sm:mt-12 sm:gap-x-8 md:grid-cols-3 md:gap-y-0 lg:max-w-[75%]">
            {footerSecondaryColumns.map((column) => (
              <FooterColumnBlock key={column.title} column={column} />
            ))}
          </div>

          {/* Legal row */}
          <div className="mt-12 flex flex-col gap-4 border-t border-border pt-6 sm:mt-14 sm:flex-row sm:items-center sm:justify-between sm:gap-6 sm:pt-7">
            <p className="text-[12px] font-medium text-secondary">
              © 2026 Tuvi | All rights reserved
            </p>
            <ul className="flex flex-wrap items-center gap-x-4 gap-y-2 sm:justify-end">
              {footerLegalLinks.map((link) => (
                <li key={link.href}>
                  <Link
                    href={link.href}
                    className="text-[12px] font-medium text-secondary transition-colors hover:text-ink"
                  >
                    {link.label}
                  </Link>
                </li>
              ))}
            </ul>
          </div>
        </div>
      </div>
    </footer>
  );
}
