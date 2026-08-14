package outreach

import (
	"errors"
	"strings"
	"testing"

	"github.com/google/uuid"
)

func TestPrepareInboxReply(t *testing.T) {
	target := Message{
		ID:            uuid.New(),
		Direction:     MessageDirectionInbound,
		FromEmail:     "OWNER@Restaurant.Example",
		Subject:       "Original subject",
		GmailThreadID: "thread-1",
		RFCMessageID:  "<original@example.com>",
	}
	request, err := prepareInboxReply(target, ReplyMessageInput{BodyText: " Thanks for replying. "})
	if err != nil {
		t.Fatalf("prepareInboxReply() error = %v", err)
	}
	if request.To != "owner@restaurant.example" || request.Subject != "Re: Original subject" {
		t.Fatalf("reply recipient/subject = %q/%q", request.To, request.Subject)
	}
	if request.TextBody != "Thanks for replying." || request.ThreadID != "thread-1" {
		t.Fatalf("reply body/thread = %q/%q", request.TextBody, request.ThreadID)
	}
	if request.InReplyTo != target.RFCMessageID || request.References != target.RFCMessageID {
		t.Fatalf("reply RFC headers = %q/%q", request.InReplyTo, request.References)
	}
}

func TestPrepareInboxReplyRejectsUnsafeInput(t *testing.T) {
	base := Message{
		ID:        uuid.New(),
		Direction: MessageDirectionInbound,
		FromEmail: "owner@example.com",
		Subject:   "Hello",
	}
	tests := []struct {
		name   string
		target Message
		input  ReplyMessageInput
	}{
		{name: "outbound target", target: Message{Direction: MessageDirectionOutbound}, input: ReplyMessageInput{BodyText: "reply"}},
		{name: "empty body", target: base, input: ReplyMessageInput{}},
		{name: "header injection", target: base, input: ReplyMessageInput{Subject: "safe\r\nBcc: bad@example.com", BodyText: "reply"}},
		{name: "invalid sender", target: Message{ID: uuid.New(), Direction: MessageDirectionInbound, FromEmail: "bad"}, input: ReplyMessageInput{BodyText: "reply"}},
		{name: "oversized body", target: base, input: ReplyMessageInput{BodyText: strings.Repeat("a", 10001)}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := prepareInboxReply(test.target, test.input)
			if !errors.Is(err, ErrInvalidInboxReply) {
				t.Fatalf("prepareInboxReply() error = %v, want ErrInvalidInboxReply", err)
			}
		})
	}
}
