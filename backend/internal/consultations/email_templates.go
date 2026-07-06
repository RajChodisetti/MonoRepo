package consultations

import (
	"fmt"
	"html"
	"strings"
)

const siteURL = "https://tuvisolutions.com"

type bookingEmailContent struct {
	Subject  string
	TextBody string
	HTMLBody string
}

func buildProspectConfirmationEmail(
	name, email, phone, code, when, dayLabel, timeLabel, calendarLink string,
) bookingEmailContent {
	safeName := html.EscapeString(strings.TrimSpace(name))
	if safeName == "" {
		safeName = "there"
	}
	safeEmail := html.EscapeString(strings.TrimSpace(email))
	safePhone := html.EscapeString(strings.TrimSpace(phone))
	safeCode := html.EscapeString(code)
	safeWhen := html.EscapeString(when)
	safeDay := html.EscapeString(dayLabel)
	safeTime := html.EscapeString(timeLabel)

	calendarBlockText := ""
	calendarButton := ""
	if strings.TrimSpace(calendarLink) != "" {
		safeCal := html.EscapeString(calendarLink)
		calendarBlockText = fmt.Sprintf("\nAdd to Google Calendar: %s\n", calendarLink)
		calendarButton = fmt.Sprintf(`
            <tr>
              <td align="center" style="padding:12px 28px 4px;">
                <a href="%s" style="display:inline-block;background:#d4a853;color:#111111;text-decoration:none;font-weight:700;font-size:14px;padding:14px 28px;border-radius:10px;">
                  Add to Google Calendar
                </a>
              </td>
            </tr>
            <tr>
              <td align="center" style="padding:6px 28px 0;">
                <p style="margin:0;font-size:12px;color:#6b7280;">Opens Google Calendar with Tuvi Consultation at your booked time</p>
              </td>
            </tr>`, safeCal)
	}

	subject := fmt.Sprintf("You're booked — Tuvi consultation confirmed (%s)", code)

	text := fmt.Sprintf(`Hi %s,

You're all set — your free consultation with Tuvi Solutions is confirmed.

APPOINTMENT
Day: %s
Time: %s
Confirmation ID: %s
Email: %s
Phone: %s
%s
TRY WHAT WE'VE BUILT
Explore the tools and experiences we've designed for businesses ready to scale:

• Custom software & premium digital experiences
• AI / ML systems and voice assistants (like the one you just used)
• Web & mobile apps engineered for growth
• Data-driven consulting and seamless integrations

Start exploring: %s

We've also got a $1,000 risk-free trial on your first phase of work — we prove value before you commit.

See you on the call.

— The Tuvi Solutions team
%s
`, name, dayLabel, timeLabel, code, email, phone, calendarBlockText, siteURL, siteURL)

	htmlBody := fmt.Sprintf(`<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="UTF-8" />
<meta name="viewport" content="width=device-width, initial-scale=1.0" />
<title>Consultation confirmed</title>
</head>
<body style="margin:0;padding:0;background:#ffffff;font-family:-apple-system,BlinkMacSystemFont,'Segoe UI',Roboto,Helvetica,Arial,sans-serif;color:#111111;">
  <table role="presentation" width="100%%" cellspacing="0" cellpadding="0" style="background:#ffffff;padding:28px 12px;">
    <tr>
      <td align="center">
        <table role="presentation" width="100%%" cellspacing="0" cellpadding="0" style="max-width:560px;background:#ffffff;border:1px solid #e6e8ec;border-radius:16px;overflow:hidden;">
          <tr>
            <td align="center" style="padding:32px 28px 8px;">
              <div style="width:56px;height:56px;border-radius:50%%;background:#e8faf2;line-height:56px;font-size:28px;color:#059669;">&#10003;</div>
              <p style="margin:16px 0 0;font-size:11px;letter-spacing:0.18em;font-weight:700;color:#0d9488;text-transform:uppercase;">Consultation booked</p>
              <h1 style="margin:10px 0 0;font-size:26px;line-height:1.25;font-weight:700;color:#111111;">You're all set, %s!</h1>
              <p style="margin:10px 0 0;font-size:14px;line-height:1.5;color:#5b6470;">Confirmation email sent to <strong style="color:#111111;">%s</strong></p>
            </td>
          </tr>

          <tr>
            <td style="padding:20px 28px 8px;">
              <table role="presentation" width="100%%" cellspacing="0" cellpadding="0" style="background:#f7f8fa;border:1px solid #e6e8ec;border-radius:14px;">
                <tr>
                  <td style="padding:16px 18px;border-bottom:1px solid #e6e8ec;">
                    <p style="margin:0;font-size:11px;letter-spacing:0.14em;font-weight:700;color:#6b7280;text-transform:uppercase;">Day</p>
                    <p style="margin:6px 0 0;font-size:16px;font-weight:600;color:#111111;">%s</p>
                  </td>
                </tr>
                <tr>
                  <td style="padding:16px 18px;">
                    <p style="margin:0;font-size:11px;letter-spacing:0.14em;font-weight:700;color:#6b7280;text-transform:uppercase;">Time</p>
                    <p style="margin:6px 0 0;font-size:16px;font-weight:600;color:#111111;">%s</p>
                    <p style="margin:4px 0 0;font-size:12px;color:#6b7280;">%s · Australia/Sydney</p>
                  </td>
                </tr>
              </table>
            </td>
          </tr>

          <tr>
            <td style="padding:12px 28px 8px;">
              <table role="presentation" width="100%%" cellspacing="0" cellpadding="0" style="background:#fffaf0;border:1px solid #e8c97a;border-radius:14px;">
                <tr>
                  <td align="center" style="padding:18px;">
                    <p style="margin:0;font-size:11px;letter-spacing:0.14em;font-weight:700;color:#8a6d1f;text-transform:uppercase;">Confirmation ID</p>
                    <p style="margin:8px 0 0;font-size:28px;letter-spacing:0.16em;font-weight:800;color:#b8860b;font-family:ui-monospace,SFMono-Regular,Menlo,Monaco,Consolas,monospace;">%s</p>
                  </td>
                </tr>
              </table>
            </td>
          </tr>

          %s

          <tr>
            <td style="padding:24px 28px 8px;">
              <p style="margin:0 0 10px;font-size:12px;letter-spacing:0.14em;font-weight:700;color:#0d9488;text-transform:uppercase;">Try what we've built</p>
              <p style="margin:0 0 14px;font-size:14px;line-height:1.55;color:#374151;">
                Before we meet, explore the products and features Tuvi ships for teams that want to move faster — including the same style of AI voice experience you just used.
              </p>
              <table role="presentation" width="100%%" cellspacing="0" cellpadding="0" style="background:#f7f8fa;border:1px solid #e6e8ec;border-radius:14px;">
                <tr><td style="padding:12px 16px;border-bottom:1px solid #e6e8ec;font-size:14px;color:#111111;">Custom software that streamlines operations</td></tr>
                <tr><td style="padding:12px 16px;border-bottom:1px solid #e6e8ec;font-size:14px;color:#111111;">AI / ML systems and voice assistants</td></tr>
                <tr><td style="padding:12px 16px;border-bottom:1px solid #e6e8ec;font-size:14px;color:#111111;">Premium web &amp; mobile experiences</td></tr>
                <tr><td style="padding:12px 16px;font-size:14px;color:#111111;">Data-driven consulting &amp; integrations</td></tr>
              </table>
              <p style="margin:14px 0 0;font-size:13px;line-height:1.5;color:#5b6470;">
                Plus our <strong style="color:#b8860b;">$1,000 risk-free trial</strong> — we prove value on the first phase before you commit.
              </p>
            </td>
          </tr>

          <tr>
            <td align="center" style="padding:18px 28px 8px;">
              <a href="%s" style="display:inline-block;background:#0d9488;color:#ffffff;text-decoration:none;font-weight:700;font-size:14px;padding:13px 26px;border-radius:10px;">
                Explore Tuvi Solutions
              </a>
            </td>
          </tr>

          <tr>
            <td style="padding:20px 28px 28px;">
              <p style="margin:0;font-size:14px;line-height:1.55;color:#374151;">We're excited to speak with you. If you need to reschedule, just reply to this email.</p>
              <p style="margin:16px 0 0;font-size:13px;color:#5b6470;">— The Tuvi Solutions team</p>
              <p style="margin:8px 0 0;font-size:12px;"><a href="%s" style="color:#0d9488;text-decoration:none;">tuvisolutions.com</a></p>
              <p style="margin:16px 0 0;font-size:11px;color:#9ca3af;">Phone on file: %s</p>
            </td>
          </tr>
        </table>
      </td>
    </tr>
  </table>
</body>
</html>`, safeName, safeEmail, safeDay, safeTime, safeWhen, safeCode, calendarButton, siteURL, siteURL, safePhone)

	return bookingEmailContent{Subject: subject, TextBody: text, HTMLBody: htmlBody}
}

func buildInternalBookingNotifyEmail(
	name, email, phone, code, when, source, calendarLink string,
) bookingEmailContent {
	safeName := html.EscapeString(name)
	safeEmail := html.EscapeString(email)
	safePhone := html.EscapeString(phone)
	safeCode := html.EscapeString(code)
	safeWhen := html.EscapeString(when)
	safeSource := html.EscapeString(source)

	calendarText := ""
	calendarHTML := ""
	if strings.TrimSpace(calendarLink) != "" {
		calendarText = fmt.Sprintf("\nCalendar: %s\n", calendarLink)
		calendarHTML = fmt.Sprintf(`<p><a href="%s">Open calendar event</a></p>`, html.EscapeString(calendarLink))
	}

	subject := fmt.Sprintf("New consultation booked — %s", code)
	text := fmt.Sprintf(`A new consultation has been booked.

Confirmation: %s
Name: %s
Email: %s
Phone: %s
When: %s
Source: %s%s`, code, name, email, phone, when, source, calendarText)

	htmlBody := fmt.Sprintf(`<div style="font-family:-apple-system,BlinkMacSystemFont,'Segoe UI',Roboto,sans-serif;color:#111;">
  <h2 style="margin:0 0 12px;">New consultation booked</h2>
  <ul style="padding-left:18px;line-height:1.7;">
    <li><strong>Confirmation:</strong> %s</li>
    <li><strong>Name:</strong> %s</li>
    <li><strong>Email:</strong> %s</li>
    <li><strong>Phone:</strong> %s</li>
    <li><strong>When:</strong> %s</li>
    <li><strong>Source:</strong> %s</li>
  </ul>
  %s
</div>`, safeCode, safeName, safeEmail, safePhone, safeWhen, safeSource, calendarHTML)

	return bookingEmailContent{Subject: subject, TextBody: text, HTMLBody: htmlBody}
}
