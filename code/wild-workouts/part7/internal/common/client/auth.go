package client

import (
	"context"
	"fmt"
	"strings"

	"cloud.google.com/go/compute/metadata"
	"github.com/pkg/errors"
	"google.golang.org/grpc/credentials"
)

type metadataServerToken struct {
	serviceURL string
}

func newMetadataServerToken(grpcAddr string) credentials.PerRPCCredentials {
	// основываясь на https://cloud.google.com/run/docs/authenticating/service-to-service#go
	// сервис должен иметь префикс https без порта
	serviceURL := "https://" + strings.Split(grpcAddr, ":")[0]

	return metadataServerToken{serviceURL: serviceURL}
}

// GetRequestMetadata вызывается при каждом запросе, поэтому мы уверены, что срок действия токена всегда не истек
func (t metadataServerToken) GetRequestMetadata(ctx context.Context, in ...string) (map[string]string, error) {
	// основываясь на https://cloud.google.com/run/docs/authenticating/service-to-service#go
	tokenURL := fmt.Sprintf("/instance/service-accounts/default/identity?audience=%s", t.serviceURL)
	idToken, err := metadata.Get(tokenURL)
	if err != nil {
		return nil, errors.Wrap(err, "cannot query id token for gRPC")
	}

	return map[string]string{
		"authorization": "Bearer " + idToken,
	}, nil
}

func (metadataServerToken) RequireTransportSecurity() bool {
	return true
}
