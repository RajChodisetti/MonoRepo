package email

import (
	"strings"
	"testing"

	"github.com/google/uuid"
)

func TestReplyToAddressAndParse(t *testing.T) {
	token := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	address := ReplyToAddress("outreach", "tuvisolutions.com", token)
	if address != "outreach+11111111-1111-1111-1111-111111111111@tuvisolutions.com" {
		t.Fatalf("ReplyToAddress() = %q", address)
	}
	parsed, ok := ParseReplyToken(address, "outreach", "tuvisolutions.com")
	if !ok || parsed != token {
		t.Fatalf("ParseReplyToken() = %s, %v", parsed, ok)
	}
}

func TestParseReplyTokenFromAddressList(t *testing.T) {
	token := uuid.MustParse("22222222-2222-2222-2222-222222222222")
	raw := `Owner <owner@restaurant.example>, outreach+22222222-2222-2222-2222-222222222222@tuvisolutions.com`
	parsed, ok := ParseReplyToken(raw, "outreach", "tuvisolutions.com")
	if !ok || parsed != token {
		t.Fatalf("ParseReplyToken() = %s, %v", parsed, ok)
	}
}

func TestParseReplyTokenRejectsInvalidUUID(t *testing.T) {
	if _, ok := ParseReplyToken("outreach+not-a-uuid@tuvisolutions.com", "outreach", "tuvisolutions.com"); ok {
		t.Fatal("expected invalid plus-address token to be rejected")
	}
	if _, ok := ParseReplyToken("sales+11111111-1111-1111-1111-111111111111@tuvisolutions.com", "outreach", "tuvisolutions.com"); ok {
		t.Fatal("expected a different local-part to be rejected")
	}
}

func TestParseReplyTokenEmpty(t *testing.T) {
	if got := ReplyToAddress("", "tuvisolutions.com", uuid.New()); got != "" {
		t.Fatalf("empty local-part produced %q", got)
	}
	if _, ok := ParseReplyToken("", "outreach", "tuvisolutions.com"); ok {
		t.Fatal("empty address should not parse")
	}
	if !strings.Contains(ReplyToAddress("outreach", "tuvisolutions.com", uuid.MustParse("33333333-3333-3333-3333-333333333333")), "33333333-3333-3333-3333-333333333333") {
		t.Fatal("expected token in reply-to address")
	}
}
