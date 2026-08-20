package http

import (
	"mime"
	"strings"
)

// officeMimeTypes keeps content types for formats LibreOffice converts to, so the
// response doesn't depend on the host's /etc/mime.types database being complete.
var officeMimeTypes = map[string]string{
	".doc":  "application/msword",
	".dot":  "application/msword",
	".docx": "application/vnd.openxmlformats-officedocument.wordprocessingml.document",
	".docm": "application/vnd.ms-word.document.macroenabled.12",
	".rtf":  "application/rtf",
	".odt":  "application/vnd.oasis.opendocument.text",
	".ott":  "application/vnd.oasis.opendocument.text-template",
	".fodt": "application/vnd.oasis.opendocument.text-flat-xml",
	".xls":  "application/vnd.ms-excel",
	".xlsx": "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
	".ods":  "application/vnd.oasis.opendocument.spreadsheet",
	".fods": "application/vnd.oasis.opendocument.spreadsheet-flat-xml",
	".ppt":  "application/vnd.ms-powerpoint",
	".pptx": "application/vnd.openxmlformats-officedocument.presentationml.presentation",
	".odp":  "application/vnd.oasis.opendocument.presentation",
	".fodp": "application/vnd.oasis.opendocument.presentation-flat-xml",
	".odg":  "application/vnd.oasis.opendocument.graphics",
	".fodg": "application/vnd.oasis.opendocument.graphics-flat-xml",
	".epub": "application/epub+zip",
}

// contentType returns the content type for a file extension, falling back to the
// system mime database and then to application/octet-stream
func contentType(ext string) string {
	if mimeType, ok := officeMimeTypes[strings.ToLower(ext)]; ok {
		return mimeType
	}

	if mimeType := mime.TypeByExtension(ext); mimeType != "" {
		return mimeType
	}

	return "application/octet-stream"
}
