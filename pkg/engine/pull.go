package engine

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"io"
	"net/url"

	"github.com/docker/docker/api/types"
)

func (e *Engine) PullImage(reference string) (io.ReadCloser, error) {
	var options types.ImagePullOptions

	registryHost := "https://index.docker.io/v1/"

	if parsed, err := url.Parse("docker://" + reference); err == nil {
		registryHost = parsed.Host
	}

	if auth, ok := e.auth.Auths[registryHost]; ok {
		authJson, _ := json.Marshal(auth)
		encoded := base64.URLEncoding.EncodeToString(authJson)
		options.RegistryAuth = encoded
	}

	// TODO is context.Background() appropriate here?
	return e.api.ImagePull(context.Background(), reference, options)
}
