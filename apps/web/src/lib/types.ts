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
  restaurant_updated_at?: string;
  profile_updated_at?: string;
  profile?: Record<string, unknown>;
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
};

export type GooglePlacePhoto = {
  url: string;
  width_px?: number;
  height_px?: number;
  author_attributions: GooglePhotoAttribution[];
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
