export type AdminUser = {
  id: string;
  email: string;
  full_name?: string;
  role: string;
};

export type ScrapeJobProgress = {
  cells_total: number;
  cells_pending: number;
  cells_completed: number;
  cells_subdivided: number;
  cells_failed: number;
  cells_saturated: number;
  candidates_total: number;
  candidates_pending: number;
  candidates_imported: number;
  candidates_duplicate: number;
  candidates_failed: number;
};

export type ScrapeJob = {
  id: string;
  city: string;
  city_key: string;
  niche: string;
  status: string;
  cycle_number: number;
  max_requests_per_window: number;
  requests_used_window: number;
  requests_used_total: number;
  window_started_at?: string;
  resume_at?: string;
  waiting_reason?: string;
  last_error?: string;
  created_at: string;
  updated_at: string;
  progress: ScrapeJobProgress;
};

export type Restaurant = {
  id: string;
  name: string;
  email?: string;
  status: string;
  is_contacted?: boolean;
  shown_interest?: boolean;
  email_sent?: boolean;
  email_send_count?: number;
  last_email_sent_at?: string;
  created_at?: string;
  updated_at?: string;
};

export type BulkSendStatus = {
  pending_eligible_count: number;
  max_sends: number;
  next_available_at?: string;
  active_job?: {
    id: string;
    status: string;
    created_at?: string;
  };
  last_completed_job?: {
    id: string;
    status: string;
    created_at?: string;
  };
};

export type Campaign = {
  id: string;
  restaurant_id?: string;
  status: string;
  subject?: string;
  body_html?: string;
  body_text?: string;
  approved_at?: string;
  approved_by?: string;
  updated_at: string;
  created_at?: string;
};

export type DemoSite = {
  id: string;
  restaurant_id?: string;
  status: string;
  slug?: string;
  published_at?: string;
  updated_at: string;
};

export type RestaurantProfile = {
  description?: string;
  opening_hours?: unknown;
  phone?: string;
  website?: string;
  address?: string;
  city?: string;
  state?: string;
  country?: string;
  latitude?: number | null;
  longitude?: number | null;
  google_place_id?: string;
  rating?: number | null;
  reviews_count?: number | null;
  price_level?: string | number | null;
  cuisines?: unknown;
  owners?: unknown;
  images?: unknown;
  dietary_options?: unknown;
  parking_info?: unknown;
  reservation_policy?: unknown;
  brand_tone?: unknown;
  scrape_status?: string;
  scrape_errors?: unknown;
};

export type ProfileReviewPreview = {
  restaurant_id: string;
  restaurant_name?: string;
  contact_email?: string;
  ocr_status?: string;
  ocr_input_fingerprint?: string;
  ocr_completed_at?: string;
  review_status?: string;
  reviewed_at?: string;
  reviewed_by?: string;
  restaurant_updated_at?: string;
  profile_updated_at?: string;
  profile?: RestaurantProfile;
};

export type SiteImage = {
  id?: string;
  url?: string;
  thumbnail_url?: string;
  image_type?: string;
  title?: string;
  source?: string;
};

export type SiteImages = {
  restaurant_id?: string;
  menu_images?: SiteImage[];
  gallery_images?: SiteImage[];
};

export type SiteMenuItem = {
  name?: string;
  category?: string;
  description?: string;
  price?: string;
  image_url?: string;
};

export type SiteReview = {
  reviewer?: string;
  review?: string;
  stars?: number | null;
  date?: string;
};

export type SiteContent = {
  restaurant_id?: string;
  place_id?: string;
  name?: string;
  menu_items?: SiteMenuItem[];
  menu_images?: SiteImage[];
  gallery_images?: SiteImage[];
  reviews?: SiteReview[];
  thumbnail?: string;
};

export type Member = {
  id?: string;
  user_id?: string;
  email?: string;
  role?: string;
  member_role?: string;
  full_name?: string;
  created_at?: string;
};
