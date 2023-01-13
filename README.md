# Unoserver Web

Web server for converting files using [unoserver](https://github.com/unoconv/unoserver)

![CI/CD](https://github.com/lynxtaa/unoserver-web/workflows/CI/CD/badge.svg)
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

## Container Environment

| Variable    | Description                           | Default |
| ----------- | ------------------------------------- | ------- |
| PORT        | Application port                      | 3000    |
| MAX_WORKERS | Maximum number of LibreOffice workers | 8       |
| BASE_PATH   | Prefix path                           |         |

## Development

To spin up the container for development use [VSCode Remote Containers](https://code.visualstudio.com/docs/devcontainers/containers) feature. See `.devcontainer/devcontainer.json` for reference.

If you need to support more custom fonts, you could add them to `fonts` folder.

Commands:

- `pnpm run dev` - runs the app in watch-mode, then you could access a Swagger UI from `http://0.0.0.0:3000`
- `pnpm run build && pnpm run start` - builds and starts a production version of the app
- `pnpm run validate` - runs linting, typechecking and formatting check
- `pnpm run test` - runs all the tests

### Building an image

```sh
docker build --build-arg NODE_ENV=production --tag unoserver-web:dev .
docker run --rm -p 3000:3000 unoserver-web:dev
```
