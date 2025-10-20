package grpc

import (
	"context"

	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	eventpkg "github.com/stormhead-org/service/community/internal/event"
	ormpkg "github.com/stormhead-org/service/community/internal/orm"
	protopkg "github.com/stormhead-org/service/community/internal/proto"
)

type CommunityServer struct {
	protopkg.UnimplementedCommunityServiceServer
	logger         *zap.Logger
	databaseClient *ormpkg.PostgresClient
	brokerClient   *eventpkg.KafkaClient
}

func NewCommunityServer(logger *zap.Logger, databaseClient *ormpkg.PostgresClient, brokerClient *eventpkg.KafkaClient) *CommunityServer {
	return &CommunityServer{
		logger:         logger,
		databaseClient: databaseClient,
		brokerClient:   brokerClient,
	}
}

func (this *CommunityServer) Create(context context.Context, request *protopkg.CreateRequest) (*protopkg.CreateResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Create not implemented")
}
