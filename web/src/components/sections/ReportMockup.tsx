import PhoneMockup from "@/components/sections/report/PhoneMockup";
import ReportDemoScreen from "@/components/sections/report/ReportDemoScreen";

export default function ReportMockup() {
  return (
    <section className="bg-bg px-4 pb-12 pt-6 sm:px-8 sm:pb-16 sm:pt-8 md:px-12">
      <div className="relative mx-auto w-full max-w-[1100px]">
        <div className="relative h-[calc(56px+520px)] overflow-hidden sm:h-[calc(56px+560px)] md:h-[calc(56px+600px)]">
          <div
            className="tuvi-forest-panel absolute inset-x-0 bottom-0 h-[520px] overflow-hidden rounded-[32px] sm:h-[560px] sm:rounded-[40px] md:h-[600px]"
            aria-hidden="true"
          >
            <svg
              className="pointer-events-none absolute inset-0 h-full w-full"
              viewBox="0 0 1100 600"
              preserveAspectRatio="none"
            >
              {[180, 280, 380, 500, 640, 800, 980].map((r, i) => (
                <circle
                  key={r}
                  cx="50"
                  cy="580"
                  r={r}
                  fill="none"
                  stroke="rgba(255,255,255,0.16)"
                  strokeWidth={1.25 - i * 0.06}
                />
              ))}
            </svg>
            <div
              className="pointer-events-none absolute inset-0 opacity-[0.22] mix-blend-overlay"
              style={{
                backgroundImage:
                  "url(\"data:image/svg+xml,%3Csvg viewBox='0 0 200 200' xmlns='http://www.w3.org/2000/svg'%3E%3Cfilter id='n'%3E%3CfeTurbulence type='fractalNoise' baseFrequency='0.85' numOctaves='4' stitchTiles='stitch'/%3E%3C/filter%3E%3Crect width='100%25' height='100%25' filter='url(%23n)'/%3E%3C/svg%3E\")",
              }}
            />
          </div>

          <div className="absolute left-1/2 top-0 z-10 -translate-x-1/2">
            <PhoneMockup>
              <ReportDemoScreen />
            </PhoneMockup>
          </div>
        </div>
      </div>
    </section>
  );
}
