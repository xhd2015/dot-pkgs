package api

import "net/http"

const (
	// HeaderTokenAuth tells cli-chat-proxy to validate a Grok CLI session token.
	HeaderTokenAuth = "X-XAI-Token-Auth"
	// TokenAuthValue is the required value of HeaderTokenAuth.
	TokenAuthValue = "xai-grok-cli"
	// HeaderModelOverride routes cli-chat-proxy to the named model.
	HeaderModelOverride = "X-Grok-Model-Override"
	// HeaderClientVersion is required by cli-chat-proxy (426 without it).
	HeaderClientVersion = "X-Grok-Client-Version"
	// DefaultClientVersion matches a current Grok CLI (grok --version).
	DefaultClientVersion = "1.0.34"
)

// ApplySessionHeaders sets Bearer, CLI session, model override, and client version.
// The client's Authorization is overwritten. model is the unsuffixed id.
func ApplySessionHeaders(h http.Header, auth Auth, model, clientVersion string) {
	if h == nil {
		return
	}
	h.Set("Authorization", "Bearer "+auth.AccessToken)
	h.Set(HeaderTokenAuth, TokenAuthValue)
	if clientVersion == "" {
		clientVersion = DefaultClientVersion
	}
	h.Set(HeaderClientVersion, clientVersion)
	if model != "" {
		h.Set(HeaderModelOverride, model)
	} else {
		h.Del(HeaderModelOverride)
	}
	if h.Get("User-Agent") == "" {
		h.Set("User-Agent", DefaultUserAgent)
	}
}
