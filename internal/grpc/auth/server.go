package authgrpc

import (
	"context"
	"errors"
	"sso/internal/services/auth"
	"sso/internal/storage"
	ssov1 "sso/streaming/go/sso"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Auth interface {
	Login(
		ctx context.Context,
		email string,
		password string,
		appID string,
	) (token string, err error)

	RegisterNewUser(
		ctx context.Context,
		email string,
		password string,
		app_id string,
	) (userUUID string, err error)

	IsAdmin(ctx context.Context, userID string) (isAdmin bool, err error)

	RegisterNewApp(ctx context.Context, appID string, appSecret string) (appUUID string, err error)
}

type serverAPI struct {
	ssov1.UnimplementedAuthServer
	auth Auth
}

func Register(ctx context.Context, router *runtime.ServeMux, gRPC *grpc.Server, auth Auth) {
	serveApi := &serverAPI{auth: auth}

	ssov1.RegisterAuthServer(gRPC, serveApi)
	err := ssov1.RegisterAuthHandlerServer(ctx, router, serveApi)
	if err != nil {
		panic(err)
	}
}

//func RegisterGateway(ctx context.Context, router *runtime.ServeMux, auth Auth) {
//	err := ssov1.RegisterAuthHandlerServer(ctx, router, &serverAPI{auth: auth})
//	if err != nil {
//		panic(err)
//	}
//}

func (s *serverAPI) Login(
	ctx context.Context,
	req *ssov1.LoginRequest,
) (*ssov1.LoginResponse, error) {

	err := validateLogin(req)

	if err != nil {
		return nil, err
	}

	token, err := s.auth.Login(ctx, req.GetEmail(), req.GetPassword(), req.GetAppUuid())

	if err != nil {
		if errors.Is(err, auth.ErrInvalidCredentials) {
			return nil, status.Error(codes.InvalidArgument, "invalid argument")
		}
		return nil, status.Error(codes.Internal, "internal error")
	}

	return &ssov1.LoginResponse{
		Token: token,
	}, nil
}

func (s *serverAPI) Register(
	ctx context.Context,
	req *ssov1.RegisterRequest,
) (*ssov1.RegisterResponse, error) {

	err := validateRegister(req)

	if err != nil {
		return nil, err
	}

	userUUID, err := s.auth.RegisterNewUser(ctx, req.GetEmail(), req.GetPassword(), req.GetAppUuid())
	if err != nil {
		if errors.Is(err, auth.ErrUserExists) {
			return nil, status.Error(codes.AlreadyExists, "user already exists")
		}
		return nil, status.Error(codes.Internal, "internal error")
	}

	return &ssov1.RegisterResponse{
		UserUuid: userUUID,
	}, nil
}

func (s *serverAPI) IsAdmin(
	ctx context.Context,
	req *ssov1.IsAdminRequest,
) (*ssov1.IsAdminResponse, error) {

	err := validateIsAdmin(req)

	if err != nil {
		return nil, err
	}

	isAdmin, err := s.auth.IsAdmin(ctx, req.GetUserUuid())
	if err != nil {
		if errors.Is(err, storage.ErrUserNotFound) {
			return nil, status.Error(codes.NotFound, "user not found")
		}
		return nil, status.Error(codes.Internal, "internal error")
	}
	return &ssov1.IsAdminResponse{
		IsAdmin: isAdmin,
	}, nil
}

func (s *serverAPI) RegisterApp(
	ctx context.Context,
	req *ssov1.RegisterAppRequest,
) (*ssov1.RegisterAppResponse, error) {

	err := validateRegisterApp(req)

	if err != nil {
		return nil, err
	}

	appID, err := s.auth.RegisterNewApp(ctx, req.GetName(), req.GetSecret())
	if err != nil {
		if errors.Is(err, auth.ErrAppExists) {
			return nil, status.Error(codes.AlreadyExists, "app already exists")
		}
		return nil, status.Error(codes.Internal, "internal error")
	}
	return &ssov1.RegisterAppResponse{
		AppUuid: appID,
	}, nil
}

func validateLogin(req *ssov1.LoginRequest) error {
	if req.GetEmail() == "" {
		return status.Error(codes.InvalidArgument, "email is required")
	}

	if req.GetPassword() == "" {
		return status.Error(codes.InvalidArgument, "password is required")
	}

	if req.GetAppUuid() == "" {
		return status.Error(codes.InvalidArgument, "app_uuid is required")
	}

	return nil
}

func validateRegister(req *ssov1.RegisterRequest) error {
	if req.GetEmail() == "" {
		return status.Error(codes.InvalidArgument, "email is required")
	}

	if req.GetPassword() == "" {
		return status.Error(codes.InvalidArgument, "password is required")
	}

	if req.GetAppUuid() == "" {
		return status.Error(codes.InvalidArgument, "app_id is required")
	}

	return nil
}

func validateIsAdmin(req *ssov1.IsAdminRequest) error {
	if req.GetUserUuid() == "" {
		return status.Error(codes.InvalidArgument, "user_uuid is required")
	}

	return nil
}

func validateRegisterApp(req *ssov1.RegisterAppRequest) error {
	if req.GetName() == "" {
		return status.Error(codes.InvalidArgument, "name is required")
	}

	if req.GetSecret() == "" {
		return status.Error(codes.InvalidArgument, "secret is required")
	}

	return nil
}
