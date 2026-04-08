package trafilatura

import (
	"strings"

	"github.com/go-shiori/dom"
	"github.com/markusmobius/go-trafilatura/internal/etree"
	"golang.org/x/net/html"
)

// isFacebookVideoURL reports whether src is a Facebook CDN or Facebook domain URL.
func isFacebookVideoURL(src string) bool {
	return strings.Contains(src, "fbcdn.net") || strings.Contains(src, "facebook.com")
}

// isVideoEmbedIframe reports whether the iframe element is a recognized video embed
// (YouTube or Vimeo).
func isVideoEmbedIframe(node *html.Node) bool {
	src := dom.GetAttribute(node, "src")
	return strings.Contains(src, "youtube.com/embed/") ||
		strings.Contains(src, "player.vimeo.com/video/")
}

// isVideoEmbedBlockquote reports whether the blockquote element is a TikTok embed.
func isVideoEmbedBlockquote(node *html.Node) bool {
	return strings.Contains(dom.ClassName(node), "tiktok-embed")
}

// isVideoEmbedVideo reports whether the video element contains a Facebook-hosted video.
func isVideoEmbedVideo(node *html.Node) bool {
	if isFacebookVideoURL(dom.GetAttribute(node, "src")) {
		return true
	}
	for _, source := range dom.GetElementsByTagName(node, "source") {
		if isFacebookVideoURL(dom.GetAttribute(source, "src")) {
			return true
		}
	}
	return false
}

// handleVideoEmbed sanitizes and returns a video embed element. Returns nil if the element
// is not a recognized embed or its source does not pass the allowlist check.
func handleVideoEmbed(node *html.Node) *html.Node {
	switch dom.TagName(node) {
	case "iframe":
		return sanitizeIframeEmbed(node)
	case "blockquote":
		return sanitizeBlockquoteEmbed(node)
	case "video":
		return sanitizeVideoEmbed(node)
	}
	return nil
}

// sanitizeIframeEmbed returns a sanitized copy of a YouTube or Vimeo iframe embed.
// Only src, width, height, frameborder, and allowfullscreen attributes are kept.
// The src is re-validated against the allowlist to prevent arbitrary iframe injection.
func sanitizeIframeEmbed(node *html.Node) *html.Node {
	src := dom.GetAttribute(node, "src")
	if !strings.Contains(src, "youtube.com/embed/") && !strings.Contains(src, "player.vimeo.com/video/") {
		return nil
	}

	result := etree.Element("iframe")
	dom.SetAttribute(result, "src", src)

	if w := dom.GetAttribute(node, "width"); w != "" {
		dom.SetAttribute(result, "width", w)
	}
	if h := dom.GetAttribute(node, "height"); h != "" {
		dom.SetAttribute(result, "height", h)
	}
	if fb := dom.GetAttribute(node, "frameborder"); fb != "" {
		dom.SetAttribute(result, "frameborder", fb)
	}
	if dom.HasAttribute(node, "allowfullscreen") {
		dom.SetAttribute(result, "allowfullscreen", "")
	}

	return result
}

// sanitizeBlockquoteEmbed returns a sanitized copy of a TikTok blockquote embed.
// Only class, cite, and data-video-id attributes are kept; all child nodes are stripped.
func sanitizeBlockquoteEmbed(node *html.Node) *html.Node {
	if !strings.Contains(dom.ClassName(node), "tiktok-embed") {
		return nil
	}

	result := etree.Element("blockquote")
	if class := dom.GetAttribute(node, "class"); class != "" {
		dom.SetAttribute(result, "class", class)
	}
	if cite := dom.GetAttribute(node, "cite"); cite != "" {
		dom.SetAttribute(result, "cite", cite)
	}
	if vid := dom.GetAttribute(node, "data-video-id"); vid != "" {
		dom.SetAttribute(result, "data-video-id", vid)
	}

	return result
}

// sanitizeVideoEmbed returns a sanitized copy of a Facebook video embed.
// Only src, width, height, and controls attributes are kept. Child <source> elements
// are preserved only when their src passes the Facebook CDN allowlist check.
// Event-handler attributes (e.g. onplay, onload) and autoplay are never included.
func sanitizeVideoEmbed(node *html.Node) *html.Node {
	result := etree.Element("video")

	if src := dom.GetAttribute(node, "src"); isFacebookVideoURL(src) {
		dom.SetAttribute(result, "src", src)
	}
	if w := dom.GetAttribute(node, "width"); w != "" {
		dom.SetAttribute(result, "width", w)
	}
	if h := dom.GetAttribute(node, "height"); h != "" {
		dom.SetAttribute(result, "height", h)
	}
	if dom.HasAttribute(node, "controls") {
		dom.SetAttribute(result, "controls", "")
	}

	for _, source := range dom.GetElementsByTagName(node, "source") {
		src := dom.GetAttribute(source, "src")
		if isFacebookVideoURL(src) {
			sourceElem := etree.Element("source")
			dom.SetAttribute(sourceElem, "src", src)
			dom.AppendChild(result, sourceElem)
		}
	}

	if dom.GetAttribute(result, "src") == "" && len(dom.Children(result)) == 0 {
		return nil
	}

	return result
}
