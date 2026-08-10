import Image from "next/image";

export function MenuItemsStackVisual() {
  const items = [
    {
      name: "Birria Tacos",
      price: "$13.99",
      image: "/menu/birria-tacos.jpg",
      bestseller: true,
    },
    {
      name: "Quesadilla",
      price: "$15.99",
      image: "/menu/tacos.jpg",
      bestseller: false,
    },
    {
      name: "Churros",
      price: "$8.99",
      image: "/menu/churros.jpg",
      bestseller: false,
    },
  ] as const;

  return (
    <div className="flex w-full max-w-[300px] flex-col gap-3 pt-1 sm:max-w-[320px]">
      {items.map((item) => (
        <div
          key={item.name}
          className="relative flex items-center gap-3.5 rounded-2xl bg-white p-3 shadow-[0_8px_24px_rgba(0,0,0,0.08)]"
        >
          <div className="relative h-[72px] w-[72px] shrink-0 overflow-hidden rounded-xl">
            <Image src={item.image} alt="" fill className="object-cover" sizes="72px" />
            {item.bestseller ? (
              <span className="absolute left-1.5 top-1.5 rounded-full bg-[#2f6b54] px-1.5 py-0.5 text-[9px] font-bold text-white">
                Bestseller
              </span>
            ) : null}
          </div>
          <div className="min-w-0 flex-1">
            <p className="text-[15px] font-bold text-[#0f271f]">{item.name}</p>
            <p className="mt-0.5 text-[13px] font-medium text-[#8a8a8a]">{item.price}</p>
            <div className="mt-2 space-y-1.5">
              <div className="h-1.5 w-[92%] rounded-full bg-[#dce6dd]" />
              <div className="h-1.5 w-[62%] rounded-full bg-[#dce6dd]" />
            </div>
          </div>
        </div>
      ))}
    </div>
  );
}

export function RewardsPhoneVisual() {
  return (
    <div className="relative mx-auto w-[220px] sm:w-[240px]">
      <div className="overflow-hidden rounded-[32px] border-[6px] border-[#0f271f] bg-white shadow-[0_24px_50px_rgba(0,0,0,0.22)]">
        <div className="bg-[#f2ecdf] px-3 pb-4 pt-3">
          <div className="mx-auto mb-2 h-1 w-16 rounded-full bg-[#ddd]" />
          <div className="rounded-xl bg-[#2f6b54] px-3 py-2 text-center text-[11px] font-semibold text-white">
            Rewards available soon
          </div>
          <div className="relative mt-3 h-[110px] overflow-hidden rounded-2xl">
            <Image
              src="/menu/pepperoni-pizza.jpg"
              alt=""
              fill
              className="object-cover"
              sizes="220px"
            />
            <div className="absolute inset-x-0 bottom-0 bg-gradient-to-t from-black/70 to-transparent p-2.5">
              <p className="text-[13px] font-bold text-white">Presto Pizza</p>
            </div>
          </div>
          <div className="mt-3 rounded-xl bg-white p-3 shadow-sm">
            <p className="text-[11px] font-medium text-[#666]">You need 40 more points</p>
            <div className="mt-2 h-2 overflow-hidden rounded-full bg-[#dce6dd]">
              <div className="h-full w-[62%] rounded-full bg-[#2f6b54]" />
            </div>
          </div>
          <p className="mt-3 text-[12px] font-bold text-[#0f271f]">Recommended for you</p>
          <div className="mt-2 flex gap-2">
            {["/menu/birria-tacos.jpg", "/menu/classic-burger.jpg", "/menu/caesar-salad.jpg"].map(
              (src) => (
                <div
                  key={src}
                  className="relative h-12 w-12 shrink-0 overflow-hidden rounded-lg"
                >
                  <Image src={src} alt="" fill className="object-cover" sizes="48px" />
                </div>
              ),
            )}
          </div>
        </div>
      </div>
    </div>
  );
}

export function PitaWrapsMenuVisual() {
  const items = [
    {
      name: "Build Your Bowl",
      price: "$14.50",
      description: "Pick a base, protein, and house sauces — topped your way.",
      image: "/menu/caesar-salad.jpg",
    },
    {
      name: "Charfold Chicken Wrap",
      price: "$13.90",
      description:
        "Grilled chicken, crisp lettuce, tomato, fries, and house garlic sauce in a warm wrap.",
      image: "/menu/classic-burger.jpg",
    },
  ] as const;

  return (
    <div className="w-full max-w-[340px] rounded-[22px] bg-white p-4 shadow-[0_12px_32px_rgba(0,0,0,0.1)] sm:p-5">
      <p className="text-[16px] font-bold text-[#0f271f]">Bowls & wraps</p>
      <div className="mt-3 space-y-4">
        {items.map((item) => (
          <div key={item.name} className="flex gap-3 border-t border-black/[0.06] pt-4 first:border-0 first:pt-0">
            <div className="min-w-0 flex-1">
              <p className="text-[14px] font-bold text-[#0f271f]">{item.name}</p>
              <p className="mt-0.5 text-[13px] font-semibold text-[#0f271f]">{item.price}</p>
              <p className="mt-1.5 text-[12px] leading-snug text-[#6b6b6b]">{item.description}</p>
            </div>
            <div className="relative h-[72px] w-[72px] shrink-0">
              <div className="relative h-full w-full overflow-hidden rounded-xl">
                <Image src={item.image} alt="" fill className="object-cover" sizes="72px" />
              </div>
              <span className="absolute -bottom-1 -right-1 flex h-6 w-6 items-center justify-center rounded-full bg-[#2f6b54] text-[16px] font-bold leading-none text-white shadow-md">
                +
              </span>
            </div>
          </div>
        ))}
      </div>
    </div>
  );
}

export function OrderTrackingPhoneVisual() {
  return (
    <div className="absolute inset-0">
      <div
        className="absolute inset-0"
        style={{
          background:
            "linear-gradient(180deg, rgba(12,12,12,0.72) 0%, rgba(12,12,12,0.25) 38%, rgba(12,12,12,0.45) 100%), url(/guides/interview.jpg) center/cover",
        }}
        aria-hidden="true"
      />
      <div className="absolute bottom-6 left-1/2 z-10 w-[200px] -translate-x-1/2 overflow-hidden rounded-[28px] border-[5px] border-white/90 bg-white shadow-[0_20px_40px_rgba(0,0,0,0.35)] sm:bottom-8 sm:w-[220px]">
        <div className="px-3 pb-3 pt-2.5">
          <div className="mx-auto mb-2 h-1 w-12 rounded-full bg-[#ddd]" />
          <p className="text-[12px] font-bold text-[#0f271f]">Manhattan Bistro</p>
          <div className="mt-3 flex items-center justify-between px-1">
            {["Ordered", "Preparing", "Ready"].map((step, i) => (
              <div key={step} className="flex flex-1 flex-col items-center">
                <span
                  className={`h-2.5 w-2.5 rounded-full ${
                    i < 2 ? "bg-[#2f6b54]" : "bg-[#ddd]"
                  }`}
                />
                <span className="mt-1 text-[8px] font-medium text-[#666]">{step}</span>
              </div>
            ))}
          </div>
          <div className="relative mt-2 h-1 rounded-full bg-[#dce6dd]">
            <div className="absolute inset-y-0 left-0 w-[55%] rounded-full bg-[#2f6b54]" />
          </div>
          <p className="mt-3 text-[11px] font-bold text-[#0f271f]">Order Again</p>
          <div className="mt-2 flex gap-2">
            {["/menu/pepperoni-pizza.jpg", "/menu/pasta-alfredo.jpg", "/menu/classic-burger.jpg"].map(
              (src) => (
                <div key={src} className="relative h-11 w-11 overflow-hidden rounded-lg">
                  <Image src={src} alt="" fill className="object-cover" sizes="44px" />
                </div>
              ),
            )}
          </div>
        </div>
      </div>
    </div>
  );
}
