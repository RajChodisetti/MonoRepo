package seoreport

import (
	"context"
	"fmt"
	"strings"
	"time"

	llmlib "github.com/rajchodisetti/restaurant-platform/backend/internal/providers/llm"
)

// Summarizer builds a short constructive AI summary.
type Summarizer interface {
	Summarize(ctx context.Context, place PlaceDetails, report Report, reviews []Review) (string, error)
}

// DeterministicSummarizer creates a 3–4 line improvement summary without an LLM.
type DeterministicSummarizer struct{}

func (DeterministicSummarizer) Summarize(_ context.Context, place PlaceDetails, report Report, reviews []Review) (string, error) {
	return buildDeterministicSummary(place, report, reviews), nil
}

// LLMSummarizer tries an LLM first, then falls back to the deterministic summary.
type LLMSummarizer struct {
	Client   llmlib.Client
	Fallback Summarizer
}

func (s LLMSummarizer) Summarize(ctx context.Context, place PlaceDetails, report Report, reviews []Review) (string, error) {
	fallback := s.Fallback
	if fallback == nil {
		fallback = DeterministicSummarizer{}
	}
	if s.Client == nil || !s.Client.Enabled() {
		return fallback.Summarize(ctx, place, report, reviews)
	}

	var b strings.Builder
	fmt.Fprintf(&b, "Restaurant: %s\nAddress: %s\nOverall SEO score: %d/100 (%s)\n", place.Name, place.Address, report.OverallScore, report.OverallLabel)
	for _, m := range report.Metrics {
		fmt.Fprintf(&b, "- %s: %d/%d (%s)\n", m.Label, m.Score, m.Max, m.Status)
	}
	if len(reviews) > 0 {
		b.WriteString("Recent review excerpts:\n")
		for i, r := range reviews {
			if i >= 3 {
				break
			}
			text := strings.TrimSpace(r.Text)
			if len(text) > 180 {
				text = text[:180] + "…"
			}
			fmt.Fprintf(&b, "- %.0f★ %s\n", r.Rating, text)
		}
	}
	prompt := "Write exactly 3 or 4 short lines of constructive feedback for this restaurant's Google/local SEO. " +
		"Focus on concrete places to improve (reviews, website, menu, order online, contact, keywords). " +
		"No bullets, no markdown, no greeting. Plain sentences only.\n\n" + b.String()

	ctx, cancel := context.WithTimeout(ctx, 8*time.Second)
	defer cancel()

	out, err := s.Client.Complete(ctx, prompt)
	if err != nil || strings.TrimSpace(out) == "" {
		return fallback.Summarize(ctx, place, report, reviews)
	}
	return normalizeSummary(out), nil
}

func buildDeterministicSummary(place PlaceDetails, report Report, reviews []Review) string {
	gaps := make([]string, 0, 4)
	for _, m := range report.Metrics {
		if m.Value >= 0.75 {
			continue
		}
		switch m.Key {
		case "seo":
			gaps = append(gaps, "local keywords and cuisine terms in your Google listing and posts")
		case "reviews":
			gaps = append(gaps, "recent 5-star review volume and reply cadence")
		case "website":
			gaps = append(gaps, "homepage design, menu clarity, and booking CTAs on your website")
		case "order_online":
			gaps = append(gaps, "clear order-online / delivery signals")
		case "menu":
			gaps = append(gaps, "structured menu items and menu photos")
		case "contact":
			if place.Phone == "" {
				gaps = append(gaps, "a verified phone number")
			}
			if place.Email == "" {
				gaps = append(gaps, "a public business email")
			}
		case "listing":
			gaps = append(gaps, "complete hours and fresh listing photos")
		}
		if len(gaps) >= 3 {
			break
		}
	}
	if len(gaps) == 0 {
		gaps = append(gaps, "weekly photo freshness and continued review replies")
	}

	line1 := fmt.Sprintf("%s scores %d/100 (%s) on local SEO visibility right now.", place.Name, report.OverallScore, report.OverallLabel)
	line2 := "Biggest gains come from improving " + joinNatural(gaps) + "."
	line3 := "Focus on guest-facing listing completeness so Google can trust and rank you in the Map Pack."
	line4 := "Unlock the full report with a Tuvi membership for a prioritized fix plan."

	if len(reviews) > 0 {
		avg := 0.0
		n := 0
		for _, r := range reviews {
			if r.Rating > 0 {
				avg += r.Rating
				n++
			}
		}
		if n > 0 {
			avg /= float64(n)
			line3 = fmt.Sprintf("Recent guest feedback averages %.1f★ — turn those themes into menu, service, and listing updates.", avg)
		}
	}

	return strings.Join([]string{line1, line2, line3, line4}, "\n")
}

func joinNatural(parts []string) string {
	switch len(parts) {
	case 0:
		return "core listing details"
	case 1:
		return parts[0]
	case 2:
		return parts[0] + " and " + parts[1]
	default:
		return strings.Join(parts[:len(parts)-1], ", ") + ", and " + parts[len(parts)-1]
	}
}

func normalizeSummary(text string) string {
	text = strings.ReplaceAll(text, "\r\n", "\n")
	lines := make([]string, 0, 4)
	for _, line := range strings.Split(text, "\n") {
		line = strings.TrimSpace(line)
		line = strings.TrimLeft(line, "-•* ")
		if line == "" {
			continue
		}
		lines = append(lines, line)
		if len(lines) >= 4 {
			break
		}
	}
	if len(lines) == 0 {
		return strings.TrimSpace(text)
	}
	if len(lines) == 1 {
		// Split long single paragraph into ~3 sentences if possible.
		sentences := splitSentences(lines[0])
		if len(sentences) >= 3 {
			if len(sentences) > 4 {
				sentences = sentences[:4]
			}
			return strings.Join(sentences, "\n")
		}
	}
	return strings.Join(lines, "\n")
}

func splitSentences(text string) []string {
	parts := strings.FieldsFunc(text, func(r rune) bool {
		return r == '.' || r == '!' || r == '?'
	})
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		out = append(out, p+".")
	}
	return out
}
