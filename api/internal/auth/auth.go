package auth

import (
	"context"
	"net"
	"net/http"
	"strings"

	"github.com/rs/zerolog/log"
)

// Auth enforces both API-key and OIDC/JWT based authentication
type Auth struct {
	AdminKey      string
	DeviceKey     string
	NoAuthIPs     []net.IP     // Individual IP addresses that bypass authentication
	NoAuthSubnets []*net.IPNet // Subnets (CIDR) that bypass authentication
	OIDCEnabled   bool
	OIDCVerifier  *OIDCVerifier
}

// isIPWhitelisted checks if the remote IP is in the no-auth whitelist (IPs or subnets)
func (a Auth) isIPWhitelisted(remoteAddr string) bool {
	if len(a.NoAuthIPs) == 0 && len(a.NoAuthSubnets) == 0 {
		return false
	}

	// Extract IP from "IP:port" format using net.SplitHostPort
	// This properly handles both IPv4 (127.0.0.1:8080) and IPv6 ([::1]:8080)
	host, _, err := net.SplitHostPort(remoteAddr)
	if err != nil {
		// If splitting fails, use the whole address (might not have a port)
		host = remoteAddr
	}

	// Parse the client IP
	clientIP := net.ParseIP(host)
	if clientIP == nil {
		log.Warn().Str("remote_addr", remoteAddr).Msg("Failed to parse client IP")
		return false
	}

	// Check against individual IPs
	for _, allowedIP := range a.NoAuthIPs {
		// Compare IPs (this handles IPv4/IPv6 equivalence like 127.0.0.1 == ::1)
		if clientIP.Equal(allowedIP) {
			return true
		}
	}

	// Check against subnets (CIDR ranges)
	for _, subnet := range a.NoAuthSubnets {
		if subnet.Contains(clientIP) {
			return true
		}
	}

	return false
}

func (a Auth) RequireAdmin(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Check if IP is whitelisted (bypass all authentication)
		if a.isIPWhitelisted(r.RemoteAddr) {
			log.Debug().
				Str("path", r.URL.Path).
				Str("method", r.Method).
				Str("remote_addr", r.RemoteAddr).
				Str("auth_type", "ip_whitelist").
				Str("role", "admin").
				Msg("Admin authentication bypassed via IP whitelist")
			next(w, r)
			return
		}

		// If OIDC is enabled, try JWT first, then fall back to API key
		if a.OIDCEnabled && a.OIDCVerifier != nil {
			if a.verifyJWT(w, r, a.OIDCVerifier.adminRole, "admin") {
				log.Debug().
					Str("path", r.URL.Path).
					Str("method", r.Method).
					Str("auth_type", "jwt").
					Str("role", "admin").
					Msg("Admin authentication successful via JWT")
				next(w, r)
				return
			}
		}

		// Fall back to API key authentication
		if a.AdminKey != "" && r.Header.Get("X-Admin-Key") == a.AdminKey {
			log.Debug().
				Str("path", r.URL.Path).
				Str("method", r.Method).
				Str("auth_type", "api_key").
				Str("role", "admin").
				Msg("Admin authentication successful via API key")
			next(w, r)
			return
		}

		log.Warn().
			Str("path", r.URL.Path).
			Str("method", r.Method).
			Str("remote_addr", r.RemoteAddr).
			Str("role", "admin").
			Msg("Admin authentication failed")
		http.Error(w, "unauthorized (admin)", http.StatusUnauthorized)
	}
}

func (a Auth) RequireDevice(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Check if IP is whitelisted (bypass all authentication)
		if a.isIPWhitelisted(r.RemoteAddr) {
			log.Debug().
				Str("path", r.URL.Path).
				Str("method", r.Method).
				Str("remote_addr", r.RemoteAddr).
				Str("auth_type", "ip_whitelist").
				Str("role", "device").
				Msg("Device authentication bypassed via IP whitelist")
			next(w, r)
			return
		}

		// If OIDC is enabled, try JWT first, then fall back to API key
		if a.OIDCEnabled && a.OIDCVerifier != nil {
			if a.verifyJWT(w, r, a.OIDCVerifier.deviceRole, "device") {
				log.Debug().
					Str("path", r.URL.Path).
					Str("method", r.Method).
					Str("auth_type", "jwt").
					Str("role", "device").
					Msg("Device authentication successful via JWT")
				next(w, r)
				return
			}
		}

		// Fall back to API key authentication
		if a.DeviceKey != "" && r.Header.Get("X-Device-Key") == a.DeviceKey {
			log.Debug().
				Str("path", r.URL.Path).
				Str("method", r.Method).
				Str("auth_type", "api_key").
				Str("role", "device").
				Msg("Device authentication successful via API key")
			next(w, r)
			return
		}

		log.Warn().
			Str("path", r.URL.Path).
			Str("method", r.Method).
			Str("remote_addr", r.RemoteAddr).
			Str("role", "device").
			Msg("Device authentication failed")
		http.Error(w, "unauthorized (device)", http.StatusUnauthorized)
	}
}

// verifyJWT validates the JWT token from the Authorization header and checks for the required role
func (a Auth) verifyJWT(w http.ResponseWriter, r *http.Request, requiredRole string, roleType string) bool {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return false
	}

	token := ExtractBearerToken(authHeader)
	if token == "" {
		log.Debug().Msg("No Bearer token found in Authorization header")
		return false
	}

	ctx := context.Background()
	idToken, err := a.OIDCVerifier.VerifyToken(ctx, token)
	if err != nil {
		log.Warn().
			Err(err).
			Str("path", r.URL.Path).
			Str("remote_addr", r.RemoteAddr).
			Msg("JWT verification failed")
		return false
	}

	hasRole, err := a.OIDCVerifier.HasRole(idToken, requiredRole)
	if err != nil {
		log.Error().
			Err(err).
			Str("required_role", requiredRole).
			Msg("Failed to check role in JWT")
		return false
	}

	if !hasRole && requiredRole != "" {
		log.Warn().
			Str("required_role", requiredRole).
			Str("role_type", roleType).
			Str("path", r.URL.Path).
			Str("remote_addr", r.RemoteAddr).
			Msg("User missing required role")
		return false
	}

	return true
}

// ExtractBearerToken is a helper to extract Bearer token from Authorization header
func ExtractBearerToken(authHeader string) string {
	parts := strings.Split(authHeader, " ")
	if len(parts) == 2 && strings.ToLower(parts[0]) == "bearer" {
		return parts[1]
	}
	return ""
}
