package grpc

import (
	"context"
	"fmt"

	ormpkg "github.com/stormhead-org/service/community/internal/orm"
	iampb "github.com/stormhead-org/service/community/internal/proto"
)

type IAMServer struct {
	iampb.UnimplementedIAMServiceServer
}

func NewIAMServer() *IAMServer {
	return &IAMServer{}
}

func (this *IAMServer) Register(context context.Context, request *iampb.RegisterRequest) (*iampb.RegisterResponse, error) {
	return nil, nil
}

func (this *IAMServer) Login(context context.Context, request *iampb.LoginRequest) (*iampb.LoginResponse, error) {
	ormpkg.Debug()
	fmt.Println(request.Login, request.Password)
	return &iampb.LoginResponse{Token: "hello"}, nil
}

func (this *IAMServer) Validate(context context.Context, request *iampb.ValidateRequest) (*iampb.ValidateResponse, error) {
	return nil, nil
}

func (this *IAMServer) Refresh(context context.Context, request *iampb.RefreshRequest) (*iampb.RefreshResponse, error) {
	return nil, nil
}
