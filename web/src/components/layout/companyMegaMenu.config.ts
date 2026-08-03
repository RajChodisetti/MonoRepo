export type CompanyLink = {
  label: string;
  href: string;
};

export type CompanyFeaturedCard = {
  href: string;
  title: string;
  variant: "hiring" | "photo";
  imageSrc?: string;
  imageAlt?: string;
};

export const companyLinks: CompanyLink[] = [
  { label: "Our Story", href: "/company/story" },
  { label: "Careers", href: "/company/careers" },
  { label: "Leadership", href: "/company/leadership" },
  { label: "Builders Wanted", href: "/company/builders" },
  { label: "Press", href: "/company/press" },
  { label: "Reviews", href: "/company/reviews" },
  { label: "Partner with Tuvi", href: "/company/partners" },
];

export const companyFeatured: CompanyFeaturedCard[] = [
  {
    href: "/company/careers",
    title: "We're hiring",
    variant: "hiring",
  },
  {
    href: "/company/press/series-c",
    title: "See the memo from Tuvi's $120M Series C raise",
    variant: "photo",
    imageSrc:
      "https://images.unsplash.com/photo-1522071820081-009f0129c71c?auto=format&fit=crop&w=900&q=80",
    imageAlt: "Team members standing together in a modern workspace",
  },
];
