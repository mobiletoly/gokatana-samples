package mailer

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/mobiletoly/gokatana-samples/iamservice/internal/core/app"
	"github.com/mobiletoly/gokatana-samples/iamservice/internal/core/outport"
	"github.com/mobiletoly/gokatana/katapp"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/gmail/v1"
	"google.golang.org/api/option"
)

type mailerAdapter struct {
	cfg *app.GCloudConfig
	srv *gmail.Service
}

// MockEmail represents an email that was sent during testing
type MockEmail struct {
	To          string    `json:"to"`
	Subject     string    `json:"subject"`
	Body        string    `json:"body"`
	ContentType string    `json:"contentType"`
	SentAt      time.Time `json:"sentAt"`
}

func NewMailer(ctx context.Context, cfg *app.GCloudConfig) outport.Mailer {
	if cfg.Mock {
		return &mailerAdapter{
			cfg: cfg,
			srv: nil, // No Gmail service needed in mock mode
		}
	}

	// Read the service account JSON key payload
	jsonConfig := ([]byte)(cfg.ServiceJson)
	jwtCfg, err := google.JWTConfigFromJSON(jsonConfig, gmail.GmailSendScope)
	if err != nil {
		katapp.Logger(ctx).Fatalf("failed to create JWT config from JSON: %v", err)
	}

	// Impersonate the target user in your Workspace domain
	jwtCfg.Subject = cfg.Email.User
	client := jwtCfg.Client(ctx)

	// Create the Gmail service
	srv, err := gmail.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		katapp.Logger(ctx).Fatalf("failed to create JWT config from JSON: %v", err)
	}

	return &mailerAdapter{
		cfg: cfg,
		srv: srv,
	}
}

func (m *mailerAdapter) SendEmail(ctx context.Context, to string, content *outport.MailContent) error {
	katapp.Logger(ctx).Info("sending email", "to", to, "title", content.Title)

	rawMessage := fmt.Sprintf(
		"From: %s\r\n"+
			"To: %s\r\n"+
			"Subject: %s\r\n"+
			"MIME-Version: 1.0\r\n"+
			"Content-Type: %s; charset=\"UTF-8\"\r\n\r\n%s",
		m.cfg.Email.From,
		to,
		content.Title,
		content.ContentType,
		content.Body,
	)

	msg := &gmail.Message{
		Raw: base64.URLEncoding.EncodeToString([]byte(rawMessage)),
	}

	if m.cfg.Mock {
		return m.saveMockEmail(ctx, to, content)
	}

	if _, err := m.srv.Users.Messages.Send("me", msg).Do(); err != nil {
		katapp.Logger(ctx).Error("failed to send email", "error", err)
		return err
	}
	katapp.Logger(ctx).Info("email sent successfully", "to", to)
	return nil
}

// saveMockEmail saves email content to a file for testing purposes
func (m *mailerAdapter) saveMockEmail(ctx context.Context, to string, content *outport.MailContent) error {
	katapp.Logger(ctx).Info("mock mode: saving email to file", "to", to, "subject", content.Title)

	mockEmail := MockEmail{
		To:          to,
		Subject:     content.Title,
		Body:        content.Body,
		ContentType: content.ContentType,
		SentAt:      time.Now(),
	}

	// Create test-emails directory if it doesn't exist
	emailDir := "test-emails"
	if err := os.MkdirAll(emailDir, 0755); err != nil {
		katapp.Logger(ctx).Error("failed to create test-emails directory", "error", err, "dir", emailDir)
		return err
	}

	// Save to test-emails.json (append mode)
	emailFile := emailDir + "/emails.json"
	file, err := os.OpenFile(emailFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		katapp.Logger(ctx).Error("failed to open test emails file", "error", err, "file", emailFile)
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	if err := encoder.Encode(mockEmail); err != nil {
		katapp.Logger(ctx).Error("failed to encode mock email", "error", err, "file", emailFile)
		return err
	}

	katapp.Logger(ctx).Info("mock email saved successfully", "to", to, "file", emailFile)
	return nil
}
