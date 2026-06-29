package domain

import (
	"mime"
	"net/url"
	"path"
	"strings"
)

type ResourceType string

const (
	ResourceHTML     ResourceType = "html"
	ResourceAPI      ResourceType = "api"
	ResourceScript   ResourceType = "script"
	ResourceStyle    ResourceType = "style"
	ResourceImage    ResourceType = "image"
	ResourceFont     ResourceType = "font"
	ResourceMedia    ResourceType = "media"
	ResourceDocument ResourceType = "document"
	ResourceArchive  ResourceType = "archive"
	ResourceOther    ResourceType = "other"
)

var extensionResourceTypes = map[string]ResourceType{
	// Page routes and server-rendered documents.
	".html": ResourceHTML,
	".htm":  ResourceHTML,
	".php":  ResourceHTML,
	".asp":  ResourceHTML,
	".aspx": ResourceHTML,
	".jsp":  ResourceHTML,

	// Structured API-like payloads.
	".json": ResourceAPI,

	// Browser-loaded assets.
	".js":   ResourceScript,
	".mjs":  ResourceScript,
	".wasm": ResourceScript,

	".css": ResourceStyle,
	".map": ResourceStyle,

	".png":  ResourceImage,
	".jpg":  ResourceImage,
	".jpeg": ResourceImage,
	".gif":  ResourceImage,
	".svg":  ResourceImage,
	".webp": ResourceImage,
	".avif": ResourceImage,
	".ico":  ResourceImage,
	".bmp":  ResourceImage,
	".tif":  ResourceImage,
	".tiff": ResourceImage,

	".woff":  ResourceFont,
	".woff2": ResourceFont,
	".ttf":   ResourceFont,
	".otf":   ResourceFont,
	".eot":   ResourceFont,

	".mp4":  ResourceMedia,
	".webm": ResourceMedia,
	".mov":  ResourceMedia,
	".avi":  ResourceMedia,
	".mp3":  ResourceMedia,
	".wav":  ResourceMedia,
	".ogg":  ResourceMedia,
	".m4a":  ResourceMedia,

	// Downloads and document-like responses.
	".pdf":  ResourceDocument,
	".txt":  ResourceDocument,
	".csv":  ResourceDocument,
	".xml":  ResourceDocument,
	".doc":  ResourceDocument,
	".docx": ResourceDocument,
	".xls":  ResourceDocument,
	".xlsx": ResourceDocument,
	".ppt":  ResourceDocument,
	".pptx": ResourceDocument,

	".zip": ResourceArchive,
	".gz":  ResourceArchive,
	".br":  ResourceArchive,
	".tar": ResourceArchive,
	".rar": ResourceArchive,
	".7z":  ResourceArchive,
}

// ClassifyResourceType maps a request URI into Streamline's controlled
// resource vocabulary. It is intentionally coarse: downstream features
// aggregate these buckets into numeric ratios for HBOS.
func ClassifyResourceType(rawURI string) ResourceType {
	requestPath := EndpointPath(rawURI)
	if requestPath == "" {
		return ResourceOther
	}

	lowerPath := strings.ToLower(requestPath)

	// API-like routes should win over extension-based classification.
	if isAPIPath(lowerPath) {
		return ResourceAPI
	}

	ext := strings.ToLower(path.Ext(lowerPath))
	if ext == "" {
		return ResourceHTML
	}

	if resourceType, ok := extensionResourceTypes[ext]; ok {
		return resourceType
	}

	if resourceType := classifyByMIMEExtension(ext); resourceType != "" {
		return resourceType
	}

	return ResourceOther
}

// EndpointPath reduces a request URI to its endpoint identity: the path with
// query and fragment removed. Two requests share an endpoint when their paths
// match, so /p/1 and /p/2 differ but /p/1?ref=a and /p/1?ref=b do not. Used by
// resource classification and by the baseline worker's distinct_endpoints_hit
// feature, which must strip paths identically on the fit and scoring sides.
func EndpointPath(rawURI string) string {
	rawURI = strings.TrimSpace(rawURI)
	if rawURI == "" {
		return ""
	}

	parsed, err := url.ParseRequestURI(rawURI)
	if err == nil && parsed.Path != "" {
		return parsed.Path
	}

	// Keep classification total for imperfect log data.
	if i := strings.IndexAny(rawURI, "?#"); i >= 0 {
		rawURI = rawURI[:i]
	}

	return rawURI
}

func isAPIPath(p string) bool {
	return p == "/api" ||
		strings.HasPrefix(p, "/api/") ||
		p == "/ajax" ||
		strings.HasPrefix(p, "/ajax/") ||
		p == "/graphql" ||
		strings.HasPrefix(p, "/graphql/")
}

func classifyByMIMEExtension(ext string) ResourceType {
	mimeType := strings.ToLower(mime.TypeByExtension(ext))
	if mimeType == "" {
		return ""
	}

	switch {
	case strings.HasPrefix(mimeType, "text/html"):
		return ResourceHTML

	case strings.Contains(mimeType, "json"):
		return ResourceAPI

	case strings.Contains(mimeType, "javascript"),
		strings.Contains(mimeType, "ecmascript"),
		strings.Contains(mimeType, "wasm"):
		return ResourceScript

	case strings.HasPrefix(mimeType, "text/css"):
		return ResourceStyle

	case strings.HasPrefix(mimeType, "image/"):
		return ResourceImage

	case strings.HasPrefix(mimeType, "font/"):
		return ResourceFont

	case strings.HasPrefix(mimeType, "audio/"),
		strings.HasPrefix(mimeType, "video/"):
		return ResourceMedia

	case strings.Contains(mimeType, "pdf"),
		strings.Contains(mimeType, "xml"),
		strings.HasPrefix(mimeType, "text/"):
		return ResourceDocument

	case strings.Contains(mimeType, "zip"),
		strings.Contains(mimeType, "gzip"),
		strings.Contains(mimeType, "x-tar"),
		strings.Contains(mimeType, "x-7z"):
		return ResourceArchive

	default:
		return ""
	}
}