package computemetadata

import (
	"context"
	"errors"
	"fmt"

	"cloud.google.com/go/compute/metadata"
	"github.com/GoogleCloudPlatform/docker-credential-gcr/v2/config"
)

var (
	errNotOnGCE                 = errors.New("not on GCE")
	errNoArtifactRegistryDomain = errors.New("no Artifact Registry domain")
)

type metadataGetter interface {
	GetWithContext(context.Context, string) (string, error)
}

func GetGCEArtifactRegistryDockerDomain(ctx context.Context) (string, error) {
	return getGCEArtifactRegistryDockerDomain(ctx, metadata.OnGCE, metadata.NewClient(nil))
}

func getGCEArtifactRegistryDockerDomain(ctx context.Context, onGCE func() bool, client metadataGetter) (string, error) {
	if !onGCE() {
		return "", errNotOnGCE
	}
	arDomain, err := client.GetWithContext(ctx, config.GCEArtifactRegistryDomainKey)
	if err != nil {
		return "", err
	}
	if arDomain == "" {
		return "", errNoArtifactRegistryDomain
	}
	return fmt.Sprintf("docker.%s", arDomain), nil
}
