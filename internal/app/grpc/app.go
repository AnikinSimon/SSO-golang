package grpcapp

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os"
	authgrpc "sso/internal/grpc/auth"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/reflection"
)

type App struct {
	log         *slog.Logger
	gRPCServer  *grpc.Server
	gRPCGateway *runtime.ServeMux
	portRPC     int
	portGateway int
}

var (
	ErrAddClientToCA = errors.New("failed to add client CA's certificate")
)

const (
	caCertFile     = "cert/ca-cert.pem"
	serverCertFile = "cert/server-cert.pem"
	serverKeyFile  = "cert/server-key.pem"
)

func New(
	log *slog.Logger,
	authService authgrpc.Auth,
	portRPC int,
	portGateway int,
) (*App, error) {

	creds, err := loadTLSCredentials()

	gRPCServer := grpc.NewServer(
		grpc.Creds(creds),
	)

	reflection.Register(gRPCServer)

	gRPCGateway := runtime.NewServeMux()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	authgrpc.Register(ctx, gRPCGateway, gRPCServer, authService)

	return &App{
		log:         log,
		gRPCServer:  gRPCServer,
		gRPCGateway: gRPCGateway,
		portRPC:     portRPC,
		portGateway: portGateway,
	}, err
}

func (a *App) MustRunRPC() {
	if err := a.RunRPC(); err != nil {
		panic(err)
	}
}

func (a *App) RunRPC() error {
	const op = "grpcapp.RunRPC"

	log := a.log.With(
		slog.String("op", op),
		slog.Int("port", a.portRPC),
	)

	l, err := net.Listen("tcp", fmt.Sprintf(":%d", a.portRPC))

	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	log.Info("gRPC server is running", slog.String("addr", l.Addr().String()))

	if err := a.gRPCServer.Serve(l); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (a *App) MustRunGateway() {
	if err := a.RunGateway(); err != nil {
		panic(err)
	}
}

func (a *App) RunGateway() error {
	const op = "grpcapp.RunGateway"

	log := a.log.With(
		slog.String("op", op),
		slog.Int("port", a.portGateway),
	)

	l, err := net.Listen("tcp", fmt.Sprintf(":%d", a.portGateway))

	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	log.Info("gRPC-gateway is running", slog.String("addr", l.Addr().String()))

	if err := http.ServeTLS(l, a.gRPCGateway, serverCertFile, serverKeyFile); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func loadTLSCredentials() (credentials.TransportCredentials, error) {
	const op = "app.grpc.loadTLSCredentials"

	pemClientCA, err := os.ReadFile(caCertFile)
	if err != nil {
		return nil, fmt.Errorf("%s %w", op, err)
	}

	certPool := x509.NewCertPool()
	if !certPool.AppendCertsFromPEM(pemClientCA) {
		return nil, fmt.Errorf("%s %w", op, ErrAddClientToCA)
	}

	serverCert, err := tls.LoadX509KeyPair(serverCertFile, serverKeyFile)
	if err != nil {
		return nil, fmt.Errorf("%s %w", op, err)
	}

	config := &tls.Config{
		Certificates: []tls.Certificate{serverCert},
		ClientAuth:   tls.RequireAndVerifyClientCert,
		ClientCAs:    certPool,
	}

	return credentials.NewTLS(config), nil
}

func (a *App) Stop() {
	const op = "grpcapp.Stop"

	a.log.With(slog.String("op", op)).Info("Stopping gRPC server", slog.Int("port", a.portRPC))

	a.gRPCServer.GracefulStop()
}
