# Unoserver Web

Web server for converting files using [unoserver](https://github.com/unoconv/unoserver)

## Example

```sh
curl \
--request POST 'http://localhost:3000/convert/pdf' \
--form 'file=@"/path/to/file.docx"' \
-o my.pdf
```

## Server options

See `.env.example`
