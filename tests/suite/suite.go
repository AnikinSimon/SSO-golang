package suite

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net"
	"os"
	"sso/internal/config"
	"strconv"
	"testing"

	ssov1 "github.com/AnikinSimon/sso-protos/gen/go/sso"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

const (
	grpcHost = "localhost"
)

type Suite struct {
	*testing.T
	Cfg        *config.Config
	AuthClient ssov1.AuthClient
}

func New(t *testing.T) (context.Context, *Suite) {
	t.Helper()
	t.Parallel()

	cfg := config.MustLoadByPath("../../config/local.yaml")

	ctx, cancel := context.WithTimeout(context.Background(), cfg.GRPC.Timeout)

	tlsCredentials, err := loadTLSCredentials()

	if err != nil {
		t.Fatalf("creating tls failed: %v", err)
	}

	t.Cleanup(func() {
		t.Helper()
		cancel()
	})

	grpcClient, err := grpc.NewClient(grpcAddress(cfg), grpc.WithTransportCredentials(tlsCredentials))

	if err != nil {
		t.Fatalf("grpc server connection failed: %v", err)
	}

	return ctx, &Suite{
		T:          t,
		Cfg:        cfg,
		AuthClient: ssov1.NewAuthClient(grpcClient),
	}
}

func loadTLSCredentials() (credentials.TransportCredentials, error) {
	// Загружаем сертификат CA, подписавшего сертификат сервера
	pemServerCA, err := os.ReadFile("../../cert/ca-cert.pem")
	if err != nil {
		return nil, err
	}

	certPool := x509.NewCertPool()
	if !certPool.AppendCertsFromPEM(pemServerCA) {
		return nil, fmt.Errorf("failed to add server CA's certificate")
	}

	// Создаём идентификационные данные и возвращаем их
	config := &tls.Config{
		RootCAs: certPool,
	}

	return credentials.NewTLS(config), nil
}

func grpcAddress(cfg *config.Config) string {
	return net.JoinHostPort(grpcHost, strconv.Itoa(cfg.GRPC.Port))
}
