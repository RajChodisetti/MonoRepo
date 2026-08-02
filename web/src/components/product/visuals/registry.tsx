import type { JSX } from "react";
import type { VisualId } from "@/content/products/types";
import {
  AiSearchMockVisual,
  GyroPreviewVisual,
  OrderingRavioliVisual,
  PhoneImprovingVisual,
} from "@/components/product/visuals/WebsiteVisuals";
import {
  ExpertsAvatarsVisual,
  GoogleUpdateVisual,
  SeoAiSearchVisual,
  SeoScoreVisual,
} from "@/components/product/visuals/SeoVisuals";
import {
  MenuItemsStackVisual,
  OrderTrackingPhoneVisual,
  PitaWrapsMenuVisual,
  RewardsPhoneVisual,
} from "@/components/product/visuals/MenuVisuals";
import {
  CustomersFlowVisual,
  GoogleReviewsStackVisual,
  ReviewsOwnerPhotoVisual,
  ReviewsPhoneMockupVisual,
} from "@/components/product/visuals/ReviewsVisuals";
import {
  AddressFixCardsVisual,
  ListingMapCardVisual,
  ListingsExpertsPhotoVisual,
  ListingsSyncedVisual,
} from "@/components/product/visuals/ListingsVisuals";
import {
  FeeSavingsToastVisual,
  OrderingAppShowcaseVisual,
  OrderingCustomerListVisual,
  OrderingPhonePreviewVisual,
} from "@/components/product/visuals/OrderingVisuals";
import {
  UpsellCheckoutVisual,
  UpsellDataAvatarsVisual,
  UpsellFlowVisual,
  UpsellImprovingPhotoVisual,
} from "@/components/product/visuals/UpsellsVisuals";
import {
  DeliveryControlMapVisual,
  DeliveryGuestPhotoVisual,
  DeliveryMapPhoneVisual,
  DeliveryTrackingCardVisual,
} from "@/components/product/visuals/DeliveryVisuals";
import {
  CateringFoodCollageVisual,
  CateringMenuStackVisual,
  CateringPhoneMockupVisual,
  CateringSearchVisual,
} from "@/components/product/visuals/CateringVisuals";
import {
  AiPhoneConversationVisual,
  AiPhoneFoodTilesVisual,
  AiPhoneLoyaltyPhotoVisual,
  AiPhoneMockupVisual,
} from "@/components/product/visuals/AiPhoneVisuals";
import {
  AnalyticsChartVisual,
  AppPhotoFillVisual,
  BrandedAppPhoneVisual,
  BrandedAppShowcaseVisual,
  CampaignCalendarVisual,
  CampaignPromoVisual,
  EmailSmsPreviewVisual,
  KitchenPhotoFillVisual,
  KitchenTicketVisual,
  LoyaltyCardVisual,
  LoyaltyRewardsGridVisual,
  OwnerDashboardVisual,
  OwnerPhotoFillVisual,
  PosSyncVisual,
  PushNotifStackVisual,
} from "@/components/product/visuals/OpsVisuals";

const VISUAL_MAP: Record<VisualId, () => JSX.Element> = {
  "gyro-preview": GyroPreviewVisual,
  "ai-search-mock": AiSearchMockVisual,
  "ordering-ravioli": OrderingRavioliVisual,
  "phone-improving": PhoneImprovingVisual,
  "seo-score": SeoScoreVisual,
  "seo-ai-search": SeoAiSearchVisual,
  "google-update": GoogleUpdateVisual,
  "experts-avatars": ExpertsAvatarsVisual,
  "menu-items-stack": MenuItemsStackVisual,
  "rewards-phone": RewardsPhoneVisual,
  "pita-wraps-menu": PitaWrapsMenuVisual,
  "order-tracking-phone": OrderTrackingPhoneVisual,
  "reviews-owner-photo": ReviewsOwnerPhotoVisual,
  "google-reviews-stack": GoogleReviewsStackVisual,
  "customers-flow": CustomersFlowVisual,
  "reviews-phone-mockup": ReviewsPhoneMockupVisual,
  "listing-map-card": ListingMapCardVisual,
  "listings-synced": ListingsSyncedVisual,
  "address-fix-cards": AddressFixCardsVisual,
  "listings-experts-photo": ListingsExpertsPhotoVisual,
  "ordering-phone-preview": OrderingPhonePreviewVisual,
  "ordering-app-showcase": OrderingAppShowcaseVisual,
  "ordering-customer-list": OrderingCustomerListVisual,
  "fee-savings-toast": FeeSavingsToastVisual,
  "upsell-flow": UpsellFlowVisual,
  "upsell-checkout": UpsellCheckoutVisual,
  "upsell-data-avatars": UpsellDataAvatarsVisual,
  "upsell-improving-photo": UpsellImprovingPhotoVisual,
  "delivery-map-phone": DeliveryMapPhoneVisual,
  "delivery-tracking-card": DeliveryTrackingCardVisual,
  "delivery-control-map": DeliveryControlMapVisual,
  "delivery-guest-photo": DeliveryGuestPhotoVisual,
  "catering-menu-stack": CateringMenuStackVisual,
  "catering-search": CateringSearchVisual,
  "catering-food-collage": CateringFoodCollageVisual,
  "catering-phone-mockup": CateringPhoneMockupVisual,
  "ai-phone-mockup": AiPhoneMockupVisual,
  "ai-phone-conversation": AiPhoneConversationVisual,
  "ai-phone-loyalty-photo": AiPhoneLoyaltyPhotoVisual,
  "ai-phone-food-tiles": AiPhoneFoodTilesVisual,
  "branded-app-phone": BrandedAppPhoneVisual,
  "branded-app-showcase": BrandedAppShowcaseVisual,
  "app-photo-fill": AppPhotoFillVisual,
  "campaign-promo": CampaignPromoVisual,
  "campaign-calendar": CampaignCalendarVisual,
  "email-sms-preview": EmailSmsPreviewVisual,
  "push-notif-stack": PushNotifStackVisual,
  "loyalty-card": LoyaltyCardVisual,
  "loyalty-rewards-grid": LoyaltyRewardsGridVisual,
  "owner-dashboard": OwnerDashboardVisual,
  "owner-photo-fill": OwnerPhotoFillVisual,
  "analytics-chart": AnalyticsChartVisual,
  "kitchen-ticket": KitchenTicketVisual,
  "kitchen-photo-fill": KitchenPhotoFillVisual,
  "pos-sync": PosSyncVisual,
};

export function ProductVisual({ id }: { id: VisualId }) {
  const Visual = VISUAL_MAP[id];
  return <Visual />;
}
