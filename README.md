<a href="https://gcr.io"><img src="https://avatars2.githubusercontent.com/u/21046548?s=400&v=4" height="120"/></a>

# docker-credential-gcr [![Build Status](https://travis-ci.org/GoogleCloudPlatform/docker-credential-gcr.svg?branch=master)](https://travis-ci.org/GoogleCloudPlatform/docker-credential-gcr) [![Go Report Card](https://goreportcard.com/badge/GoogleCloudPlatform/docker-credential-gcr)](https://goreportcard.com/report/GoogleCloudPlatform/docker-credential-gcr)

## Introduction

`docker-credential-gcr` is [Google Container Registry](https://cloud.google.com/container-registry/)'s _standalone_, `gcloud` SDK-independent Docker credential helper. It allows for **Docker clients since v1.11** to easily make authenticated requests to GCR's repositories (gcr.io, eu.gcr.io, etc.).

**Note:** `docker-credential-gcr` is primarily intended for users wishing to authenticate with GCR in the **absence of `gcloud`**, though they are [not mutually exclusive](#gcr-credentials). For normal development setups, users are encouraged to use [`gcloud auth configure-docker`](https://cloud.google.com/sdk/gcloud/reference/auth/configure-docker), instead.

The helper implements the [Docker Credential Store](https://docs.docker.com/engine/reference/commandline/login/#/credentials-store) API, but enables more advanced authentication schemes for GCR's users. In particular, it respects [Application Default Credentials](https://developers.google.com/identity/protocols/application-default-credentials) and is capable of generating credentials automatically (without an explicit login operation) when running in App Engine or Compute Engine.

For even more authentication options, see GCR's documentation on [advanced authentication methods](https://cloud.google.com/container-registry/docs/advanced-authentication).

## GCR Credentials

By default, the helper searches for GCR credentials in the following order:

1. In the helper's private credential store (i.e. those stored via `docker-credential-gcr gcr-login`)
1. From the `gcloud` SDK (i.e. the one printed via `gcloud config config-helper --force-auth-refresh --format='value(credential.access_token)'`).
1. In a JSON file whose path is specified by the GOOGLE_APPLICATION_CREDENTIALS environment variable.
1. In a JSON file in a location known to the helper:
	* On Windows, this is `%APPDATA%/gcloud/application_default_credentials.json`.
	* On other systems, `$HOME/.config/gcloud/application_default_credentials.json`.
1. On Google App Engine, it uses the `appengine.AccessToken` function.
1. On Google Compute Engine, Kubernetes Engine, and App Engine Managed VMs, it fetches the credentials of the _service account_ associated with the VM from the metadata server (if available).

Users may limit or re-order how the helper searches for GCR credentials using `docker-credential-gcr config --token-source`. Number 1 above is designated by the `store`, 2 by `gcloud`, and 3-6 by `env` (which cannot be individually restricted or re-ordered). Multiple sources are separated by commas, and the default is `"store, gcloud, env"`.

**Examples:**

To configure the credential helper to use only the gcloud SDK's access token:
```shell
docker-credential-gcr config --token-source="gcloud"
```

To try the private store, followed by the environment:
```shell
docker-credential-gcr config --token-source="store, env"
```

To verify that credentials are being returned for a given registry, e.g. for `https://gcr.io`:

```shell
echo "https://gcr.io" | docker-credential-gcr get
```

## Other Credentials

The helper implements the [Docker Credential Store](https://docs.docker.com/engine/reference/commandline/login/#/credentials-store) API and can be used to store credentials for other repositories. **WARNING**: Credentials are stored in plain text in a file under the user's home directory (e.g. $HOME/.config/gcloud/docker_credentials.json on non-windows systems).

### Building from Source

The program in this repository is written with the Go programming language and built with `make`. These instructions assume that [**Go 1.7+**](https://golang.org/) and `make` are installed on a *nix system.

You can download the source code, compile the binary, and put it in your `$GOPATH` with `go get`.

```shell
go get -u github.com/GoogleCloudPlatform/docker-credential-gcr
```

If `$GOPATH/bin` is in your system `$PATH`, this will also automatically install the compiled binary. You can confirm using `which docker-credential-gcr` and continue to the [section on Configuration and Usage](#configuration-and-usage).

Alternatively, you can use `make` to build the program. The executable will be output to the `bin` directory inside the repository.

```shell
cd $GOPATH/src/github.com/GoogleCloudPlatform/docker-credential-gcr
make
```

Then, you can put that binary in your `$PATH` to make it visible to `docker`. For example, if `/usr/bin` is present in your system path:

```shell
sudo mv ./bin/docker-credential-gcr /usr/bin/docker-credential-gcr
```

## Configuration and Usage

* Configure the Docker CLI to use docker-credential-gcr as its credential store:

	```shell
	docker-credential-gcr configure-docker
	```

  * Alternatively, use the [manual configuration instructions](#manual-docker-client-configuration) below to configure your version of the Docker client.

* Log in to GCR (or don't! See the [GCR Credentials section](#gcr-credentials))

	```shell
	docker-credential-gcr gcr-login
	```

* Use Docker!

	```shell
	docker pull gcr.io/project-id/neato-container
	```

* Log out from GCR

	```shell
	docker-credential-gcr gcr-logout
	```

### Manual Docker Client Configuration
#### **(Recommended)** Using `credHelpers`, for Docker clients since v1.13.0

Add a `credHelpers` entry in the Docker config file (usually `~/.docker/config.json` on OSX and Linux, `%USERPROFILE%\.docker\config.json` on Windows) for each GCR registry that you care about. The key should be the domain of the registry (**without** the "https://") and the key should be the suffix of the credential helper binary (everything after "docker-credential-").

	e.g. for `docker-credential-gcr`:

  <pre>
    {
      "auths" : {
            ...
      }
      "credHelpers": {
            "coolregistry.com": ... ,
            <b>"gcr.io": "gcr",
            "asia.gcr.io": "gcr",
            ...</b>
      },
      "HttpHeaders": ...
      "psFormat": ...
      "imagesFormat": ...
      "detachKeys": ...
    }
  </pre>


#### Using the `credsStore`, for Docker clients since v1.11.0
Set the `credsStore` and `auths` fields in your Docker config file (usually `~/.docker/config.json` on OSX and Linux, `%USERPROFILE%\.docker\config.json` on Windows). The value of `credsStore` should be the suffix of the compiled binary (everything after "docker-credential-") and `auths` should have an empty entry for each GCR endpoint that you care about (**with** the "https://").

	e.g. for `docker-credential-gcr`:

  <pre>
    {
      "auths": {
            "https://coolregistry.com": { ... },
            <b>"https://gcr.io": {},
            "https://asia.gcr.io": {},
            ...</b>
      },
      <b>"credsStore": "gcr",</b>
      "HttpHeaders": ...
      "psFormat": ...
      "imagesFormat": ...
      "detachKeys": ...
    }
  </pre>

## License

Apache 2.0. See [LICENSE](LICENSE) for more information.
