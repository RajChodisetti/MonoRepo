export type HealthMetric = {
  key?: string;
  label: string;
  status: string;
  statusColor: string;
  score: string | number;
  max?: number;
  value: number;
};

export type CompetitorRow = {
  rank: string;
  name: string;
  rating: string;
  score: string;
  scoreColor: string;
  highlight: boolean;
};

export type ReportIssue = {
  title: string;
  description: string;
};

export type RecentReview = {
  author?: string;
  text?: string;
  rating?: number;
  relativeTime?: string;
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

/** Format metric score for UI. Prefer status-only displays — avoid x/y fractions. */
export function formatMetricScore(metric: HealthMetric): string {
  return metric.status || "";
}
