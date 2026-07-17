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

export type ProfileReviewPreview = {
  restaurant_id: string;
  restaurant_name?: string;
  contact_email?: string;
  ocr_status?: string;
  ocr_input_fingerprint?: string;
  review_status?: string;
  restaurant_updated_at?: string;
  profile_updated_at?: string;
  profile?: Record<string, unknown>;
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
