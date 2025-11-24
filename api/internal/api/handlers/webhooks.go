package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"firmware-registry-api/internal/auth"
	"firmware-registry-api/internal/util"
	"firmware-registry-api/internal/webhook"
)

// WebhookHandler manages webhook CRUD.
type WebhookHandler struct {
	Auth auth.Auth
	Repo webhook.Repository
}

func (h *WebhookHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/api/webhooks" {
		switch r.Method {
		case http.MethodGet:
			h.Auth.RequireAdmin(func(w http.ResponseWriter, r *http.Request) {
				h.list(w)
			})(w, r)
		case http.MethodPost:
			h.Auth.RequireAdmin(func(w http.ResponseWriter, r *http.Request) {
				h.create(w, r)
			})(w, r)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
		return
	}

	// /api/webhooks/{id}
	idStr := strings.TrimPrefix(r.URL.Path, "/api/webhooks/")
	id, _ := strconv.ParseInt(idStr, 10, 64)
	if id <= 0 {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	h.Auth.RequireAdmin(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPut:
			h.update(w, r, id)
		case http.MethodDelete:
			h.delete(w, id)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})(w, r)
}

// list godoc
// @Summary      List webhooks
// @Description  Get all registered webhooks
// @Tags         webhooks
// @Produce      json
// @Success      200  {array}   webhook.WebhookDTO
// @Failure      401  {string}  string  "Unauthorized"
// @Failure      500  {string}  string  "Database error"
// @Security     ApiKeyAuth
// @Security     BearerAuth
// @Router       /webhooks [get]
func (h *WebhookHandler) list(w http.ResponseWriter) {
	hooks, err := h.Repo.List()
	if err != nil {
		http.Error(w, "db error", http.StatusInternalServerError)
		return
	}

	out := make([]webhook.WebhookDTO, 0, len(hooks))
	for _, x := range hooks {
		out = append(out, webhook.WebhookDTO{
			ID: x.ID, URL: x.URL, Events: x.Events, Enabled: x.Enabled,
		})
	}
	util.WriteJSON(w, out)
}

// create godoc
// @Summary      Create webhook
// @Description  Register a new webhook endpoint
// @Tags         webhooks
// @Accept       json
// @Produce      json
// @Param        webhook  body      webhook.WebhookDTO  true  "Webhook configuration"
// @Success      200      {object}  map[string]int      "Created webhook ID"
// @Failure      400      {string}  string              "Invalid JSON or missing required fields"
// @Failure      401      {string}  string              "Unauthorized"
// @Failure      500      {string}  string              "Database error"
// @Security     ApiKeyAuth
// @Security     BearerAuth
// @Router       /webhooks [post]
func (h *WebhookHandler) create(w http.ResponseWriter, r *http.Request) {
	var dto webhook.WebhookDTO
	if err := json.NewDecoder(r.Body).Decode(&dto); err != nil {
		http.Error(w, "bad json", http.StatusBadRequest)
		return
	}
	if dto.URL == "" || len(dto.Events) == 0 {
		http.Error(w, "url/events required", http.StatusBadRequest)
		return
	}
	if dto.Enabled == false {
		dto.Enabled = true
	}

	id, err := h.Repo.Create(webhook.Webhook{
		URL: dto.URL, Events: dto.Events, Enabled: dto.Enabled,
	})
	if err != nil {
		http.Error(w, "db error", http.StatusInternalServerError)
		return
	}

	util.WriteJSON(w, map[string]any{"id": id})
}

// update godoc
// @Summary      Update webhook
// @Description  Update an existing webhook configuration
// @Tags         webhooks
// @Accept       json
// @Produce      json
// @Param        id       path      int                 true  "Webhook ID"
// @Param        webhook  body      webhook.WebhookDTO  true  "Updated webhook configuration"
// @Success      200      {object}  map[string]bool     "Update confirmation"
// @Failure      400      {string}  string              "Invalid JSON or webhook ID"
// @Failure      401      {string}  string              "Unauthorized"
// @Failure      500      {string}  string              "Database error"
// @Security     ApiKeyAuth
// @Security     BearerAuth
// @Router       /webhooks/{id} [put]
func (h *WebhookHandler) update(w http.ResponseWriter, r *http.Request, id int64) {
	var dto webhook.WebhookDTO
	if err := json.NewDecoder(r.Body).Decode(&dto); err != nil {
		http.Error(w, "bad json", http.StatusBadRequest)
		return
	}

	if err := h.Repo.Update(id, webhook.Webhook{
		URL: dto.URL, Events: dto.Events, Enabled: dto.Enabled,
	}); err != nil {
		http.Error(w, "db error", http.StatusInternalServerError)
		return
	}

	util.WriteJSON(w, map[string]any{"updated": true})
}

// delete godoc
// @Summary      Delete webhook
// @Description  Remove a webhook from the registry
// @Tags         webhooks
// @Produce      json
// @Param        id   path      int              true  "Webhook ID"
// @Success      200  {object}  map[string]bool  "Deletion confirmation"
// @Failure      400  {string}  string           "Invalid webhook ID"
// @Failure      401  {string}  string           "Unauthorized"
// @Failure      500  {string}  string           "Database error"
// @Security     ApiKeyAuth
// @Security     BearerAuth
// @Router       /webhooks/{id} [delete]
func (h *WebhookHandler) delete(w http.ResponseWriter, id int64) {
	if err := h.Repo.Delete(id); err != nil {
		http.Error(w, "db error", http.StatusInternalServerError)
		return
	}
	util.WriteJSON(w, map[string]any{"deleted": true})
}
