import type { PlaceAttribution } from "@/lib/places";

export type HealthMetric = {
  key?: string;
  label: string;
  status: string;
  statusColor: string;
  score: string | number;
  max?: number;
  value: number;
  rationale?: string;
  recommendation?: string;
  evidence?: string[];
};

export type CompetitorRow = {
  rank: string | number;
  name: string;
  rating: string | number;
  score: string | number;
  scoreColor: string;
  highlight: boolean;
  placeId?: string;
  distanceKm?: number;
  userRatingCount?: number;
  visibilityScore?: number;
  scoreMax?: number;
  reasons?: string[];
  attributions?: PlaceAttribution[];
};

export type MenuEvidence = {
  status: "present" | "not_found" | "unknown" | string;
  menuUrl?: string;
  hasWebsiteLink?: boolean;
  hasStructuredData?: boolean;
  source?: string;
  rationale?: string;
  explanation?: string;
};

export type SocialProfile = {
  platform: string;
  handle?: string;
  url: string;
  source?: string;
};

export type SocialPresence = {
  status: "present" | "not_found" | "unknown" | string;
  score: number;
  max: number;
  profiles: SocialProfile[];
  rationale?: string;
};

export type CompetitorScan = {
  status: "complete" | "partial" | "unavailable" | string;
  radiusKm?: number;
  cuisine?: string;
  scoreKind?: string;
  sampleSize?: number;
  currentScore?: number;
  currentPosition?: number;
  currentRestaurantLeading?: boolean;
  notice?: string;
  rows?: CompetitorRow[];
};

export type ReportIssue = {
  title: string;
  description: string;
};

export type RecentReview = {
  author?: string;
  authorUri?: string;
  authorPhotoUri?: string;
  googleMapsUri?: string;
  flagContentUri?: string;
  text?: string;
  rating?: number;
  relativeTime?: string;
  publishTime?: string;
  visitDate?: {
    year?: number;
    month?: number;
    day?: number;
  };
  sentiment?: "positive" | "mixed" | "negative" | string;
};

export type RestaurantReport = {
  restaurantName: string;
  address: string;
  overallScore: number;
  overallLabel: string;
  overallColor: string;
  metrics: HealthMetric[];
  competitors: CompetitorRow[];
  competitorScan?: CompetitorScan;
  menuEvidence?: MenuEvidence;
  socialPresence?: SocialPresence;
  issues: ReportIssue[];
  aiSummary?: string;
  estimatedMonthlyLoss: number;
  fullReportLocked?: boolean;
  unlockCta?: string;
  /** Homepage screenshot as data URL (JPEG). */
  websiteScreenshot?: string;
  /** Mobile viewport homepage screenshot for scan phone mockup. */
  websiteMobileScreenshot?: string;
  /** Strict visual quality 0–100 (typical 20–60). */
  websiteQualityScore?: number;
  websiteReview?: string;
  /** "ai-assisted" only when the provider actually contributed. */
  analysisSource?: "ai-assisted" | "automated" | string;
  /** Partial reports use conservative fallbacks for timed-out live signals. */
  analysisStatus?: "complete" | "partial" | string;
  analysisNotice?: string;
  generatedInMs?: number;
  /** Live Google reviews for the scan map UI. */
  recentReviews?: RecentReview[];
};

/** Format a metric's human-readable status for compact UI surfaces. */
export function formatMetricScore(metric: HealthMetric): string {
  return metric.status || "";
}
