package outreach

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/rajchodisetti/restaurant-platform/backend/internal/campaigns"
	"github.com/rajchodisetti/restaurant-platform/backend/internal/platform/config"
	emailprovider "github.com/rajchodisetti/restaurant-platform/backend/internal/providers/email"
)

type InboundService struct {
	store      inboundStore
	campaigns  campaigns.Repository
	reader     emailprovider.InboxReader
	mailboxKey string
	cfg        config.OutreachConfig
	log        *slog.Logger
}

type inboundStore interface {
	GetInboundSync(context.Context, string) (string, error)
	SetInboundSync(context.Context, string, string) error
	MarkInboundPollAttempt(context.Context, string) error
	RecordInboundPollResult(context.Context, string, error) error
	GetMessageByGmailID(context.Context, string, string) (Message, error)
	GetOutboundByReplyToken(context.Context, uuid.UUID) (Message, error)
	GetOutboundByRFCMessageID(context.Context, string) (Message, error)
	GetOutboundByThreadID(context.Context, string) (Message, error)
	FindRestaurantIDByEmail(context.Context, string) (uuid.UUID, error)
	InsertMessage(context.Context, Message) (Message, error)
}

func NewInboundService(
	store inboundStore,
	campaignsRepo campaigns.Repository,
	reader emailprovider.InboxReader,
	mailboxKey string,
	cfg config.OutreachConfig,
	log *slog.Logger,
) *InboundService {
	if log == nil {
		log = slog.Default()
	}
	return &InboundService{
		store:      store,
		campaigns:  campaignsRepo,
		reader:     reader,
		mailboxKey: strings.TrimSpace(mailboxKey),
		cfg:        cfg,
		log:        log,
	}
}

func (service *InboundService) Poll(ctx context.Context) error {
	if service == nil || service.reader == nil || service.store == nil {
		return nil
	}
	if !service.cfg.InboundEnabled {
		return nil
	}
	mailboxKey := strings.TrimSpace(service.mailboxKey)
	if mailboxKey == "" {
		return fmt.Errorf("inbound mailbox key is required")
	}
	if err := service.store.MarkInboundPollAttempt(ctx, mailboxKey); err != nil {
		return err
	}
	pollErr := service.pollMailbox(ctx, mailboxKey)
	statusCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 5*time.Second)
	defer cancel()
	statusErr := service.store.RecordInboundPollResult(statusCtx, mailboxKey, pollErr)
	return errors.Join(pollErr, statusErr)
}

func (service *InboundService) pollMailbox(ctx context.Context, mailboxKey string) error {
	startHistoryID, err := service.store.GetInboundSync(ctx, mailboxKey)
	if err != nil {
		return err
	}

	var messageIDs []string
	newHistoryID := startHistoryID
	if startHistoryID != "" {
		messageIDs, newHistoryID, err = service.reader.ListHistoryMessageIDs(ctx, startHistoryID)
		if err != nil {
			service.log.WarnContext(ctx, "outreach_inbound_history_unavailable", "error", err)
			messageIDs, newHistoryID, err = service.fullInboxSync(ctx)
		}
	} else {
		messageIDs, newHistoryID, err = service.fullInboxSync(ctx)
	}
	if err != nil {
		return err
	}

	var captureErr error
	for _, messageID := range messageIDs {
		if err := service.capture(ctx, mailboxKey, messageID); err != nil {
			service.log.ErrorContext(ctx, "outreach_inbound_capture_failed", "gmail_message_id", messageID, "error", err)
			captureErr = errors.Join(captureErr, err)
		}
	}
	if captureErr != nil {
		return fmt.Errorf("capture inbound outreach messages: %w", captureErr)
	}
	if strings.TrimSpace(newHistoryID) != "" {
		return service.store.SetInboundSync(ctx, mailboxKey, newHistoryID)
	}
	return nil
}

func (service *InboundService) fullInboxSync(ctx context.Context) ([]string, string, error) {
	// Read the history cursor before listing. Messages arriving during the list
	// are therefore present either in the list or in the next history delta.
	historyID, err := service.reader.ProfileHistoryID(ctx)
	if err != nil {
		return nil, "", err
	}
	messageIDs, err := service.reader.ListRecentMessageIDs(ctx, service.recentMailQuery())
	if err != nil {
		return nil, "", err
	}
	return messageIDs, historyID, nil
}

func (service *InboundService) capture(ctx context.Context, mailboxKey, gmailMessageID string) error {
	if existing, err := service.store.GetMessageByGmailID(ctx, mailboxKey, gmailMessageID); err == nil {
		if existing.Direction == MessageDirectionInbound &&
			!existing.Unmatched &&
			existing.CampaignID != nil &&
			service.campaigns != nil {
			return service.stopCampaignOnReply(ctx, *existing.CampaignID)
		}
		return nil
	} else if !errors.Is(err, pgx.ErrNoRows) {
		return err
	}

	inbound, err := service.reader.GetMessage(ctx, gmailMessageID)
	if err != nil {
		return err
	}
	if inbound.ID == "" {
		inbound.ID = gmailMessageID
	}
	if !inbound.Inbox || inbound.ReceivedAt.IsZero() || inbound.ReceivedAt.Before(time.Now().UTC().Add(-10*24*time.Hour)) {
		return nil
	}

	match := service.matchInbound(ctx, inbound)
	record := Message{
		Direction:      MessageDirectionInbound,
		FromEmail:      inbound.From,
		ToEmail:        firstNonEmpty(inbound.DeliveredTo, inbound.To),
		Subject:        inbound.Subject,
		BodyText:       inbound.BodyText,
		GmailMessageID: inbound.ID,
		GmailThreadID:  inbound.ThreadID,
		RFCMessageID:   inbound.RFCMessageID,
		MailboxKey:     mailboxKey,
		Unmatched:      match.outbound == nil && match.restaurantID == uuid.Nil,
		ReceivedAt:     inbound.ReceivedAt,
	}
	if match.outbound != nil {
		record.RestaurantID = match.outbound.RestaurantID
		record.CampaignID = match.outbound.CampaignID
		record.DeliveryAttemptID = match.outbound.DeliveryAttemptID
		record.ReplyToken = match.outbound.ReplyToken
		if record.GmailThreadID == "" {
			record.GmailThreadID = match.outbound.GmailThreadID
		}
	} else if match.restaurantID != uuid.Nil {
		id := match.restaurantID
		record.RestaurantID = &id
	}

	if _, err := service.store.InsertMessage(ctx, record); err != nil {
		return err
	}
	if record.Unmatched || record.CampaignID == nil || service.campaigns == nil {
		return nil
	}
	return service.stopCampaignOnReply(ctx, *record.CampaignID)
}

type inboundMatch struct {
	outbound     *Message
	restaurantID uuid.UUID
}

func (service *InboundService) matchInbound(ctx context.Context, inbound emailprovider.InboxMessage) inboundMatch {
	addresses := strings.Join([]string{inbound.To, inbound.DeliveredTo}, ",")
	if token, ok := emailprovider.ParseReplyToken(addresses, service.cfg.InboundLocalPart, service.cfg.InboundDomain); ok {
		if outbound, err := service.store.GetOutboundByReplyToken(ctx, token); err == nil {
			return inboundMatch{outbound: &outbound}
		}
	}
	for _, rfcID := range rfcMessageIDs(inbound.InReplyTo, inbound.References) {
		if outbound, err := service.store.GetOutboundByRFCMessageID(ctx, rfcID); err == nil {
			return inboundMatch{outbound: &outbound}
		}
	}
	if outbound, err := service.store.GetOutboundByThreadID(ctx, inbound.ThreadID); err == nil {
		return inboundMatch{outbound: &outbound}
	}
	if restaurantID, err := service.store.FindRestaurantIDByEmail(ctx, inbound.From); err == nil {
		return inboundMatch{restaurantID: restaurantID}
	}
	return inboundMatch{}
}

func (service *InboundService) stopCampaignOnReply(ctx context.Context, campaignID uuid.UUID) error {
	campaign, err := service.campaigns.GetByID(ctx, campaignID)
	if err != nil {
		return err
	}
	if campaign.Status == campaigns.StatusStopped {
		return nil
	}
	_, err = service.campaigns.Stop(ctx, campaignID, StoppedReasonInboundReply)
	if err != nil {
		return fmt.Errorf("stop campaign after inbound reply: %w", err)
	}
	service.log.InfoContext(ctx, "outreach_inbound_reply_paused_campaign", "campaign_id", campaignID)
	return nil
}

func (service *InboundService) recentMailQuery() string {
	return "in:inbox newer_than:10d"
}

func rfcMessageIDs(values ...string) []string {
	seen := make(map[string]struct{})
	ids := make([]string, 0)
	for _, value := range values {
		for _, part := range strings.Fields(value) {
			part = strings.TrimSpace(part)
			if part == "" {
				continue
			}
			key := strings.ToLower(part)
			if _, exists := seen[key]; exists {
				continue
			}
			seen[key] = struct{}{}
			ids = append(ids, part)
		}
	}
	return ids
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}
