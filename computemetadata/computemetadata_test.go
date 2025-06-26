package computemetadata

import (
	"context"
	"errors"
	"testing"
)

var getMetadataErr = errors.New("can't get metadata")

type fakeMetadataGetter struct {
	md  string
	err error
}

func (m fakeMetadataGetter) GetWithContext(_ context.Context, suffix string) (string, error) {
	return m.md, m.err
}

func TestGetGCEArtifactRegistryDockerDomain(t *testing.T) {
	cases := []struct {
		name           string
		metadataGetter metadataGetter
		onGCE          func() bool
		wantErr        error
		wantDomain     string
	}{
		{
			name:           "success",
			metadataGetter: fakeMetadataGetter{md: "artifact.registry"},
			onGCE:          func() bool { return true },
			wantDomain:     "docker.artifact.registry",
		},
		{
			name:           "error not on GCE",
			metadataGetter: fakeMetadataGetter{md: "artifact.registry"},
			onGCE:          func() bool { return false },
			wantErr:        errNotOnGCE,
		},
		{
			name:           "get metadata error",
			metadataGetter: fakeMetadataGetter{err: getMetadataErr},
			onGCE:          func() bool { return true },
			wantErr:        getMetadataErr,
		},
		{
			name:           "artifact registry domain empty",
			metadataGetter: fakeMetadataGetter{err: errNoArtifactRegistryDomain},
			onGCE:          func() bool { return true },
			wantErr:        errNoArtifactRegistryDomain,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			gotDomain, gotErr := getGCEArtifactRegistryDockerDomain(context.Background(), tc.onGCE, tc.metadataGetter)
			if !errors.Is(gotErr, tc.wantErr) {
				t.Errorf("unexpected error: got %v, want %v", gotErr, tc.wantErr)
			}
			if gotErr != nil {
				if gotDomain != tc.wantDomain {
					t.Errorf("unexpected domain: got %v, want %v", gotDomain, tc.wantDomain)
				}
			}
		})
	}
}
