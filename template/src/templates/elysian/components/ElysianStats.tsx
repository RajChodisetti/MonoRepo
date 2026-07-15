import type { ElysianContent } from "../lib/mapContent";

export default function ElysianStats({ stats }: { stats: ElysianContent["stats"] }) {
  if (!stats.length) return null;

  return (
    <section className="stats section">
      <div className="container stats-grid">
        {stats.map((stat) => (
          <div key={stat.label} className="stat-item reveal fade-up">
            {stat.animate !== false ? (
              <span className="stat-num" data-count={stat.value.replace(/\D/g, "") || "0"}>
                0
              </span>
            ) : (
              <span className="stat-num">{stat.value}</span>
            )}
            {stat.suffix ? <span className="stat-suffix">{stat.suffix}</span> : null}
            <p>{stat.label}</p>
          </div>
        ))}
      </div>
    </section>
  );
}
