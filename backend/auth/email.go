package auth

import (
	"fmt"
	"log"

	"github.com/resend/resend-go/v2"
)

// sendEmail sends an email via Resend. Returns silently on failure (best-effort).
func sendEmail(to, subject, htmlBody string) {
	if secrets.ResendAPIKey == "" {
		log.Println("[email] ResendAPIKey not configured, skipping email to", to)
		return
	}

	client := resend.NewClient(secrets.ResendAPIKey)
	params := &resend.SendEmailRequest{
		From:    "Comics Galore <noreply@comicsgalore.com>",
		To:      []string{to},
		Subject: subject,
		Html:    htmlBody,
	}

	resp, err := client.Emails.Send(params)
	if err != nil {
		log.Printf("[email] failed to send to %s: %v", to, err)
		return
	}
	log.Printf("[email] sent to %s: %s", to, resp.Id)
}

func sendNewComicFromFollowingEmail(to, title string) {
	subject := "New comic from an uploader you follow"
	body := fmt.Sprintf(`
		<p>A new comic you might like was just published:</p>
		<p style="font-size:18px;font-weight:600">%s</p>
		<p><a href="https://comicsgalore.com/comics">Read it on Comics Galore</a></p>
	`, title)
	sendEmail(to, subject, body)
}

func sendSupportReplyEmail(to, subject string) {
	body := fmt.Sprintf(`
		<p>A member of our support team replied to your ticket:</p>
		<p style="font-size:18px;font-weight:600">%s</p>
		<p><a href="https://comicsgalore.com/support">View the full conversation</a></p>
	`, subject)
	sendEmail(to, "Support reply — Comics Galore", body)
}
