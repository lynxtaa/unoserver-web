# Unoserver Web

Web server for converting files using [unoserver](https://github.com/unoconv/unoserver)

![CI](https://github.com/lynxtaa/unoserver-web/workflows/CI/badge.svg)
![Codecov](https://img.shields.io/codecov/c/github/lynxtaa/unoserver-web)

## Example

Using [Dockerhub Image](https://hub.docker.com/r/lynxtaa/unoserver-web):

```sh
docker run -d -p 3000:3000 lynxtaa/unoserver-web:latest

curl \
--request POST 'http://localhost:3000/convert/pdf' \
--form 'file=@"/path/to/file.docx"' \
-o my.pdf
```

## Endpoints

| Route                    | Description                                         |
| ------------------------ | --------------------------------------------------- |
| `POST /convert/{format}` | Converts the uploaded `file` to `format`            |
| `GET /documentation/`    | Swagger UI                                          |
| `GET /health`            | Health check, returns `200 OK`                      |
| `GET /metrics`           | Go runtime and process metrics in Prometheus format |

`/health` and `/metrics` are always served from the root, ignoring `BASE_PATH`.

## Container Environment

| Variable             | Description                                                                | Default |
| -------------------- | -------------------------------------------------------------------------- | ------- |
| PORT                 | Application port                                                           | 3000    |
| MAX_WORKERS          | Maximum number of LibreOffice workers                                      | 8       |
| CONVERSION_RETRIES   | Number of retries for converting input file                                | 3       |
| MAX_FILE_SIZE        | Maximum uploaded file size in bytes (0 means unlimited)                    | 0       |
| BASE_PATH            | Prefix path the app is served behind, used in Swagger and redirects        |         |
| LOG_LEVEL            | Minimum logged level: debug, info, warn or error                           | info    |
| PRETTY_LOGS          | Log human-readable text instead of JSON                                    | false   |
| REQUEST_ID_HEADER    | The header name used to set the request-id                                 |         |
| REQUEST_ID_LOG_LABEL | Defines the label used for the request identifier when logging the request | reqId   |

## Development

To spin up the container for development use [VSCode Remote Containers](https://code.visualstudio.com/docs/devcontainers/containers) feature. See `.devcontainer/devcontainer.json` for reference.

If you need to support more custom fonts, you could add them to `fonts` folder.

Commands:

- `make run` - runs the app, then you could access a Swagger UI from `http://0.0.0.0:3000`
- `make test` - runs all the tests

### Building an image

```sh
docker build --tag unoserver-web:dev .
docker run --rm -p 3000:3000 unoserver-web:dev
```
