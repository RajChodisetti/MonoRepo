import type { ElysianContent } from "../lib/mapContent";

export default function ElysianContact({
  contact,
  name,
  showMap,
}: {
  contact: ElysianContent["contact"];
  name: string;
  showMap: boolean;
}) {
  return (
    <section className="contact section" id="contact">
      <div className="container contact-grid">
        <div className="contact-info reveal fade-up">
          <p className="eyebrow">Visit Us</p>
          <h2 className="section-title">
            Find Your Way to <span className="gold-text">{name.split(" ").pop()}</span>
          </h2>
          <ul className="contact-list">
            {contact.address ? (
              <li>
                <span className="contact-icon">📍</span> {contact.address}
              </li>
            ) : null}
            {contact.phone ? (
              <li>
                <span className="contact-icon">📞</span>{" "}
                <a href={`tel:${contact.phone.replace(/\s/g, "")}`}>{contact.phone}</a>
              </li>
            ) : null}
            {contact.email ? (
              <li>
                <span className="contact-icon">✉</span>{" "}
                <a href={`mailto:${contact.email}`}>{contact.email}</a>
              </li>
            ) : null}
            {contact.hoursLine ? (
              <li>
                <span className="contact-icon">🕐</span> {contact.hoursLine}
              </li>
            ) : null}
          </ul>
          {contact.mapsUrl ? (
            <div className="social-icons">
              <a href={contact.mapsUrl} target="_blank" rel="noopener noreferrer" aria-label="Maps">
                Map
              </a>
              {contact.phone ? (
                <a href={`tel:${contact.phone.replace(/\s/g, "")}`} aria-label="Phone">
                  Call
                </a>
              ) : null}
            </div>
          ) : null}
        </div>
        {showMap ? (
          <div className="contact-map reveal fade-up">
            {contact.embedSrc ? (
              <iframe
                title={`Map of ${name}`}
                src={contact.embedSrc}
                style={{ width: "100%", aspectRatio: "4/3", border: 0, borderRadius: "var(--radius-lg)" }}
                loading="lazy"
              />
            ) : (
              <a href={contact.mapsUrl} className="map-placeholder" target="_blank" rel="noopener noreferrer">
                <span>View on Map</span>
                <p>{contact.address}</p>
              </a>
            )}
          </div>
        ) : null}
      </div>
    </section>
  );
}
