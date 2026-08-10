import React from "react";
import {
  Document,
  Image,
  Link,
  Page,
  StyleSheet,
  Text,
  View,
} from "@react-pdf/renderer";
import type { RestaurantDetails } from "@/lib/places";
import type { HealthMetric, RestaurantReport } from "@/lib/report";

const palette = {
  forest: "#174c3a",
  forestSoft: "#e8f1eb",
  orange: "#e86a2d",
  orangeSoft: "#fff1e8",
  ink: "#17211d",
  muted: "#65716b",
  line: "#d9dfdb",
  cream: "#f7f4ef",
  white: "#ffffff",
};

const A4_PAGE = { width: 595.28, height: 841.89 };

const styles = StyleSheet.create({
  page: {
    paddingTop: 26,
    paddingBottom: 24,
    paddingHorizontal: 30,
    backgroundColor: palette.white,
    color: palette.ink,
    fontFamily: "Helvetica",
    fontSize: 9,
    lineHeight: 1.35,
  },
  pageBorder: {
    position: "absolute",
    top: 12,
    right: 12,
    bottom: 12,
    left: 12,
    borderWidth: 1.25,
    borderColor: palette.forest,
    borderRadius: 10,
  },
  header: {
    flexDirection: "row",
    alignItems: "center",
    justifyContent: "space-between",
    paddingBottom: 10,
    borderBottomWidth: 1,
    borderBottomColor: palette.line,
  },
  brand: { flexDirection: "row", alignItems: "center" },
  logo: { width: 30, height: 30, objectFit: "contain", marginRight: 8 },
  brandName: { fontSize: 16, fontFamily: "Helvetica-Bold", color: palette.forest },
  brandLine: { marginTop: 3, fontSize: 7.5, color: palette.muted },
  meta: { textAlign: "right", fontSize: 7.5, color: palette.muted },
  eyebrow: {
    marginTop: 13,
    fontSize: 7.5,
    fontFamily: "Helvetica-Bold",
    textTransform: "uppercase",
    letterSpacing: 1.2,
    color: palette.orange,
  },
  title: {
    marginTop: 3,
    fontSize: 22,
    lineHeight: 1.1,
    fontFamily: "Helvetica-Bold",
    color: palette.ink,
  },
  subtitle: { marginTop: 4, fontSize: 9, color: palette.muted },
  sourceRow: {
    marginTop: 5,
    flexDirection: "row",
    flexWrap: "wrap",
    gap: 7,
    fontSize: 12,
    fontFamily: "Helvetica",
    color: "#5e5e5e",
  },
  sourceLink: { color: "#5e5e5e", textDecoration: "underline" },
  scoreRow: { marginTop: 12, flexDirection: "row", gap: 10 },
  scoreBox: {
    width: 128,
    flexShrink: 0,
    minHeight: 74,
    borderRadius: 9,
    padding: 12,
    backgroundColor: palette.forest,
    color: palette.white,
  },
  scoreValue: { fontSize: 28, fontFamily: "Helvetica-Bold", lineHeight: 1 },
  scoreLabel: { marginTop: 5, fontSize: 8, color: "#d8e9e1" },
  summaryBox: {
    width: 0,
    flexGrow: 1,
    flexShrink: 1,
    minHeight: 74,
    borderWidth: 1,
    borderColor: palette.line,
    borderRadius: 9,
    padding: 10,
    backgroundColor: palette.cream,
  },
  boxTitle: { fontSize: 9, fontFamily: "Helvetica-Bold", color: palette.forest },
  boxBody: { marginTop: 4, fontSize: 8.2, color: palette.ink },
  sectionTitle: {
    marginTop: 13,
    marginBottom: 6,
    fontSize: 11,
    fontFamily: "Helvetica-Bold",
    color: palette.forest,
  },
  criteriaHeader: {
    flexDirection: "row",
    paddingVertical: 5,
    paddingHorizontal: 7,
    backgroundColor: palette.forestSoft,
    borderWidth: 1,
    borderColor: palette.line,
    borderTopLeftRadius: 6,
    borderTopRightRadius: 6,
  },
  criterion: {
    flexDirection: "row",
    paddingVertical: 5,
    paddingHorizontal: 7,
    borderBottomWidth: 1,
    borderLeftWidth: 1,
    borderRightWidth: 1,
    borderColor: palette.line,
  },
  colName: { width: "20%", paddingRight: 5, fontFamily: "Helvetica-Bold" },
  colScore: { width: "10%", paddingRight: 5, fontFamily: "Helvetica-Bold", color: palette.orange },
  colWhy: { width: "36%", paddingRight: 6, color: palette.muted },
  colAction: { width: "34%", color: palette.muted },
  signalRow: { marginTop: 8, flexDirection: "row", gap: 8 },
  signalBox: {
    width: "50%",
    padding: 8,
    borderWidth: 1,
    borderColor: palette.line,
    borderRadius: 7,
  },
  card: {
    marginBottom: 7,
    padding: 9,
    borderWidth: 1,
    borderColor: palette.line,
    borderLeftWidth: 4,
    borderLeftColor: palette.orange,
    borderRadius: 7,
    backgroundColor: palette.white,
  },
  cardTitle: { fontSize: 9.5, fontFamily: "Helvetica-Bold", color: palette.ink },
  cardBody: { marginTop: 3, fontSize: 8.2, color: palette.muted },
  tuviPanel: {
    marginTop: 10,
    padding: 12,
    borderRadius: 9,
    backgroundColor: palette.forest,
    color: palette.white,
  },
  tuviTitle: { fontSize: 13, fontFamily: "Helvetica-Bold" },
  tuviBody: { marginTop: 5, fontSize: 8.5, color: "#e5f1eb" },
  tuviColumns: { marginTop: 8, flexDirection: "row", gap: 8 },
  tuviItem: { width: "25%", padding: 7, borderRadius: 6, backgroundColor: "#245d49" },
  tuviItemTitle: { fontSize: 8.2, fontFamily: "Helvetica-Bold" },
  tuviItemBody: { marginTop: 3, fontSize: 7.2, color: "#d9e9e1" },
  cta: {
    marginTop: 9,
    padding: 8,
    borderRadius: 6,
    backgroundColor: palette.orange,
    color: palette.white,
    textAlign: "center",
    fontFamily: "Helvetica-Bold",
  },
  footerNote: {
    position: "absolute",
    left: 30,
    right: 92,
    bottom: 20,
    borderTopWidth: 1,
    borderTopColor: palette.line,
    paddingTop: 5,
    fontSize: 6.8,
    color: palette.muted,
  },
  footerPage: {
    position: "absolute",
    right: 30,
    width: 58,
    bottom: 20,
    borderTopWidth: 1,
    borderTopColor: palette.line,
    paddingTop: 5,
    fontSize: 6.8,
    color: palette.muted,
    textAlign: "right",
  },
});

function scoreText(metric: HealthMetric): string {
  return `${metric.score}/${metric.max ?? "-"}`;
}

function whyText(metric: HealthMetric): string {
  if (metric.rationale?.trim()) return metric.rationale.trim();
  if (metric.evidence?.length) return metric.evidence.join("; ");
  return `${metric.status} based on the observable signals available during this review.`;
}

function recommendationText(metric: HealthMetric): string {
  return (
    metric.recommendation?.trim() ||
    `Improve the evidence and customer journey covered by ${metric.label.toLowerCase()}.`
  );
}

function Footer({ page, generatedAt }: { page: number; generatedAt: string }) {
  return (
    <>
      <Text style={styles.footerNote}>
        Tuvi venue pulse | Generated {generatedAt} | Results and listings can change.
      </Text>
      <Text style={styles.footerPage}>Page {page} of 2</Text>
    </>
  );
}

function safeExternalLink(raw?: string): string | undefined {
  const value = (raw || "").trim();
  if (!value) return undefined;
  try {
    const parsed = new URL(value);
    return parsed.protocol === "http:" || parsed.protocol === "https:"
      ? parsed.toString()
      : undefined;
  } catch {
    return undefined;
  }
}

export function RestaurantReportPDF({
  place,
  report,
  logoDataUrl,
  generatedAt,
}: {
  place: RestaurantDetails;
  report: RestaurantReport;
  logoDataUrl: string;
  generatedAt: string;
}) {
  const priorities = report.metrics
    .slice()
    .sort((a, b) => a.value - b.value)
    .slice(0, 4);
  const social = report.socialPresence;
  const menu = report.menuEvidence;

  return (
    <Document
      title={`${report.restaurantName} - Tuvi digital footprint report`}
      author="Tuvi Solutions"
      subject="Restaurant digital footprint, website, menu, social, and local visibility review"
      keywords="restaurant, digital footprint, SEO, reviews, website, Tuvi"
    >
      <Page size={A4_PAGE} style={styles.page}>
        <View style={styles.pageBorder} />
        <View style={styles.header}>
          <View style={styles.brand}>
            {/* React PDF's Image API has no alt prop; the adjacent brand text identifies the logo. */}
            {/* eslint-disable-next-line jsx-a11y/alt-text */}
            <Image src={logoDataUrl} style={styles.logo} />
            <View>
              <Text style={styles.brandName}>Tuvi</Text>
              <Text style={styles.brandLine}>Restaurant growth, under your brand</Text>
            </View>
          </View>
          <Text style={styles.meta}>CONFIDENTIAL VENUE PULSE{`\n`}Two-page review</Text>
        </View>

        <Text style={styles.eyebrow}>Digital footprint review</Text>
        <Text style={styles.title}>{report.restaurantName}</Text>
        <Text style={styles.subtitle}>{report.address || place.address || "Address not available"}</Text>
        <View style={styles.sourceRow}>
          <Link
            src={safeExternalLink(place.mapsUri) || "https://www.google.com/maps"}
            style={styles.sourceLink}
          >
            Google Maps
          </Link>
          {place.attributions?.map((attribution, index) => {
            const provider = attribution.provider?.trim() || "Data source";
            const providerLink = safeExternalLink(attribution.providerUri);
            return providerLink ? (
              <Link
                key={`${provider}-${index}`}
                src={providerLink}
                style={styles.sourceLink}
              >
                Source: {provider}
              </Link>
            ) : (
              <Text key={`${provider}-${index}`}>Source: {provider}</Text>
            );
          })}
        </View>

        <View style={styles.scoreRow}>
          <View style={styles.scoreBox}>
            <Text style={styles.scoreValue}>{report.overallScore}/100</Text>
            <Text style={styles.scoreLabel}>{report.overallLabel} digital footprint</Text>
          </View>
          <View style={styles.summaryBox}>
            <Text style={styles.boxTitle}>What this score means</Text>
            <Text style={styles.boxBody}>
              {report.aiSummary ||
                "Tuvi reviewed the venue's public discovery, trust, conversion, and contact signals."}
            </Text>
          </View>
        </View>

        <Text style={styles.sectionTitle}>How the 100-point score is weighted</Text>
        <View style={styles.criteriaHeader}>
          <Text style={styles.colName}>Criterion</Text>
          <Text style={styles.colScore}>Earned</Text>
          <Text style={styles.colWhy}>Why Tuvi scored it this way</Text>
          <Text style={styles.colAction}>How to improve</Text>
        </View>
        {report.metrics.map((metric) => (
          <View key={metric.key || metric.label} style={styles.criterion} wrap={false}>
            <Text style={styles.colName}>{metric.label}</Text>
            <Text style={styles.colScore}>{scoreText(metric)}</Text>
            <Text style={styles.colWhy}>{whyText(metric)}</Text>
            <Text style={styles.colAction}>{recommendationText(metric)}</Text>
          </View>
        ))}

        <View style={styles.signalRow}>
          <View style={styles.signalBox}>
            <Text style={styles.boxTitle}>Menu evidence</Text>
            <Text style={styles.boxBody}>
              {menu?.status === "present"
                ? `Confirmed${menu.menuUrl ? " from a public menu link" : ""}.`
                : menu?.rationale ||
                  "No verifiable menu link or structured menu evidence was confirmed. Generic listing photos are not treated as menu proof."}
            </Text>
          </View>
          <View style={styles.signalBox}>
            <Text style={styles.boxTitle}>Social presence ({social?.score ?? 0}/{social?.max ?? 3})</Text>
            <Text style={styles.boxBody}>
              {social?.profiles?.length
                ? social.profiles.map((profile) => profile.platform).join(", ")
                : "No public social profile was confirmed from the venue's website during this review."}
            </Text>
          </View>
        </View>
        <Footer page={1} generatedAt={generatedAt} />
      </Page>

      <Page size={A4_PAGE} style={styles.page}>
        <View style={styles.pageBorder} />
        <View style={styles.header}>
          <View style={styles.brand}>
            {/* React PDF's Image API has no alt prop; the adjacent brand text identifies the logo. */}
            {/* eslint-disable-next-line jsx-a11y/alt-text */}
            <Image src={logoDataUrl} style={styles.logo} />
            <View>
              <Text style={styles.brandName}>Tuvi improvement plan</Text>
              <Text style={styles.brandLine}>{report.restaurantName}</Text>
            </View>
          </View>
          <Text style={styles.meta}>PRIORITIZED ACTIONS{`\n`}Designed to improve direct sales</Text>
        </View>

        <Text style={styles.sectionTitle}>What to improve first</Text>
        {priorities.map((metric, index) => (
          <View key={metric.key || metric.label} style={styles.card} wrap={false}>
            <Text style={styles.cardTitle}>
              {index + 1}. {metric.label} - {scoreText(metric)}
            </Text>
            <Text style={styles.cardBody}>Observation: {whyText(metric)}</Text>
            <Text style={styles.cardBody}>Action: {recommendationText(metric)}</Text>
          </View>
        ))}

        <Text style={styles.sectionTitle}>Live market comparison</Text>
        <View style={styles.card}>
          <Text style={styles.cardTitle}>Live comparisons are not embedded in this PDF</Text>
          <Text style={styles.cardBody}>
            Nearby restaurant names, ratings, distances, and derived visibility scores are intentionally excluded. Open the unlocked web report to refresh and view the current market comparison with its live source attribution.
          </Text>
        </View>

        <View style={styles.tuviPanel} wrap={false}>
          <Text style={styles.tuviTitle}>Tuvi can turn these fixes into measurable sales growth</Text>
          <Text style={styles.tuviBody}>
            Tuvi brings discovery, a fast restaurant website, direct ordering, reputation, and repeat-guest marketing into one managed growth stack.
          </Text>
          <View style={styles.tuviColumns}>
            <View style={styles.tuviItem}>
              <Text style={styles.tuviItemTitle}>Get found</Text>
              <Text style={styles.tuviItemBody}>Local SEO, listing consistency, fresh content.</Text>
            </View>
            <View style={styles.tuviItem}>
              <Text style={styles.tuviItemTitle}>Convert visits</Text>
              <Text style={styles.tuviItemBody}>Mobile-first site, menu, bookings and ordering.</Text>
            </View>
            <View style={styles.tuviItem}>
              <Text style={styles.tuviItemTitle}>Own orders</Text>
              <Text style={styles.tuviItemBody}>First-party checkout and guest data under your brand.</Text>
            </View>
            <View style={styles.tuviItem}>
              <Text style={styles.tuviItemTitle}>Bring guests back</Text>
              <Text style={styles.tuviItemBody}>Reviews, loyalty, email, SMS and push campaigns.</Text>
            </View>
          </View>
          <Link src="https://tuvisolutions.com/book" style={styles.cta}>
            Book a free growth review at tuvisolutions.com/book
          </Link>
        </View>

        <Text style={[styles.cardBody, { marginTop: 9 }]}>
          Method note: the digital-footprint score uses the weights shown on page 1. Live market comparisons are intentionally excluded from this PDF and must be refreshed in the unlocked web report. Missing, blocked, or unconfirmed evidence is never scored as present. Results vary by market and execution.
        </Text>
        <Footer page={2} generatedAt={generatedAt} />
      </Page>
    </Document>
  );
}
