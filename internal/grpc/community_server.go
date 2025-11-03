package grpc

import (
	"context"

	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
	"gorm.io/gorm"

	"github.com/google/uuid"
	eventpkg "github.com/stormhead-org/backend/internal/event"
	middlewarepkg "github.com/stormhead-org/backend/internal/middleware"
	ormpkg "github.com/stormhead-org/backend/internal/orm"
	protopkg "github.com/stormhead-org/backend/internal/proto"
)

type CommunityServer struct {
	protopkg.UnimplementedCommunityServiceServer
	log      *zap.Logger
	database *ormpkg.PostgresClient
	broker   *eventpkg.KafkaClient
}

func NewCommunityServer(log *zap.Logger, database *ormpkg.PostgresClient, broker *eventpkg.KafkaClient) *CommunityServer {
	return &CommunityServer{
		log:      log,
		database: database,
		broker:   broker,
	}
}

func (s *CommunityServer) ValidateCommunitySlug(ctx context.Context, request *protopkg.ValidateCommunitySlugRequest) (*protopkg.ValidateCommunitySlugResponse, error) {
	_, err := s.database.SelectCommunityBySlug(
		request.Slug,
	)
	if err != gorm.ErrRecordNotFound {
		return nil, status.Errorf(codes.InvalidArgument, "slug already exist")
	}

	return &protopkg.ValidateCommunitySlugResponse{}, nil
}

func (s *CommunityServer) Create(ctx context.Context, request *protopkg.CreateCommunityRequest) (*protopkg.CreateCommunityResponse, error) {
	err := ValidateCommunitySlug(request.Slug)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "slug not match conditions")
	}

	err = ValidateCommunityName(request.Name)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "name not match conditions")
	}

	_, err = s.database.SelectCommunityBySlug(
		request.Slug,
	)
	if err != gorm.ErrRecordNotFound {
		return nil, status.Errorf(codes.InvalidArgument, "slug already exist")
	}

	userUUID, err := middlewarepkg.GetUserUUID(ctx)
	if err != nil {
		s.log.Error("internal error", zap.Error(err))
		return nil, status.Errorf(codes.Internal, "")
	}

	community := &ormpkg.Community{
		OwnerID:     userUUID,
		Slug:        request.Slug,
		Name:        request.Name,
		Description: request.Description,
		Rules:       request.Rules,
	}

	err = s.database.InsertCommunity(community)
	if err != nil {
		s.log.Error("internal error", zap.Error(err))
		return nil, status.Errorf(codes.Internal, "")
	}

	s.log.Info("community created",
		zap.String("id", community.ID.String()),
		zap.String("owner_id", community.OwnerID.String()),
		zap.String("name", community.Name),
	)

	return &protopkg.CreateCommunityResponse{
		Community: &protopkg.Community{
			Id:          community.ID.String(),
			OwnerId:     community.OwnerID.String(),
			Slug:        community.Slug,
			Name:        community.Name,
			Description: community.Description,
			Rules:       community.Rules,
			CreatedAt:   timestamppb.New(community.CreatedAt),
			UpdatedAt:   timestamppb.New(community.UpdatedAt),
		},
	}, nil
}

func (s *CommunityServer) Get(ctx context.Context, request *protopkg.GetCommunityRequest) (*protopkg.GetCommunityResponse, error) {
	community, err := s.database.SelectCommunityByID(request.CommunityId)
	if err == gorm.ErrRecordNotFound {
		s.log.Debug("community not found", zap.String("community_id", request.CommunityId))
		return nil, status.Errorf(codes.NotFound, "")
	}
	if err != nil {
		s.log.Error("internal error", zap.Error(err))
		return nil, status.Errorf(codes.Internal, "")
	}

	return &protopkg.GetCommunityResponse{
		Community: &protopkg.Community{
			Id:          community.ID.String(),
			OwnerId:     community.OwnerID.String(),
			Slug:        community.Slug,
			Name:        community.Name,
			Description: community.Description,
			Rules:       community.Rules,
			CreatedAt:   timestamppb.New(community.CreatedAt),
			UpdatedAt:   timestamppb.New(community.UpdatedAt),
		},
	}, nil
}

func (s *CommunityServer) Update(ctx context.Context, request *protopkg.UpdateCommunityRequest) (*protopkg.UpdateCommunityResponse, error) {
	community, err := s.database.SelectCommunityByID(request.CommunityId)
	if err == gorm.ErrRecordNotFound {
		s.log.Debug("community not found", zap.String("community_id", request.CommunityId))
		return nil, status.Errorf(codes.NotFound, "")
	}
	if err != nil {
		s.log.Error("internal error", zap.Error(err))
		return nil, status.Errorf(codes.Internal, "")
	}

	userUUID, err := middlewarepkg.GetUserUUID(ctx)
	if err != nil {
		s.log.Error("internal error", zap.Error(err))
		return nil, status.Errorf(codes.Internal, "")
	}

	if community.OwnerID != userUUID {
		s.log.Error("wrong community ownership")
		return nil, status.Errorf(codes.PermissionDenied, "not an owner")
	}

	community.Name = *request.Name
	community.Description = *request.Description
	community.Rules = *request.Rules

	err = s.database.UpdateCommunity(community)
	if err != nil {
		s.log.Error("internal error", zap.Error(err))
		return nil, status.Errorf(codes.Internal, "")
	}

	return &protopkg.UpdateCommunityResponse{
		Community: &protopkg.Community{
			Id:          community.ID.String(),
			OwnerId:     community.OwnerID.String(),
			Slug:        community.Slug,
			Name:        community.Name,
			Description: community.Description,
			Rules:       community.Rules,
			CreatedAt:   timestamppb.New(community.CreatedAt),
			UpdatedAt:   timestamppb.New(community.UpdatedAt),
		},
	}, nil
}

func (s *CommunityServer) Delete(ctx context.Context, request *protopkg.DeleteCommunityRequest) (*protopkg.DeleteCommunityResponse, error) {
	community, err := s.database.SelectCommunityByID(request.CommunityId)
	if err == gorm.ErrRecordNotFound {
		s.log.Debug("community not found", zap.String("community_id", request.CommunityId))
		return nil, status.Errorf(codes.NotFound, "")
	}
	if err != nil {
		s.log.Error("internal error", zap.Error(err))
		return nil, status.Errorf(codes.Internal, "")
	}

	userUUID, err := middlewarepkg.GetUserUUID(ctx)
	if err != nil {
		s.log.Error("internal error", zap.Error(err))
		return nil, status.Errorf(codes.Internal, "")
	}

	if community.OwnerID != userUUID {
		s.log.Error("wrong community ownership")
		return nil, status.Errorf(codes.PermissionDenied, "not an owner")
	}

	err = s.database.DeleteCommunity(community)
	if err != nil {
		s.log.Error("internal error", zap.Error(err))
		return nil, status.Errorf(codes.Internal, "")
	}

	return &protopkg.DeleteCommunityResponse{}, nil
}

func (s *CommunityServer) ListCommunities(ctx context.Context, request *protopkg.ListCommunitiesRequest) (*protopkg.ListCommunitiesResponse, error) {
	// Установка лимита по умолчанию
	limit := int(request.Limit)
	if limit <= 0 || limit > 40 {
		limit = 40
	}

	// Запрашиваем limit+1 элементов, чтобы определить, есть ли ещё данные
	communities, err := s.database.SelectCommunitiesWithPagination(limit+1, request.Cursor)
	if err != nil {
		s.log.Error("error fetching communities", zap.Error(err))
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
			Slug:        community.Slug,
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

	s.log.Debug("listed communities",
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

func (s *CommunityServer) Join(ctx context.Context, request *protopkg.JoinCommunityRequest) (*protopkg.JoinCommunityResponse, error) {
	community, err := s.database.SelectCommunityByID(request.CommunityId)
	if err == gorm.ErrRecordNotFound {
		s.log.Debug("community not found", zap.String("community_id", request.CommunityId))
		return nil, status.Errorf(codes.NotFound, "")
	}
	if err != nil {
		s.log.Error("internal error", zap.Error(err))
		return nil, status.Errorf(codes.Internal, "")
	}

	userUUID, err := middlewarepkg.GetUserUUID(ctx)
	if err != nil {
		s.log.Error("internal error", zap.Error(err))
		return nil, status.Errorf(codes.Internal, "")
	}

	_, err = s.database.SelectCommunityUser(
		community.ID.String(),
		userUUID.String(),
	)
	if err != gorm.ErrRecordNotFound {
		s.log.Debug(
			"user already in community",
			zap.String("community_id", community.ID.String()),
			zap.String("user_id", userUUID.String()),
		)
		return nil, status.Errorf(codes.InvalidArgument, "")
	}

	communityUser := ormpkg.CommunityUser{
		CommunityID: community.ID,
		UserID:      userUUID,
	}

	err = s.database.InsertCommunityUser(&communityUser)
	if err != nil {
		s.log.Error("internal error", zap.Error(err))
		return nil, status.Errorf(codes.Internal, "")
	}

	return &protopkg.JoinCommunityResponse{}, nil
}

func (s *CommunityServer) Leave(ctx context.Context, request *protopkg.LeaveCommunityRequest) (*protopkg.LeaveCommunityResponse, error) {
	community, err := s.database.SelectCommunityByID(request.CommunityId)
	if err == gorm.ErrRecordNotFound {
		s.log.Debug("community not found", zap.String("community_id", request.CommunityId))
		return nil, status.Errorf(codes.NotFound, "")
	}
	if err != nil {
		s.log.Error("internal error", zap.Error(err))
		return nil, status.Errorf(codes.Internal, "")
	}

	userUUID, err := middlewarepkg.GetUserUUID(ctx)
	if err != nil {
		s.log.Error("internal error", zap.Error(err))
		return nil, status.Errorf(codes.Internal, "")
	}

	communityUser, err := s.database.SelectCommunityUser(
		community.ID.String(),
		userUUID.String(),
	)
	if err == gorm.ErrRecordNotFound {
		s.log.Debug(
			"user not found in community",
			zap.String("community_id", community.ID.String()),
			zap.String("user_id", userUUID.String()),
		)
		return nil, status.Errorf(codes.InvalidArgument, "")
	}
	if err != nil {
		s.log.Error("internal error", zap.Error(err))
		return nil, status.Errorf(codes.Internal, "")
	}

	err = s.database.DeleteCommunityUser(communityUser)
	if err != nil {
		s.log.Error("internal error", zap.Error(err))
		return nil, status.Errorf(codes.Internal, "")
	}

	return &protopkg.LeaveCommunityResponse{}, nil
}

func (s *CommunityServer) Ban(ctx context.Context, request *protopkg.BanCommunityRequest) (*protopkg.BanCommunityResponse, error) {
	community, err := s.database.SelectCommunityByID(request.CommunityId)
	if err == gorm.ErrRecordNotFound {
		s.log.Debug("community not found", zap.String("community_id", request.CommunityId))
		return nil, status.Errorf(codes.NotFound, "")
	}
	if err != nil {
		s.log.Error("internal error", zap.Error(err))
		return nil, status.Errorf(codes.Internal, "")
	}

	community.IsBanned = true
	community.BanReason = request.Reason

	err = s.database.UpdateCommunity(community)
	if err != nil {
		s.log.Error("internal error", zap.Error(err))
		return nil, status.Errorf(codes.Internal, "")
	}

	return &protopkg.BanCommunityResponse{}, nil
}

func (s *CommunityServer) Unban(ctx context.Context, request *protopkg.UnbanCommunityRequest) (*protopkg.UnbanCommunityResponse, error) {
	community, err := s.database.SelectCommunityByID(request.CommunityId)
	if err == gorm.ErrRecordNotFound {
		s.log.Debug("community not found", zap.String("community_id", request.CommunityId))
		return nil, status.Errorf(codes.NotFound, "")
	}
	if err != nil {
		s.log.Error("internal error", zap.Error(err))
		return nil, status.Errorf(codes.Internal, "")
	}

	community.IsBanned = false
	community.BanReason = ""

	err = s.database.UpdateCommunity(community)
	if err != nil {
		s.log.Error("internal error", zap.Error(err))
		return nil, status.Errorf(codes.Internal, "")
	}

	return &protopkg.UnbanCommunityResponse{}, nil
}

func (s *CommunityServer) TransferOwnership(ctx context.Context, request *protopkg.TransferCommunityOwnershipRequest) (*protopkg.TransferCommunityOwnershipResponse, error) {
	community, err := s.database.SelectCommunityByID(request.CommunityId)
	if err == gorm.ErrRecordNotFound {
		s.log.Debug("community not found", zap.String("community_id", request.CommunityId))
		return nil, status.Errorf(codes.NotFound, "")
	}
	if err != nil {
		s.log.Error("internal error", zap.Error(err))
		return nil, status.Errorf(codes.Internal, "")
	}

	userUUID, err := middlewarepkg.GetUserUUID(ctx)
	if err != nil {
		s.log.Error("internal error", zap.Error(err))
		return nil, status.Errorf(codes.Internal, "")
	}

	if community.OwnerID != userUUID {
		s.log.Error("wrong community ownership")
		return nil, status.Errorf(codes.PermissionDenied, "not an owner")
	}

	newOwnerUUID, err := uuid.Parse(request.NewOwnerId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid new_owner_id")
	}

	community.OwnerID = newOwnerUUID

	err = s.database.UpdateCommunity(community)
	if err != nil {
		s.log.Error("internal error", zap.Error(err))
		return nil, status.Errorf(codes.Internal, "")
	}

	return &protopkg.TransferCommunityOwnershipResponse{}, nil
}
