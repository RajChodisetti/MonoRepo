import type { ReactNode } from "react";

function StatusBar() {
  return (
    <div className="relative z-30 flex h-11 items-end justify-between bg-white px-7 pb-1.5 text-[13px] font-semibold text-black">
      <span className="w-14 tracking-tight">9:41</span>
      <div className="flex w-16 items-center justify-end gap-1">
        <svg viewBox="0 0 17 12" className="h-3 w-[17px]" aria-hidden="true">
          <rect x="0" y="3.5" width="3" height="8.5" rx="0.6" fill="currentColor" />
          <rect x="4.5" y="2" width="3" height="10" rx="0.6" fill="currentColor" />
          <rect x="9" y="0.5" width="3" height="11.5" rx="0.6" fill="currentColor" />
          <rect x="13.5" y="0" width="3" height="12" rx="0.6" fill="currentColor" opacity="0.35" />
        </svg>
        <svg viewBox="0 0 16 12" className="h-3 w-4" aria-hidden="true">
          <path
            fill="currentColor"
            d="M8 3.6c1.7 0 3.2.7 4.3 1.8l1.1-1.2A7.4 7.4 0 0 0 8 1.6 7.4 7.4 0 0 0 2.6 4.2l1.1 1.2A5.7 5.7 0 0 1 8 3.6Zm0 3.1c.9 0 1.7.3 2.4.9l1.1-1.2A5 5 0 0 0 8 5.2a5 5 0 0 0-3.5 1.4l1.1 1.2c.7-.6 1.5-.9 2.4-.9Zm0 3.2a1.6 1.6 0 1 0 0 3.2 1.6 1.6 0 0 0 0-3.2Z"
          />
        </svg>
        <svg viewBox="0 0 25 12" className="h-3 w-[25px]" aria-hidden="true">
          <rect x="0.5" y="0.5" width="21" height="11" rx="2.2" stroke="currentColor" strokeWidth="1" fill="none" opacity="0.4" />
          <rect x="2" y="2" width="16" height="8" rx="1.2" fill="currentColor" />
          <path d="M23 3.5v5a1.6 1.6 0 0 0 0-5Z" fill="currentColor" opacity="0.4" />
        </svg>
      </div>
    </div>
  );
}

export default function PhoneMockup({ children }: { children: ReactNode }) {
  return (
    <div className="relative mx-auto h-[760px] w-[390px] shrink-0 sm:h-[800px] sm:w-[420px] md:h-[840px] md:w-[440px]">
      <div className="absolute inset-0 rounded-[52px] bg-[#1c1c1c] p-[3px] shadow-[0_28px_60px_rgba(0,0,0,0.28)]">
        <div className="relative h-full w-full overflow-hidden rounded-[49px] bg-black p-[11px]">
          <div className="relative h-full w-full overflow-hidden rounded-[38px] bg-white">
            <div className="pointer-events-none absolute left-1/2 top-[12px] z-40 h-[30px] w-[118px] -translate-x-1/2 rounded-full bg-black" />

            <StatusBar />

            <div className="relative h-[calc(100%-2.75rem)] w-full overflow-hidden bg-white">{children}</div>
          </div>
        </div>
      </div>
    </div>
  );
}
