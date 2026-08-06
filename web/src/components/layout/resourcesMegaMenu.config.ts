export type ResourcesLink = {
  label: string;
  href: string;
};

export type ResourcesFeaturedCard = {
  href: string;
  title: string;
  variant: "photo" | "cta";
  imageSrc?: string;
  imageAlt?: string;
};

export const resourcesLinks: ResourcesLink[] = [
  { label: "Case Studies", href: "/resources/case-studies" },
  { label: "Blog", href: "/resources/blog" },
  { label: "Help Center", href: "/resources/help" },
  { label: "Restaurant Grader", href: "/resources/grader" },
  { label: "Support Center", href: "/resources/support" },
];

export const resourcesFeatured: ResourcesFeaturedCard[] = [
  {
    href: "/resources/case-studies/orzo-vale-kouzina",
    title: "How Jordan from Orzo Vale Kouzina grew online sales to $40K/month with Tuvi",
    variant: "photo",
    imageSrc: "/resources/resource-ordering-hero.png",
    imageAlt: "Kitchen pass ready for first-party orders",
  },
  {
    href: "/resources/case-studies",
    title: "See all case studies",
    variant: "cta",
  },
];
