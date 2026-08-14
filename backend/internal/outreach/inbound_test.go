package outreach

import (
	"testing"

	"github.com/rajchodisetti/restaurant-platform/backend/internal/platform/config"
	emailprovider "github.com/rajchodisetti/restaurant-platform/backend/internal/providers/email"
)

func TestRFCMessageIDsDedupes(t *testing.T) {
	ids := rfcMessageIDs("<one@example.com>", "<one@example.com> <two@example.com>")
	if len(ids) != 2 || ids[0] != "<one@example.com>" || ids[1] != "<two@example.com>" {
		t.Fatalf("rfcMessageIDs() = %#v", ids)
	}
}

func TestStoppedReasonInboundReply(t *testing.T) {
	if StoppedReasonInboundReply != "inbound_reply" {
		t.Fatalf("StoppedReasonInboundReply = %q", StoppedReasonInboundReply)
	}
}

func TestRecentMailQueryUsesBoundedDedicatedInbox(t *testing.T) {
	service := &InboundService{cfg: config.OutreachConfig{
		InboundLocalPart: "outreach",
		InboundDomain:    "tuvisolutions.com",
	}}
	got := service.recentMailQuery()
	if got != "in:inbox newer_than:7d" {
		t.Fatalf("recentMailQuery() = %q", got)
	}
}

func TestAddressedToInboundMailbox(t *testing.T) {
	service := &InboundService{cfg: config.OutreachConfig{
		InboundLocalPart: "outreach",
		InboundDomain:    "tuvisolutions.com",
	}}
	if !service.addressedToInboundMailbox(emailprovider.InboxMessage{To: "outreach+aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee@tuvisolutions.com"}) {
		t.Fatal("plus-address should be kept")
	}
	if service.addressedToInboundMailbox(emailprovider.InboxMessage{To: "contact@tuvisolutions.com"}) {
		t.Fatal("unrelated tuvi mailbox should be skipped")
	}
}

func TestSnapshotBodyPrefersText(t *testing.T) {
	if got := snapshotBody(" plain ", "<p>html</p>"); got != "plain" {
		t.Fatalf("snapshotBody() = %q", got)
	}
	if got := snapshotBody("  ", "<p>html</p>"); got != "<p>html</p>" {
		t.Fatalf("snapshotBody html fallback = %q", got)
	}
}
