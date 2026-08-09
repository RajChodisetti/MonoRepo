export type TrustedOwner = {
  id: string;
  quote: string;
  name: string;
  business: string;
  imageUrl: string;
  metrics: {
    value: string;
    label: string;
  }[];
};

export const trustedOwners: TrustedOwner[] = [
  {
    id: "sandy-sei",
    quote:
      "I would recommend Tuvi. Don't take my word for it, read the reviews, see their videos, and just give them a call.",
    name: "Sandy Sei",
    business: "Owner of Steamrail Noodles",
    imageUrl: "/owners/cyclo.jpg",
    metrics: [
      { value: "+$104,500", label: "Online sales" },
      { value: "$31,000", label: "Savings in third-party fees" },
    ],
  },
  {
    id: "marco-r",
    quote:
      "Tuvi gave us the same ordering experience as the big chains. Our online sales jumped in the first month.",
    name: "Marco R.",
    business: "Owner of Dockweave Pasta Co.",
    imageUrl: "/owners/pasta.jpg",
    metrics: [
      { value: "+$82,000", label: "Online sales" },
      { value: "$18,400", label: "Fees saved yearly" },
    ],
  },
  {
    id: "priya-n",
    quote:
      "Setup was effortless and the support team actually knows restaurants. We finally own our guest relationship.",
    name: "Priya N.",
    business: "Owner of Charfold Burgers",
    imageUrl: "/owners/burger.jpg",
    metrics: [
      { value: "+$67,200", label: "Online sales" },
      { value: "2.4x", label: "Repeat orders" },
    ],
  },
  {
    id: "james-k",
    quote:
      "We switched off DoorDash fees and guests still order every week through our own app and website.",
    name: "James K.",
    business: "Owner of Parcel Oven Pizza",
    imageUrl: "/owners/pizza.jpg",
    metrics: [
      { value: "+$91,800", label: "Online sales" },
      { value: "$24,000", label: "Third-party savings" },
    ],
  },
];
