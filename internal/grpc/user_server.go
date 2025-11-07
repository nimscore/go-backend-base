package grpc

import (
	"context"

	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"

	eventpkg "github.com/stormhead-org/backend/internal/event"
	ormpkg "github.com/stormhead-org/backend/internal/orm"
	protopkg "github.com/stormhead-org/backend/internal/proto"
)

type UserServer struct {
	protopkg.UnimplementedUserServiceServer
	log      *zap.Logger
	database *ormpkg.PostgresClient
	broker   *eventpkg.KafkaClient
}

func NewUserServer(log *zap.Logger, database *ormpkg.PostgresClient, broker *eventpkg.KafkaClient) *UserServer {
	return &UserServer{
		log:      log,
		database: database,
		broker:   broker,
	}
}

func (s *UserServer) Get(ctx context.Context, request *protopkg.GetUserRequest) (*protopkg.GetUserResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Get not implemented")
}

func (s *UserServer) GetCurrent(ctx context.Context, request *protopkg.GetCurrentUserRequest) (*protopkg.GetCurrentUserResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetCurrent not implemented")
}

func (s *UserServer) UpdateProfile(ctx context.Context, request *protopkg.UpdateProfileRequest) (*protopkg.UpdateProfileResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method UpdateProfile not implemented")
}

func (s *UserServer) GetStatistics(ctx context.Context, request *protopkg.GetUserStatisticsRequest) (*protopkg.GetUserStatisticsResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetStatistics not implemented")
}

func (s *UserServer) ListCommunities(ctx context.Context, request *protopkg.ListUserCommunitiesRequest) (*protopkg.ListUserCommunitiesResponse, error) {
	limit := int(request.Limit)
	if limit <= 0 || limit > 50 {
		limit = 50
	}

	communities, err := s.database.SelectCommunitiesWithPagination(request.UserId, limit+1, request.Cursor)
	if err != nil {
		s.log.Error("internal error", zap.Error(err))
		return nil, status.Errorf(codes.Internal, "")
	}

	hasMore := len(communities) > limit
	if hasMore {
		communities = communities[:limit]
	}

	var nextCursor string
	if hasMore && len(communities) > 0 {
		nextCursor = communities[len(communities)-1].ID.String()
	}

	result := make([]*protopkg.Community, len(communities))
	for i, community := range communities {
		result[i] = &protopkg.Community{
			Id:          community.ID.String(),
			OwnerId:     community.OwnerID.String(),
			Slug:        community.Slug,
			Name:        community.Name,
			Description: community.Description,
			CreatedAt:   timestamppb.New(community.CreatedAt),
			UpdatedAt:   timestamppb.New(community.UpdatedAt),
		}
	}

	return &protopkg.ListUserCommunitiesResponse{
		Communities: result,
		NextCursor:  nextCursor,
		HasMore:     hasMore,
	}, nil
}

func (s *UserServer) ListPosts(ctx context.Context, request *protopkg.ListUserPostsRequest) (*protopkg.ListUserPostsResponse, error) {
	limit := int(request.Limit)
	if limit <= 0 || limit > 50 {
		limit = 50
	}

	posts, err := s.database.SelectPostsWithPagination(request.UserId, limit+1, request.Cursor)
	if err != nil {
		s.log.Error("internal error", zap.Error(err))
		return nil, status.Errorf(codes.Internal, "")
	}

	hasMore := len(posts) > limit
	if hasMore {
		posts = posts[:limit]
	}

	var nextCursor string
	if hasMore && len(posts) > 0 {
		nextCursor = posts[len(posts)-1].ID.String()
	}

	result := make([]*protopkg.Post, len(posts))
	for i, post := range posts {
		result[i] = &protopkg.Post{
			Id:            post.ID.String(),
			CommunityId:   post.CommunityID.String(),
			CommunityName: post.Community.Name,
			AuthorId:      post.AuthorID.String(),
			AuthorName:    post.Author.Name,
			Title:         post.Title,
			Content:       post.Content,
			Status:        protopkg.PostStatus(post.Status),
			CreatedAt:     timestamppb.New(post.CreatedAt),
			UpdatedAt:     timestamppb.New(post.UpdatedAt),
			PublishedAt:   timestamppb.New(post.PublishedAt),
		}
	}

	return &protopkg.ListUserPostsResponse{
		Posts:      result,
		NextCursor: nextCursor,
		HasMore:    hasMore,
	}, nil
}

func (s *UserServer) ListComments(ctx context.Context, request *protopkg.ListUserCommentsRequest) (*protopkg.ListUserCommentsResponse, error) {
	limit := int(request.Limit)
	if limit <= 0 || limit > 50 {
		limit = 50
	}

	comments, err := s.database.SelectCommentsWithPagination("", request.UserId, limit+1, request.Cursor)
	if err != nil {
		s.log.Error("internal error", zap.Error(err))
		return nil, status.Errorf(codes.Internal, "")
	}

	hasMore := len(comments) > limit
	if hasMore {
		comments = comments[:limit]
	}

	var nextCursor string
	if hasMore && len(comments) > 0 {
		nextCursor = comments[len(comments)-1].ID.String()
	}

	result := make([]*protopkg.CommentWithPostInfo, len(comments))
	for i, comment := range comments {
		parentCommentID := ""
		if comment.ParentCommentID != nil {
			parentCommentID = comment.ParentCommentID.String()
		}

		result[i] = &protopkg.CommentWithPostInfo{
			PostId:    comment.Post.ID.String(),
			PostTitle: comment.Post.Title,
			Comment: &protopkg.Comment{
				Id:              comment.ID.String(),
				ParentCommentId: parentCommentID,
				PostId:          comment.PostID.String(),
				AuthorId:        comment.AuthorID.String(),
				AuthorName:      comment.Author.Name,
				Content:         comment.Content,
				CreatedAt:       timestamppb.New(comment.CreatedAt),
				UpdatedAt:       timestamppb.New(comment.UpdatedAt),
			},
		}
	}

	return &protopkg.ListUserCommentsResponse{
		Comments:   result,
		NextCursor: nextCursor,
		HasMore:    hasMore,
	}, nil
}

func (s *UserServer) Follow(ctx context.Context, request *protopkg.FollowRequest) (*protopkg.FollowResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Follow not implemented")
}

func (s *UserServer) Unfollow(ctx context.Context, request *protopkg.UnfollowRequest) (*protopkg.UnfollowResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Unfollow not implemented")
}

func (s *UserServer) ListFollowers(ctx context.Context, request *protopkg.ListFollowersRequest) (*protopkg.ListFollowersResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ListFollowers not implemented")
}

func (s *UserServer) ListFollowing(ctx context.Context, request *protopkg.ListFollowingRequest) (*protopkg.ListFollowingResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ListFollowing not implemented")
}

func (s *UserServer) Heartbeat(ctx context.Context, request *protopkg.HeartbeatRequest) (*protopkg.HeartbeatResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Heartbeat not implemented")
}
