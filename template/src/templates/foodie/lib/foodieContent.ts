// Static content for the Foodie landing template (TEMPLATE=4).
// Kept as a plain object so it can later be produced by a
// mapFoodieContent(restaurant) adapter without touching the components.

export interface FoodieNavLink {
  href: string;
  label: string;
  active?: boolean;
}

export interface FoodieDishCard {
  name: string;
  rating: number; // 0-5
  description: string;
  price: string;
  image: string;
}

export interface FoodieMenuItem {
  id: string;
  name: string;
  ratingLabel: string; // e.g. "(5.6k)"
  price: string;
  description: string;
  image: string;
  featured?: boolean;
}

export interface FoodieReviewItem {
  id: string;
  name: string;
  location: string;
  avatar: string;
  rating: number;
  quote?: string;
}

export interface FoodieContent {
  brand: {
    name: string;
    logo: string;
  };
  nav: FoodieNavLink[];
  hero: {
    eyebrow: string;
    titleLead: string; // "Foodie Restaurant and Enjoy"
    titleAccent: string; // "The Food" (underlined)
    description: string;
    primaryCta: string;
    secondaryCta: string;
    hours: string;
    badge: string;
    plate: string;
    garnish: {
      tomato: string;
      onion: string;
      basil: string;
    };
  };
  dish: FoodieDishCard;
  menu: {
    titleLead: string;
    titleAccent: string;
    items: FoodieMenuItem[];
  };
  reviews: {
    titleBefore: string;
    titleAccent: string;
    titleAfter: string;
    description: string;
    chefImage: string;
    reviewCount: string;
    badgeLabel: string;
    avatars: string[];
    items: FoodieReviewItem[];
  };
  cta: {
    titleLead: string;
    titleAccent: string;
    description: string;
    primaryCta: string;
    secondaryCta: string;
    wrapImage: string;
    friesImage: string;
  };
  contact: {
    eyebrow: string;
    titleLead: string;
    titleAccent: string;
    address: string;
    phone: string;
    email: string;
    hoursLine: string;
    coordinates: { latitude: number; longitude: number };
    directionsLabel: string;
    callLabel: string;
  };
  footer: {
    tagline: string;
    links: FoodieNavLink[];
  };
}

export const foodieContent: FoodieContent = {
  brand: {
    name: "Foodie",
    logo: "/foodie/logo.svg",
  },
  nav: [
    { href: "#home", label: "Home", active: true },
    { href: "#menu", label: "Menu" },
    { href: "#reservation", label: "Reservation" },
    { href: "#contact", label: "Contact" },
    { href: "#about", label: "about us" },
    { href: "#blog", label: "Blog" },
  ],
  hero: {
    eyebrow: "Welcome to",
    titleLead: "Foodie Restaurant and Enjoy",
    titleAccent: "The Food",
    description:
      "Lorem ipsum dolor sit amet, consectetur adipiscing elit. Vestibulum convallis ante ante, ut tempor neque bibendum non. Ut enim lacus, auctor nec convallis sed, vehicula ut eros.",
    primaryCta: "Reserve a Table",
    secondaryCta: "Onlnne Order",
    hours: "Open: 11:00am-11.00pm",
    badge: "Best Food",
    plate: "/foodie/hero-salad.png",
    garnish: {
      tomato: "/foodie/tomato.png",
      onion: "/foodie/red-onion.png",
      basil: "/foodie/basil.png",
    },
  },
  dish: {
    name: "Salman Salad",
    rating: 5,
    description: "Lorem ipsum dolor sit amet, consectetur adipiscing elit.",
    price: "$12",
    image: "/foodie/salman-salad.png",
  },
  menu: {
    titleLead: "Our Popular",
    titleAccent: "Dishes",
    items: [
      {
        id: "noodles",
        name: "Chinese noodles Pasta",
        ratingLabel: "(5.6k)",
        price: "$20.00",
        description:
          "Lorem ipsum dolor sit amet, consectetur adipiscing elit. Vestibulum.",
        image: "/foodie/dish-noodles.png",
      },
      {
        id: "chowmein",
        name: "Vegetable Chowmien",
        ratingLabel: "(5.6k)",
        price: "$20.00",
        description:
          "Lorem ipsum dolor sit amet, consectetur adipiscing elit. Vestibulum.",
        image: "/foodie/dish-chowmein.png",
      },
      {
        id: "penne",
        name: "Pasta al pomoddoro",
        ratingLabel: "(5.6k)",
        price: "$20.00",
        description:
          "Lorem ipsum dolor sit amet, consectetur adipiscing elit. Vestibulum.",
        image: "/foodie/dish-penne.png",
        featured: true,
      },
      {
        id: "curry",
        name: "Rice and curry",
        ratingLabel: "(5.6k)",
        price: "$20.00",
        description:
          "Lorem ipsum dolor sit amet, consectetur adipiscing elit. Vestibulum.",
        image: "/foodie/dish-curry.png",
      },
    ],
  },
  reviews: {
    titleBefore: "What are our",
    titleAccent: "Customer",
    titleAfter: "say about us",
    description:
      "Lorem ipsum dolor sit amet, consectetur adipiscing elit. Vestibulum convallis ante ante, ut tempor neque bibendum non. Ut enim lacus, auctor nec convallis sed, vehicula ut eros.",
    chefImage: "/foodie/chef.png",
    reviewCount: "22k+",
    badgeLabel: "Our review",
    avatars: [
      "/foodie/avatar-1.png",
      "/foodie/avatar-2.png",
      "/foodie/avatar-3.png",
      "/foodie/avatar-4.png",
    ],
    items: [
      {
        id: "theresa",
        name: "Theresa Webb",
        location: "Westheimer, Santa Ana",
        avatar: "/foodie/avatar-1.png",
        rating: 5,
        quote:
          "Amazing flavors and warm hospitality — every visit feels special.",
      },
      {
        id: "marcus",
        name: "Marcus Chen",
        location: "Downtown, Austin",
        avatar: "/foodie/avatar-2.png",
        rating: 5,
        quote: "Best pasta in town. Fresh ingredients and friendly staff.",
      },
      {
        id: "sofia",
        name: "Sofia Alvarez",
        location: "Midtown, Chicago",
        avatar: "/foodie/avatar-3.png",
        rating: 5,
        quote: "We come back every week. The salad plate is unreal.",
      },
    ],
  },
  cta: {
    titleLead: "Are You Ready to Enjoy",
    titleAccent: "Our Food?",
    description:
      "Lorem ipsum dolor sit amet, consectetur adipiscing elit. Vestibulum convallis ante ante, ut tempor neque bibendum non. Ut enim lacus, auctor nec convallis sed, vehicula ut eros.",
    primaryCta: "Reserve a Table",
    secondaryCta: "Online Order",
    wrapImage: "/foodie/cta-wrap.png",
    friesImage: "/foodie/cta-fries.png",
  },
  contact: {
    eyebrow: "Visit Us",
    titleLead: "Find Your Way to",
    titleAccent: "Foodie",
    address: "214 Westheimer Rd, Santa Ana, CA 92701",
    phone: "+1 (555) 214-8900",
    email: "hello@foodie.demo",
    hoursLine: "Open daily · 11:00am – 11:00pm",
    coordinates: { latitude: 33.7455, longitude: -117.8677 },
    directionsLabel: "Get Directions",
    callLabel: "Call Now",
  },
  footer: {
    tagline: "Fresh plates, warm hospitality, and flavors you’ll come back for.",
    links: [
      { href: "#home", label: "Home" },
      { href: "#menu", label: "Menu" },
      { href: "#reservation", label: "Reservation" },
      { href: "#contact", label: "Contact" },
    ],
  },
};
