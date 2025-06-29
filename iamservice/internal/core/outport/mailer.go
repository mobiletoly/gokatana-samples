package outport

import "context"

//go:generate go tool gobetter -input $GOFILE

type MailContent struct { //+gob:Constructor
	ContentType string
	Title       string
	Body        string
}

type Mailer interface {
	SendEmail(ctx context.Context, to string, content *MailContent) error
}
