package grpc

import (
	"context"

	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
	"gorm.io/gorm"

	eventpkg "github.com/stormhead-org/backend/internal/event"
	ormpkg "github.com/stormhead-org/backend/internal/orm"
	protopkg "github.com/stormhead-org/backend/internal/proto"
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

func (this *CommunityServer) Create(ctx context.Context, request *protopkg.CreateCommunityRequest) (*protopkg.CreateCommunityResponse, error) {
	// // Валидация входных данных
	// if request.Slug == "" || request.Name == "" || request.OwnerId == "" {
	// 	return nil, status.Errorf(codes.InvalidArgument, "slug, name and owner_id are required")
	// }

	// // Проверка уникальности slug
	// _, err := this.databaseClient.SelectCommunityBySlug(request.Slug)
	// if err != gorm.ErrRecordNotFound {
	// 	return nil, status.Errorf(codes.AlreadyExists, "community with slug %s already exists", request.Slug)
	// }

	// // Парсинг owner_id
	// ownerUUID, err := uuid.Parse(request.OwnerId)
	// if err != nil {
	// 	this.logger.Error("invalid owner_id format", zap.Error(err), zap.String("owner_id", request.OwnerId))
	// 	return nil, status.Errorf(codes.InvalidArgument, "invalid owner_id format")
	// }

	// // Создание сообщества
	// community := &ormpkg.Community{
	// 	OwnerID:     ownerUUID,
	// 	Slug:        request.Slug,
	// 	Name:        request.Name,
	// 	Description: request.Description,
	// }

	// err = this.databaseClient.InsertCommunity(community)
	// if err != nil {
	// 	this.logger.Error("error inserting community", zap.Error(err))
	// 	return nil, status.Errorf(codes.Internal, "failed to create community")
	// }

	// this.logger.Info("community created",
	// 	zap.String("id", community.ID.String()),
	// 	zap.String("slug", community.Slug),
	// 	zap.String("owner_id", request.OwnerId),
	// )

	return &protopkg.CreateCommunityResponse{}, nil
}

func (this *CommunityServer) Get(ctx context.Context, request *protopkg.GetCommunityRequest) (*protopkg.GetCommunityResponse, error) {
	var community *ormpkg.Community
	var err error

	// Получение сообщества по ID или slug
	// switch identifier := request.Identifier.(type) {
	// case *protopkg.GetCommunityRequest_Id:
	// 	community, err = this.databaseClient.SelectCommunityByID(identifier.Id)
	// case *protopkg.GetCommunityRequest_Slug:
	// 	community, err = this.databaseClient.SelectCommunityBySlug(identifier.Slug)
	// default:
	// 	return nil, status.Errorf(codes.InvalidArgument, "either id or slug must be provided")
	// }

	if err == gorm.ErrRecordNotFound {
		return nil, status.Errorf(codes.NotFound, "community not found")
	}
	if err != nil {
		this.logger.Error("error fetching community", zap.Error(err))
		return nil, status.Errorf(codes.Internal, "failed to get community")
	}

	return &protopkg.GetCommunityResponse{
		Community: &protopkg.Community{
			Id:          community.ID.String(),
			OwnerId:     community.OwnerID.String(),
			Name:        community.Name,
			Description: community.Description,
			CreatedAt:   timestamppb.New(community.CreatedAt),
			UpdatedAt:   timestamppb.New(community.UpdatedAt),
		},
	}, nil
}

func (this *CommunityServer) Update(context.Context, *protopkg.UpdateCommunityRequest) (*protopkg.UpdateCommunityResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Update not implemented")
}

func (this *CommunityServer) Delete(context.Context, *protopkg.DeleteCommunityRequest) (*protopkg.DeleteCommunityResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Delete not implemented")
}

func (this *CommunityServer) ListCommunities(ctx context.Context, request *protopkg.ListCommunitiesRequest) (*protopkg.ListCommunitiesResponse, error) {
	// Установка лимита по умолчанию
	limit := int(request.Limit)
	if limit <= 0 || limit > 40 {
		limit = 40
	}

	// Запрашиваем limit+1 элементов, чтобы определить, есть ли ещё данные
	communities, err := this.databaseClient.SelectCommunitiesWithPagination(limit+1, request.Cursor)
	if err != nil {
		this.logger.Error("error fetching communities", zap.Error(err))
		return nil, status.Errorf(codes.Internal, "failed to list communities")
	}

	// Определяем, есть ли ещё элементы
	hasMore := len(communities) > limit
	if hasMore {
		communities = communities[:limit]
	}

	// Формирование ответа
	protoCommunities := make([]*protopkg.Community, len(communities))
	for i, community := range communities {
		protoCommunities[i] = &protopkg.Community{
			Id:          community.ID.String(),
			OwnerId:     community.OwnerID.String(),
			Name:        community.Name,
			Description: community.Description,
			CreatedAt:   timestamppb.New(community.CreatedAt),
			UpdatedAt:   timestamppb.New(community.UpdatedAt),
		}
	}

	// Определение next_cursor
	var nextCursor string
	if hasMore && len(communities) > 0 {
		nextCursor = communities[len(communities)-1].ID.String()
	}

	this.logger.Debug("listed communities",
		zap.Int("count", len(protoCommunities)),
		zap.Bool("has_more", hasMore),
		zap.String("cursor", request.Cursor),
		zap.String("next_cursor", nextCursor),
	)

	return &protopkg.ListCommunitiesResponse{
		Communities: protoCommunities,
		NextCursor:  nextCursor,
		HasMore:     hasMore,
	}, nil
}

func (this *CommunityServer) Join(context.Context, *protopkg.JoinCommunityRequest) (*protopkg.JoinCommunityResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Join not implemented")
}

func (this *CommunityServer) Leave(context.Context, *protopkg.LeaveCommunityRequest) (*protopkg.LeaveCommunityResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Leave not implemented")
}

func (this *CommunityServer) Ban(context.Context, *protopkg.BanCommunityRequest) (*protopkg.BanCommunityResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Ban not implemented")
}

func (this *CommunityServer) Unban(context.Context, *protopkg.UnbanCommunityRequest) (*protopkg.UnbanCommunityResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Unban not implemented")
}

func (this *CommunityServer) TransferOwnership(context.Context, *protopkg.TransferCommunityOwnershipRequest) (*protopkg.TransferCommunityOwnershipResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method TransferOwnership not implemented")
}
