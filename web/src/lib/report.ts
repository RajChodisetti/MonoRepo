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
  /** Strict visual quality 0–100 (typical 20–60). */
  websiteQualityScore?: number;
  websiteReview?: string;
};

/** Format metric score for UI (supports Go numeric score/max or legacy "12/20" strings). */
export function formatMetricScore(metric: HealthMetric): string {
  if (typeof metric.score === "number" && typeof metric.max === "number") {
    return `${metric.score}/${metric.max}`;
  }
  return String(metric.score);
}
