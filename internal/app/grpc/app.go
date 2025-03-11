package grpcapp

import (
	"crypto/tls"
	"errors"
	"fmt"
	"log/slog"
	"net"
	authgrpc "sso/internal/grpc/auth"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

type App struct {
	log        *slog.Logger
	gRPCServer *grpc.Server
	port       int
}

var (
	ErrAddClientToCA = errors.New("failed to add client CA's certificate")
)

func New(
	log *slog.Logger,
	authService authgrpc.Auth,
	port int,
) (*App, error) {

	creds, err := loadTLSCredentials()

	gRPCServer := grpc.NewServer(
		grpc.Creds(creds),
	)

	authgrpc.Register(gRPCServer, authService)

	return &App{
		log:        log,
		gRPCServer: gRPCServer,
		port:       port,
	}, err
}

func (a *App) MustRun() {
	if err := a.Run(); err != nil {
		panic(err)
	}
}

func (a *App) Run() error {
	const op = "grpcapp.Run"

	log := a.log.With(
		slog.String("op", op),
		slog.Int("port", a.port),
	)

	l, err := net.Listen("tcp", fmt.Sprintf(":%d", a.port))

	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	log.Info("gRPC server is running", slog.String("addr", l.Addr().String()))

	if err := a.gRPCServer.Serve(l); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func loadTLSCredentials() (credentials.TransportCredentials, error) {
	const op = "app.grpc.loadTLSCredentials"

	serverCert, err := tls.LoadX509KeyPair("cert/server-cert.pem", "cert/server-key.pem")
	if err != nil {
		return nil, fmt.Errorf("%s %w", op, err)
	}

	config := &tls.Config{
		Certificates: []tls.Certificate{serverCert},
		ClientAuth:   tls.NoClientCert,
	}

	return credentials.NewTLS(config), nil
}

func (a *App) Stop() {
	const op = "grpcapp.Stop"

	a.log.With(slog.String("op", op)).Info("Stopping gRPC server", slog.Int("port", a.port))

	a.gRPCServer.GracefulStop()
}
