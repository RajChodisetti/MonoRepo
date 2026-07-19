import type { ElysianContent } from "../lib/mapContent";
import ElysianImage from "./ElysianImage";
import PhotoAttribution from "@/components/PhotoAttribution";

export default function ElysianAbout({ about }: { about: ElysianContent["about"] }) {
  return (
    <section className="about section" id="about">
      <div className="container about-grid">
        <div className="about-image reveal fade-up">
          {about.image ? (
            <ElysianImage
              src={about.image}
              alt="Restaurant interior"
              media={about.imageMedia}
              fill
              className="about-photo"
              sizes="(max-width: 768px) 100vw, 50vw"
            />
          ) : null}
          {about.imageMedia?.sourceKind === "google_places_live" ? (
            <div className="absolute inset-x-4 bottom-4 z-10 rounded bg-black/65 px-3 py-2 text-white/80">
              <PhotoAttribution media={about.imageMedia} compact />
            </div>
          ) : null}
          <div className="about-image-frame" />
          <div className="about-badge">
            <span className="badge-num">{about.badgeYears}</span>
            <span
              className="badge-label"
              dangerouslySetInnerHTML={{ __html: about.badgeLabel }}
            />
          </div>
        </div>
        <div className="about-content">
          <p className="eyebrow reveal fade-up">Our Story</p>
          <h2 className="section-title reveal fade-up">
            A Table Built on <span className="gold-text">Patience</span>, Fire &amp; Memory
          </h2>
          {about.paragraphs.map((p, i) => (
            <p key={i} className="about-text reveal fade-up">
              {p}
            </p>
          ))}
          {about.showTimeline && (
            <div className="timeline reveal fade-up">
              {about.timeline.map((item) => (
                <div key={item.year} className="timeline-item">
                  <span className="timeline-year">{item.year}</span>
                  <div className="timeline-dot" />
                  <p>{item.text}</p>
                </div>
              ))}
            </div>
          )}
        </div>
      </div>
    </section>
  );
}
