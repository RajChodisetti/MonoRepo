package outreach

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

type nullRestaurantInboxRow struct{}

func (nullRestaurantInboxRow) Scan(dest ...any) error {
	*(dest[0].(**uuid.UUID)) = nil
	*(dest[1].(**string)) = nil
	*(dest[2].(*string)) = "owner@example.com"
	*(dest[3].(*string)) = "sales-one"
	*(dest[4].(*string)) = "sales@example.com"
	*(dest[5].(*bool)) = true
	*(dest[6].(*int)) = 1
	*(dest[7].(*string)) = MessageDirectionInbound
	*(dest[8].(*string)) = "Hello"
	*(dest[9].(*time.Time)) = time.Date(2026, 8, 14, 1, 2, 3, 0, time.UTC)
	*(dest[10].(*uuid.UUID)) = uuid.MustParse("aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee")
	return nil
}

func TestScanInboxThreadAcceptsUnmatchedMessageWithoutRestaurantName(t *testing.T) {
	thread, err := scanInboxThread(nullRestaurantInboxRow{})
	if err != nil {
		t.Fatalf("scanInboxThread() error = %v", err)
	}
	if thread.RestaurantName != "" || !thread.Unmatched || thread.MailboxKey != "sales-one" || thread.MailboxEmail != "sales@example.com" {
		t.Fatalf("scanInboxThread() = %#v", thread)
	}
}
