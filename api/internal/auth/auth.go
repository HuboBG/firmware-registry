package auth

import (
	"context"
	"net/http"
	"strings"

	"github.com/rs/zerolog/log"
)

// Auth enforces both API-key and OIDC/JWT based authentication
type Auth struct {
	AdminKey     string
	DeviceKey    string
	OIDCEnabled  bool
	OIDCVerifier *OIDCVerifier
}

func (a Auth) RequireAdmin(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
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
