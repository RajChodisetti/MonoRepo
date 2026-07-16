"use client";

import Link from "next/link";
import { useEffect, useState } from "react";
import { adminFetch } from "@/lib/client-api";
import type { BulkSendStatus, Restaurant, ScrapeJob } from "@/lib/types";
import { formatDate } from "@/lib/constants";
import { EmptyState, ErrorBanner, PageHeader, StatusBadge } from "@/components/ui";

export default function DashboardPage() {
  const [jobs, setJobs] = useState<ScrapeJob[]>([]);
  const [outreach, setOutreach] = useState<BulkSendStatus | null>(null);
  const [restaurants, setRestaurants] = useState<Restaurant[]>([]);
  const [error, setError] = useState<string | null>(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    let cancelled = false;
    async function load() {
      setLoading(true);
      setError(null);
      try {
        const [jobsRes, outreachRes, restRes] = await Promise.all([
          adminFetch<{ jobs: ScrapeJob[] }>("scrape-jobs", {
            query: { limit: 10 },
          }),
          adminFetch<BulkSendStatus>("outreach/bulk-send/status").catch(
            () => null,
          ),
          adminFetch<{ items: Restaurant[] }>("restaurants", {
            query: { limit: 20 },
          }).catch(() => ({ items: [] as Restaurant[] })),
        ]);
        if (cancelled) return;
        setJobs(jobsRes.jobs || []);
        setOutreach(outreachRes);
        setRestaurants(restRes.items || []);
      } catch (err) {
        if (!cancelled) {
          setError(err instanceof Error ? err.message : "Failed to load");
        }
      } finally {
        if (!cancelled) setLoading(false);
      }
    }
    load();
    return () => {
      cancelled = true;
    };
  }, []);

  const activeJobs = jobs.filter((j) =>
    ["queued", "running", "waiting"].includes(j.status),
  );

  return (
    <div>
      <PageHeader
        title="Dashboard"
        subtitle="Scrape, review, and outreach overview"
        actions={
          <>
            <Link className="btn btn-secondary" href="/scrape-jobs">
              Scrape jobs
            </Link>
            <Link className="btn btn-primary" href="/outreach">
              Outreach
            </Link>
          </>
        }
      />
      <ErrorBanner message={error} />
      {loading ? <EmptyState message="Loading dashboard…" /> : null}

      <div
        style={{
          display: "grid",
          gridTemplateColumns: "repeat(auto-fit, minmax(180px, 1fr))",
          gap: "0.85rem",
          marginBottom: "1.25rem",
        }}
      >
        <div className="card">
          <div style={{ color: "var(--muted)", fontSize: "0.8rem" }}>
            Active scrapes
          </div>
          <div style={{ fontSize: "1.8rem", fontWeight: 700 }}>
            {activeJobs.length}
          </div>
        </div>
        <div className="card">
          <div style={{ color: "var(--muted)", fontSize: "0.8rem" }}>
            Eligible to email
          </div>
          <div style={{ fontSize: "1.8rem", fontWeight: 700 }}>
            {outreach?.pending_eligible_count ?? "—"}
          </div>
        </div>
        <div className="card">
          <div style={{ color: "var(--muted)", fontSize: "0.8rem" }}>
            Recent leads loaded
          </div>
          <div style={{ fontSize: "1.8rem", fontWeight: 700 }}>
            {restaurants.length}
          </div>
        </div>
      </div>

      <div
        style={{
          display: "grid",
          gridTemplateColumns: "repeat(auto-fit, minmax(280px, 1fr))",
          gap: "1rem",
        }}
      >
        <section className="card">
          <h2 style={{ margin: "0 0 0.75rem", fontSize: "1.05rem" }}>
            Recent scrape jobs
          </h2>
          {jobs.length === 0 ? (
            <p style={{ color: "var(--muted)", margin: 0 }}>No jobs yet.</p>
          ) : (
            <ul style={{ listStyle: "none", margin: 0, padding: 0 }}>
              {jobs.slice(0, 5).map((job) => (
                <li
                  key={job.id}
                  style={{
                    display: "flex",
                    justifyContent: "space-between",
                    gap: "0.75rem",
                    padding: "0.55rem 0",
                    borderBottom: "1px solid var(--line)",
                  }}
                >
                  <Link href={`/scrape-jobs/${job.id}`}>
                    {job.city} · {job.niche}
                  </Link>
                  <StatusBadge status={job.status} />
                </li>
              ))}
            </ul>
          )}
        </section>

        <section className="card">
          <h2 style={{ margin: "0 0 0.75rem", fontSize: "1.05rem" }}>
            Outreach status
          </h2>
          {outreach ? (
            <div style={{ display: "grid", gap: "0.45rem", fontSize: "0.95rem" }}>
              <div>
                Pending eligible: <strong>{outreach.pending_eligible_count}</strong>
              </div>
              <div>
                Max sends / window: <strong>{outreach.max_sends}</strong>
              </div>
              <div>
                Active job:{" "}
                <StatusBadge status={outreach.active_job?.status || "none"} />
              </div>
              {outreach.next_available_at ? (
                <div style={{ color: "var(--muted)" }}>
                  Next available: {formatDate(outreach.next_available_at)}
                </div>
              ) : null}
            </div>
          ) : (
            <p style={{ color: "var(--muted)", margin: 0 }}>
              Outreach status unavailable (email may be disabled).
            </p>
          )}
        </section>
      </div>
    </div>
  );
}
