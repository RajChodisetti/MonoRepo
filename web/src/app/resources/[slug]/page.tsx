import type { Metadata } from "next";
import { notFound } from "next/navigation";
import ResourceGuideView from "@/components/resources/ResourceGuideView";
import { getResourceGuide, resourceGuideSlugs } from "@/content/resourceGuides";

type PageProps = {
  params: Promise<{ slug: string }>;
};

export function generateStaticParams() {
  return resourceGuideSlugs.map((slug) => ({ slug }));
}

export async function generateMetadata({ params }: PageProps): Promise<Metadata> {
  const { slug } = await params;
  const guide = getResourceGuide(slug);
  if (!guide) return { title: "Resources | Tuvi" };
  return {
    title: `${guide.title} | Tuvi`,
    description: guide.description,
  };
}

export default async function ResourceGuidePage({ params }: PageProps) {
  const { slug } = await params;
  const guide = getResourceGuide(slug);
  if (!guide) notFound();
  return <ResourceGuideView guide={guide} />;
}
