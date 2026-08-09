import { caseStudies, caseStudyPath } from "@/content/caseStudies";

export type SuccessStory = {
  id: string;
  slug: string;
  href: string;
  name: string;
  business: string;
  metricValue: string;
  metricDescription: string;
  imageUrl: string;
};

/** Homepage carousel — five fictional case studies. */
export const successStories: SuccessStory[] = caseStudies.map((study) => ({
  id: study.slug,
  slug: study.slug,
  href: caseStudyPath(study.slug),
  name: study.name,
  business: `${study.role} of ${study.restaurant}`,
  metricValue: study.metricValue,
  metricDescription: study.metricDescription,
  imageUrl: study.imageUrl,
}));
