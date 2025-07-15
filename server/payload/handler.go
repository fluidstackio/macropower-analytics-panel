package payload

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/MacroPower/macropower-analytics-panel/server/cacher"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
)

// Handler is the handler for incoming payloads.
type Handler struct {
	logger         log.Logger
	ch             chan Payload
	grafanaAuthURL string
	httpClient     *http.Client
}

// NewHandler creates a new Handler.
func NewHandler(cache *cacher.Cacher, buffer int, sessionLog bool, variableLog bool, raw bool, grafanaAuthURL string, logger log.Logger) *Handler {
	ch := make(chan Payload, buffer)
	go startProcessor(cache, ch, sessionLog, variableLog, raw, logger)

	// Create HTTP client with timeout for authentication requests
	httpClient := &http.Client{
		Timeout: 10 * time.Second,
	}

	return &Handler{
		logger:         logger,
		ch:             ch,
		grafanaAuthURL: grafanaAuthURL,
		httpClient:     httpClient,
	}
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	p := Payload{}

	// Extract the cookie from the request
	cookie, err := r.Cookie("grafana_session")
	if err != nil {
		level.Debug(h.logger).Log("msg", "Missing grafana_session cookie", "error", err)
		http.Error(w, "Authentication required", http.StatusUnauthorized)
		return
	}

	// Authenticate with Grafana if auth URL is configured
	if h.grafanaAuthURL != "" {
		if !h.authenticateWithGrafana(cookie) {
			level.Debug(h.logger).Log("msg", "Grafana authentication failed", "cookie_name", cookie.Name)
			http.Error(w, "Authentication failed", http.StatusUnauthorized)
			return
		}
		level.Debug(h.logger).Log("msg", "Grafana authentication successful")
	}

	// Parse the JSON payload
	err = json.NewDecoder(r.Body).Decode(&p)
	if err != nil {
		level.Error(h.logger).Log("msg", "Failed to decode JSON payload", "error", err)
		http.Error(w, "Invalid JSON payload", http.StatusBadRequest)
		return
	}

	// Send payload to processor
	h.ch <- p

	w.WriteHeader(http.StatusCreated)
	fmt.Fprint(w, "")
}

// authenticateWithGrafana validates the session cookie with Grafana
func (h *Handler) authenticateWithGrafana(cookie *http.Cookie) bool {
	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Create request to Grafana auth endpoint
	req, err := http.NewRequestWithContext(ctx, "GET", h.grafanaAuthURL, nil)
	if err != nil {
		level.Error(h.logger).Log("msg", "Failed to create auth request", "error", err)
		return false
	}

	// Add the session cookie
	req.AddCookie(cookie)

	// Make the request
	resp, err := h.httpClient.Do(req)
	if err != nil {
		level.Error(h.logger).Log("msg", "Failed to authenticate with Grafana", "error", err)
		return false
	}
	defer resp.Body.Close()

	// Check if authentication was successful
	if resp.StatusCode != http.StatusOK {
		level.Debug(h.logger).Log("msg", "Grafana authentication failed", "status_code", resp.StatusCode)
		return false
	}

	return true
}

// startProcessor starts a receiver and optional logger for the Payload channel.
func startProcessor(cache *cacher.Cacher, c <-chan Payload, sessionLog bool, variableLog bool, raw bool, logger log.Logger) {
	for p := range c {
		if p.Dashboard.UID != "new" {
			processPayload(cache, p, logger)
		}
		if sessionLog {
			LogPayload(p, variableLog, logger, raw)
		}
	}
}

// processPayload is a receiver for Payloads.
func processPayload(cache *cacher.Cacher, p Payload, logger log.Logger) {
	switch p.Type {
	case "start":
		addStart(cache, p)
	case "heartbeat":
		addHeartbeat(cache, p)
	case "end":
		addEnd(cache, p)
	default:
		addHeartbeat(cache, p)
		_ = level.Warn(logger).Log(
			"msg", "Session has invalid type, defaulted to heartbeat",
			"uuid", p.UUID,
			"type", p.Type,
		)
	}
}

// LogPayload writes a log describing the Payload.
func LogPayload(p Payload, logVars bool, logger log.Logger, raw bool) {
	if !logVars {
		p.Variables = p.Variables[:0]
	}

	if raw {
		level.Info(logger).Log("msg", "Received session data", "data", p)
		return
	}

	h := p.Host
	bi := h.BuildInfo
	li := h.LicenseInfo
	u := p.User
	tr := p.TimeRange

	var theme string
	if u.LightTheme {
		theme = "light"
	} else {
		theme = "dark"
	}

	var role string
	if u.IsGrafanaAdmin {
		role = "admin"
	} else if u.HasEditPermissionInFolders {
		role = "editor"
	} else {
		role = "user"
	}

	labels := []interface{}{
		"msg", "Received session data",
		"uuid", p.UUID,
		"type", p.Type,
		"has_focus", p.HasFocus,
		"host", fmt.Sprintf("%s//%s:%s", h.Protocol, h.Hostname, h.Port),
		"build", fmt.Sprintf("(commit=%s, edition=%s, env=%s, version=%s)", bi.Commit, bi.Edition, bi.Env, bi.Version),
		"license", fmt.Sprintf("(state=%s, expiry=%d, license=%t)", li.StateInfo, li.Expiry, li.HasLicense),
		"dashboard_name", p.Dashboard.Name,
		"dashboard_uid", p.Dashboard.UID,
		"dashboard_timezone", p.TimeZone,
		"user_id", u.ID,
		"user_login", u.Login,
		"user_email", u.Email,
		"user_name", u.Name,
		"user_theme", theme,
		"user_role", role,
		"user_locale", u.Locale,
		"user_timezone", u.Timezone,
		"time_from", tr.From,
		"time_to", tr.To,
		"time_from_raw", tr.Raw.From,
		"time_to_raw", tr.Raw.To,
		"timeorigin", p.TimeOrigin,
		"time", p.Time,
	}

	for _, v := range p.Variables {
		var variableValues []string
		for _, value := range v.Values {
			variableValues = append(variableValues, value.(string))
		}
		d := fmt.Sprintf("(label=%s, type=%s, multi=%t, count=%d, values=[%s])", v.Label, v.Type, v.Multi, len(v.Values), strings.Join(variableValues, ","))
		labels = append(labels, v.Name, d)
	}

	_ = level.Info(logger).Log(labels...)
}
