package worker

import (
	"context"
	"fmt"
	"sync"

	"go.uber.org/zap"

	clientpkg "github.com/stormhead-org/worker/mail/internal/client"
	mailpkg "github.com/stormhead-org/worker/mail/internal/mail"
	templatepkg "github.com/stormhead-org/worker/mail/internal/template"
)

type Worker struct {
	context     context.Context
	cancel      func()
	waitGroup   sync.WaitGroup
	logger      *zap.Logger
	kafkaClient *clientpkg.KafkaClient
	mailClient  *clientpkg.MailClient
}

func NewWorker(logger *zap.Logger, kafkaClient *clientpkg.KafkaClient, mailClient *clientpkg.MailClient) *Worker {
	context, cancel := context.WithCancel(context.Background())
	return &Worker{
		context:     context,
		cancel:      cancel,
		logger:      logger,
		kafkaClient: kafkaClient,
		mailClient:  mailClient,
	}
}

func (this *Worker) Start() error {
	this.logger.Info("starting mail worker")

	this.waitGroup.Add(1)
	go this.worker()
	return nil
}

func (this *Worker) Stop() error {
	this.logger.Info("stopping mail worker")

	this.cancel()
	this.waitGroup.Wait()
	return nil
}

func (this *Worker) worker() {
	defer this.waitGroup.Done()

	for {
		_, value, err := this.kafkaClient.Read(this.context)
		if err != nil {
			this.logger.Error("error receiving kafka message", zap.Error(err))
			continue
		}

		message, err := mailpkg.MessageFromJson(value)
		if err != nil {
			this.logger.Error("error unmarshalling message", zap.Error(err))
			continue
		}

		path := "mail_stub.html"
		switch message.Kind {
		case mailpkg.KIND_MAIL_CONFIRM:
			path = "mail_confirm.html"
		case mailpkg.KIND_MAIL_RECOVER:
			path = "mail_recover.html"
		}

		content, err := templatepkg.Render(fmt.Sprintf("template/%s", path), message.Arguments)
		if err != nil {
			this.logger.Error("error rendering mail", zap.Error(err))
			continue
		}

		err = this.mailClient.SendHTML(
			message.From,
			message.To,
			message.Subject,
			content,
		)
		if err != nil {
			this.logger.Error("error sending mail", zap.Error(err))
			continue
		}

		this.logger.Info(
			"mail sent to recipient",
			zap.String("kind", message.Kind),
			zap.String("from", message.From),
			zap.String("to", message.To),
		)
	}
}
