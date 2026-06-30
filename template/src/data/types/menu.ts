export interface MenuItem {
  name: string;
  description: string;
  price?: string;
  image?: string;
  tags?: string[];
  spicyLevel?: 0 | 1 | 2 | 3;
  isChefSpecial?: boolean;
  category: string;
}

export interface MenuCategory {
  name: string;
  items: MenuItem[];
}
