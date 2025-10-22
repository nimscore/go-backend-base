package worker

import (
	"context"
	"sync"
	"time"

	"go.uber.org/zap"

	clientpkg "github.com/stormhead-org/backend/internal/client"
	eventpkg "github.com/stormhead-org/backend/internal/event"
)

type Worker struct {
	context      context.Context
	cancel       func()
	waitGroup    sync.WaitGroup
	logger       *zap.Logger
	router       *Router
	brokerClient *eventpkg.KafkaClient
	mailClient   *clientpkg.MailClient
}

func NewWorker(logger *zap.Logger, brokerClient *eventpkg.KafkaClient, mailClient *clientpkg.MailClient) *Worker {
	context, cancel := context.WithCancel(context.Background())
	this := &Worker{
		context:      context,
		cancel:       cancel,
		logger:       logger,
		brokerClient: brokerClient,
		mailClient:   mailClient,
	}
	this.router = NewRouter(
		map[string][]EventHandler{
			eventpkg.AUTHORIZATION_LOGIN: {
				this.AuthorizationLoginHandler,
			},
		},
	)
	return this
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
		select {
		case <-this.context.Done():
			return
		case <-time.After(1 * time.Millisecond):
		}

		event, data, err := this.brokerClient.ReadMessage(this.context)
		if err != nil {
			this.logger.Error("error receiving kafka message", zap.Error(err))
			continue
		}

		err = this.router.Handle(event, []byte(data))
		if err != nil {
			this.logger.Error("error handling kafka message", zap.Error(err))
			continue
		}
	}
}
