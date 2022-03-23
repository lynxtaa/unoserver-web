# Unoserver Web

Web server for converting files using [unoserver](https://github.com/unoconv/unoserver)

![CI/CD](https://github.com/lynxtaa/unoserver-web/workflows/CI/CD/badge.svg)
![Codecov](https://img.shields.io/codecov/c/github/lynxtaa/unoserver-web)

## Example

```sh
curl \
--request POST 'http://localhost:3000/convert/pdf' \
--form 'file=@"/path/to/file.docx"' \
-o my.pdf
```

## Container Environment

| Variable    | Description                           | Default | Required |
| ----------- | ------------------------------------- | ------- | -------- |
| PORT        | Application port                      | 3000    | no       |
| MAX_WORKERS | Maximum number of LibreOffice workers | 8       | no       |
| BASE_PATH   | Prefix path                           |         | no       |
