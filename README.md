# docker-credential-gcr [![Build Status](https://travis-ci.org/GoogleCloudPlatform/docker-credential-gcr.svg?branch=master)](https://travis-ci.org/GoogleCloudPlatform/docker-credential-gcr)

## Introduction

docker-credential-gcr is [Google Container Registry](https://cloud.google.com/container-registry/)'s Docker credential helper. It allows for **Docker clients v1.11+** to easily make authenticated requests to GCR's repositories (gcr.io, eu.gcr.io, etc.).

The helper implements the [Docker Credential Store](https://docs.docker.com/engine/reference/commandline/login/#/credentials-store) API, but enables more advanced authentication schemes for GCR's users. In particular, it respects [Application Default Credentials](https://developers.google.com/identity/protocols/application-default-credentials) and is capable of generating credentials automatically (without an explicit login operation) when running in App Engine or Compute Engine.

## GCR Credentials

By default, the helper searches for GCR credentials in the following order:

1. In a JSON file whose path is specified by the GOOGLE_APPLICATION_CREDENTIALS environment variable.
2. In a JSON file in a location known to the helper. 
	On Windows, this is %APPDATA%/gcloud/application_default_credentials.json.
	On other systems, $HOME/.config/gcloud/application_default_credentials.json.
3. On Google App Engine it uses the appengine.AccessToken function.
4. On Google Compute Engine and Google App Engine Managed VMs, it fetches credentials from the metadata server.
5. From the gcloud SDK (i.e. the one printed via `gcloud config config-helper --format='value(credential.access_token)'`).
6. In the helper's private credential store (i.e. those stored via `docker-credential-gcr gcr-login`)

However, the user may limit or re-order how the helper searches for GCR credentials using `docker-credential-gcr config --token-source`. Numbers 1-4 above are designated by the "env" source, 5 by "gcloud" and 6 by "store". Multiple sources are separated by commas, and the default is "env, gcloud, store".

**Examples:**

To configure the credential helper to use only the gcloud SDK's access token:
```shell
docker-credential-gcr config --token-source="gcloud"
```

To try the private store, followed by the environment:
```shell
docker-credential-gcr config --token-source="store, env"
```

## Other Credentials

The helper implements the [Docker Credential Store](https://docs.docker.com/engine/reference/commandline/login/#/credentials-store) API and can be used to store credentials for other repositories. **WARNING**: Credentials are stored in plain text in a file under the user's home directory (e.g. $HOME/.config/gcloud/docker_credentials.json on non-windows systems).

### Building from Source

The program in this repository is written with the Go programming language and built with `make`. These instructions assume that [**Go 1.7+**](https://golang.org/) and `make` are installed on a *nix system.

1. Download the source and put it in your `$GOPATH` with `go get`.

	```shell
    go get github.com/GoogleCloudPlatform/docker-credential-gcr
	```

2. Use `make` to build the program. The executable will be output to the `bin` directory inside the repository.

	```shell
    cd $GOPATH/src/github.com/GoogleCloudPlatform/docker-credential-gcr
    make
	```

3. Put that binary in your `$PATH`.
	e.g. if `/usr/bin` is present on your path:

	```shell
    sudo mv ./bin/docker-credential-gcr /usr/bin/docker-credential-gcr
	```

## Installation and Usage
* Configure the Docker CLI to use docker-credential-gcr as its credential store

	```shell
    docker-credential-gcr configure-docker
    ```
  * Alternatively, use the instructions below to configure your version of the Docker client.
  
* Log in to GCR (or don't! ```gcloud auth login``` is sufficient, too)

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

### Docker Clients v1.13(.0-rc4)+ Manual Installation

Add a `credHelpers` entry in the Docker config file (usually `~/.docker/config.json`) for each GCR registry that you care about. The key should be the domain of the registry (without the "https://") and the key chould be the suffix of the credential helper binary (everything after "docker-credential-").

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


### Docker Clients v1.11 - v1.12 Manual Installation
Set the `credsStore` and `auths` fields in your Docker config file (usually `~/.docker/config.json`). `credsStore` should be the suffix of the compiled binary (everything after "docker-credential-") and `auths` should have an empty entry for each GCR endpoint that you care about (with the "https://").

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
