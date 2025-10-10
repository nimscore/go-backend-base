package gql

import (
	"context"
	"fmt"
	"net/http"
	"os"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/handler/extension"
	"github.com/99designs/gqlgen/graphql/handler/lru"
	"github.com/99designs/gqlgen/graphql/handler/transport"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/prometheus/client_golang/prometheus"
	prometheushttp "github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/vektah/gqlparser/v2/ast"
	"go.uber.org/zap"

	clientpkg "github.com/stormhead-org/service/iam/internal/client"
	"github.com/stormhead-org/service/iam/internal/graph"
	mailpkg "github.com/stormhead-org/service/iam/internal/mail"
	metricpkg "github.com/stormhead-org/service/iam/internal/metric"
)

type API struct {
	logger   *zap.Logger
	host     string
	port     string
	registry *prometheus.Registry
}

func NewAPI(logger *zap.Logger, host string, port string) *API {
	gqlServer := handler.New(
		graph.NewExecutableSchema(
			graph.Config{
				Resolvers: &graph.Resolver{},
			},
		),
	)

	gqlServer.AddTransport(transport.Options{})
	gqlServer.AddTransport(transport.GET{})
	gqlServer.AddTransport(transport.POST{})

	gqlServer.SetQueryCache(lru.New[*ast.QueryDocument](1000))

	gqlServer.Use(extension.Introspection{})
	gqlServer.Use(extension.AutomaticPersistedQuery{
		Cache: lru.New[string](100),
	})

	this := &API{
		logger:   logger,
		host:     host,
		port:     port,
		registry: metricpkg.CreateRegistry(),
	}

	// GQL handlers
	if os.Getenv("DEBUG") == "1" {
		http.Handle("/", playground.Handler("GraphQL playground", "/query"))
	}
	http.Handle("/query", gqlServer)

	// Prometheus metrics
	http.Handle("/metrics", prometheushttp.HandlerFor(this.registry, prometheushttp.HandlerOpts{Registry: this.registry}))

	// Kubernetes probes
	http.Handle("/probe/startup", this.startupHandler())
	http.Handle("/probe/readiness", this.readinessHandler())
	http.Handle("/probe/liveness", this.livenessHandler())
	http.Handle("/debug", this.debugHandler())

	return this
}

func (this *API) Start() error {
	this.logger.Info("start")

	go func() {
		http.ListenAndServe(
			fmt.Sprintf("%s:%s", this.host, this.port),
			nil,
		)
	}()

	return nil
}

func (this *API) Stop() error {
	this.logger.Info("stop")
	return nil
}

func (this *API) startupHandler() http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		writer.WriteHeader(http.StatusOK)
	}
}

func (this *API) readinessHandler() http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		writer.WriteHeader(http.StatusOK)
	}
}

func (this *API) livenessHandler() http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		writer.WriteHeader(http.StatusOK)
	}
}

func (this *API) debugHandler() http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		kafkaHost := os.Getenv("KAFKA_HOST")
		if kafkaHost == "" {
			kafkaHost = "localhost"
		}

		kafkaPort := os.Getenv("KAFKA_PORT")
		if kafkaPort == "" {
			kafkaPort = "9092"
		}

		kafkaTopic := os.Getenv("KAFKA_TOPIC")
		if kafkaTopic == "" {
			kafkaTopic = "mail"
		}

		this.logger.Error(fmt.Sprintf("%s:%s", kafkaHost, kafkaPort))

		client := clientpkg.NewKafkaClient(
			fmt.Sprintf("%s:%s", kafkaHost, kafkaPort),
			kafkaTopic,
		)

		content, err := mailpkg.MessageToJson(
			mailpkg.NewMessageMailConfirm("staff@stormhead.org", "dgemojkod@yandex.ru", "Confirm", "User", "24", "https://stormhead.org"),
		)
		if err != nil {
			this.logger.Error("error", zap.Error(err))
		}

		client.Write(context.Background(), mailpkg.KIND_MAIL_CONFIRM, content)
	}
}
