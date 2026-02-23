package email

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
)

type Message struct {
	To      string
	From    string
	Subject string
	Text    string
	HTML    string
}

type Sender interface {
	Send(ctx context.Context, msg Message) error
}

type noopSender struct{}

func NewNoop() Sender { return &noopSender{} }

func (s *noopSender) Send(context.Context, Message) error { return nil }

type ConsoleSender struct{}

func NewConsole() Sender { return &ConsoleSender{} }

func (s *ConsoleSender) Send(_ context.Context, msg Message) error {
	slog.Info("email send", "to", msg.To, "from", msg.From, "subject", msg.Subject, "text_len", len(msg.Text), "html_len", len(msg.HTML))
	return nil
}

func ValidateMessage(msg Message) error {
	if strings.TrimSpace(msg.To) == "" {
		return fmt.Errorf("missing To")
	}
	if strings.TrimSpace(msg.From) == "" {
		return fmt.Errorf("missing From")
	}
	if strings.TrimSpace(msg.Subject) == "" {
		return fmt.Errorf("missing Subject")
	}
	if strings.TrimSpace(msg.Text) == "" && strings.TrimSpace(msg.HTML) == "" {
		return fmt.Errorf("missing body")
	}
	return nil
}
