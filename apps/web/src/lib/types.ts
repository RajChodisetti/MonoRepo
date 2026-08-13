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
  phone?: string;
  address?: string;
  ocr_status?: string;
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
    job_id: string;
    status: string;
  };
  last_completed_job?: {
    job_id: string;
    status: string;
    summary?: {
      attempted: number;
      sent: number;
      failed: number;
      skipped: number;
      stopped_reason?: string;
    };
  };
  email_job: {
    enabled: boolean;
    enabled_at?: string;
    enabled_by?: string;
    updated_at?: string;
  };
};

export type EmailAccountHealth = {
  account_key: string;
  provider: string;
  provider_identity: string;
  from_email: string;
  enabled: boolean;
  status: string;
  last_checked_at?: string;
  next_check_at?: string;
  provider_message_id?: string;
  last_error?: string;
};

export type EmailAccountHealthResponse = {
  enabled: boolean;
  recipient: string;
  interval_hours: number;
  accounts: EmailAccountHealth[];
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
  ocr_checked?: boolean;
  ocr_input_fingerprint?: string;
  ocr_started_at?: string;
  ocr_completed_at?: string;
  ocr_attempts?: number;
  ocr_verification_errors?: unknown[];
  ocr_images_discovered?: number;
  ocr_images_analyzed?: number;
  ocr_images_succeeded?: number;
  ocr_images_failed?: number;
  ocr_all_images_processed?: boolean;
  ocr_provider?: string;
  ocr_model?: string;
  apollo_status?: string;
  apollo_email_found?: boolean;
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

export type DemoTranscript = {
  id: string;
  role: string;
  content: string;
  occurred_at: string;
};

export type DemoSession = {
  id: string;
  demo_site_id?: string;
  restaurant_id: string;
  template_id: "1" | "2" | "3";
  started_at: string;
  last_seen_at: string;
  ended_at?: string;
  duration_seconds: number;
  transcript: DemoTranscript[];
};

export type GooglePhotoAttribution = {
  display_name: string;
  uri?: string;
  photo_uri?: string;
};

export type GooglePlacePhoto = {
  url: string;
  width_px?: number;
  height_px?: number;
  author_attributions: GooglePhotoAttribution[];
  google_maps_uri?: string;
  flag_content_uri?: string;
};

export type GooglePlacePhotos = {
  restaurant_id: string;
  google_place_id: string;
  photos: GooglePlacePhoto[];
  refreshed_at: string;
  urls_are_temporary: boolean;
};

export type RestaurantImage = {
  id: string;
  restaurant_id: string;
  url: string;
  thumbnail_url?: string;
  image_type?: string;
  title?: string;
  source?: string;
  sort_order?: number;
  created_at?: string;
  updated_at?: string;
  hidden_at?: string;
  hidden_by?: string;
};

export type RestaurantImages = {
  restaurant_id: string;
  menu_images: RestaurantImage[];
  gallery_images: RestaurantImage[];
  owned_media: RestaurantOwnedMedia[];
};

export type RestaurantOwnedMedia = {
  id: string;
  url: string;
  source_kind: "owner_upload" | "licensed";
  media_type: "exterior" | "interior" | "food" | "drink" | "logo" | "team" | "event" | "other";
  caption?: string;
  alt_text?: string;
  tags?: string[];
  quality_score?: number;
  hero_score?: number;
  width_px?: number;
  height_px?: number;
  orientation?: string;
  placement_role?: string;
  approval_status?: "draft" | "approved" | "rejected";
  rights_status?: "owner_granted" | "licensed";
  vision_status?: "pending" | "running" | "verified" | "failed";
  vision_last_error?: string;
  vision_analyzed_at?: string;
  hidden_at?: string;
};

export type DemoLink = {
  demo_site_id: string;
  slug: string;
  status: string;
  expires_at?: string;
  preview_url?: string;
  created_at: string;
  updated_at: string;
};

export type GeneratedSiteTemplate = {
  id: string;
  name: string;
  url: string;
};

export type GeneratedSite = {
  restaurant_id: string;
  restaurant_name: string;
  google_place_id: string;
  site_index: number;
  templates: GeneratedSiteTemplate[];
  shareable: boolean;
};

export type AdHocPreview = {
  restaurant_id: string;
  restaurant_name?: string;
  recipient_email?: string;
  subject: string;
  body_html: string;
  body_text: string;
};

export type AdHocSendResult = {
  restaurant_id: string;
  sent: boolean;
  error?: string;
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

export type EmailMessage = {
  id: string;
  restaurant_id?: string;
  campaign_id?: string;
  delivery_attempt_id?: string;
  reply_token?: string;
  direction: "outbound" | "inbound";
  from_email: string;
  to_email: string;
  reply_to?: string;
  subject: string;
  body_text: string;
  gmail_message_id?: string;
  gmail_thread_id?: string;
  rfc_message_id?: string;
  mailbox_key?: string;
  unmatched?: boolean;
  read_at?: string;
  created_at: string;
};

export type InboxThread = {
  restaurant_id?: string;
  restaurant_name?: string;
  email?: string;
  unmatched: boolean;
  unread_count: number;
  last_direction: string;
  last_snippet: string;
  last_at: string;
  last_message_id: string;
};

export type InboxListResponse = {
  threads: InboxThread[];
  total: number;
};
