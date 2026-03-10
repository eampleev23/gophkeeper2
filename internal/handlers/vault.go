package handlers

import (
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"errors"
	"github.com/eampleev23/gophkeeper2.git/internal/auth"
	"github.com/eampleev23/gophkeeper2.git/internal/models"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
	"io"
	"net/http"
	"strconv"
)

func getUserIDFromRequest(r *http.Request) (int, bool) {
	v := r.Context().Value(auth.KeyUserIDCtx)
	if v == nil {
		return 0, false
	}
	userID, ok := v.(int)
	return userID, ok
}

// GetUserKeys возвращает зашифрованный ключевой блоб пользователя (тело ответа — сырые байты).
func (h *Handlers) GetUserKeys(w http.ResponseWriter, r *http.Request) {
	userID, ok := getUserIDFromRequest(r)
	if !ok || userID == 0 {
		sendResponse(true, "Unauthorized", http.StatusUnauthorized, w)
		return
	}
	blob, err := h.store.GetUserEncryptedKeys(r.Context(), userID)
	if err != nil {
		h.logger.ZL.Error("GetUserEncryptedKeys failed", zap.Error(err))
		sendResponse(true, "Internal server error", http.StatusInternalServerError, w)
		return
	}
	if blob == nil {
		w.WriteHeader(http.StatusNoContent)
		return
	}
	w.Header().Set("Content-Type", "application/octet-stream")
	w.Write(blob)
}

// SetUserKeys сохраняет зашифрованный ключевой блоб (тело запроса — сырые байты).
func (h *Handlers) SetUserKeys(w http.ResponseWriter, r *http.Request) {
	userID, ok := getUserIDFromRequest(r)
	if !ok || userID == 0 {
		sendResponse(true, "Unauthorized", http.StatusUnauthorized, w)
		return
	}
	blob, err := io.ReadAll(r.Body)
	if err != nil {
		sendResponse(true, "Bad request", http.StatusBadRequest, w)
		return
	}
	if len(blob) == 0 {
		sendResponse(true, "Empty body", http.StatusBadRequest, w)
		return
	}
	if err := h.store.SetUserEncryptedKeys(r.Context(), userID, blob); err != nil {
		h.logger.ZL.Error("SetUserEncryptedKeys failed", zap.Error(err))
		sendResponse(true, "Internal server error", http.StatusInternalServerError, w)
		return
	}
	sendResponse(false, "OK", http.StatusOK, w)
}

// CreateVaultItemRequest — тело запроса на создание записи.
type CreateVaultItemRequest struct {
	Type     string `json:"type"`
	MetaName string `json:"meta_name"`
	Payload  string `json:"payload"` // base64
}

// CreateVaultItem создаёт запись хранилища.
func (h *Handlers) CreateVaultItem(w http.ResponseWriter, r *http.Request) {
	userID, ok := getUserIDFromRequest(r)
	if !ok || userID == 0 {
		sendResponse(true, "Unauthorized", http.StatusUnauthorized, w)
		return
	}
	var req CreateVaultItemRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendResponse(true, "Invalid JSON", http.StatusBadRequest, w)
		return
	}
	if req.Type == "" {
		sendResponse(true, "type is required", http.StatusBadRequest, w)
		return
	}
	payload, err := base64.StdEncoding.DecodeString(req.Payload)
	if err != nil {
		sendResponse(true, "payload must be base64", http.StatusBadRequest, w)
		return
	}
	id, err := h.store.CreateVaultItem(r.Context(), userID, req.Type, req.MetaName, payload)
	if err != nil {
		h.logger.ZL.Error("CreateVaultItem failed", zap.Error(err))
		sendResponse(true, "Internal server error", http.StatusInternalServerError, w)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(map[string]int64{"id": id})
}

// ListVaultItems возвращает список метаданных записей пользователя.
func (h *Handlers) ListVaultItems(w http.ResponseWriter, r *http.Request) {
	userID, ok := getUserIDFromRequest(r)
	if !ok || userID == 0 {
		sendResponse(true, "Unauthorized", http.StatusUnauthorized, w)
		return
	}
	list, err := h.store.ListVaultItems(r.Context(), userID)
	if err != nil {
		h.logger.ZL.Error("ListVaultItems failed", zap.Error(err))
		sendResponse(true, "Internal server error", http.StatusInternalServerError, w)
		return
	}
	if list == nil {
		list = []models.VaultItemMeta{}
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(list)
}

// GetVaultItem возвращает одну запись (включая payload в base64).
func (h *Handlers) GetVaultItem(w http.ResponseWriter, r *http.Request) {
	userID, ok := getUserIDFromRequest(r)
	if !ok || userID == 0 {
		sendResponse(true, "Unauthorized", http.StatusUnauthorized, w)
		return
	}
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || id <= 0 {
		sendResponse(true, "Invalid id", http.StatusBadRequest, w)
		return
	}
	item, err := h.store.GetVaultItem(r.Context(), userID, id)
	if err != nil {
		h.logger.ZL.Error("GetVaultItem failed", zap.Error(err))
		sendResponse(true, "Internal server error", http.StatusInternalServerError, w)
		return
	}
	if item == nil {
		sendResponse(true, "Not found", http.StatusNotFound, w)
		return
	}
	out := struct {
		ID        int64  `json:"id"`
		Type      string `json:"type"`
		MetaName  string `json:"meta_name"`
		Payload   string `json:"payload"` // base64
		CreatedAt string `json:"created_at"`
		UpdatedAt string `json:"updated_at"`
	}{
		ID:        item.ID,
		Type:      item.Type,
		MetaName:  item.MetaName,
		Payload:   base64.StdEncoding.EncodeToString(item.Payload),
		CreatedAt: item.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt: item.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(out)
}

// UpdateVaultItemRequest — тело запроса на обновление записи.
type UpdateVaultItemRequest struct {
	MetaName string `json:"meta_name"`
	Payload  string `json:"payload"` // base64
}

// UpdateVaultItem обновляет запись.
func (h *Handlers) UpdateVaultItem(w http.ResponseWriter, r *http.Request) {
	userID, ok := getUserIDFromRequest(r)
	if !ok || userID == 0 {
		sendResponse(true, "Unauthorized", http.StatusUnauthorized, w)
		return
	}
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || id <= 0 {
		sendResponse(true, "Invalid id", http.StatusBadRequest, w)
		return
	}
	var req UpdateVaultItemRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendResponse(true, "Invalid JSON", http.StatusBadRequest, w)
		return
	}
	payload, err := base64.StdEncoding.DecodeString(req.Payload)
	if err != nil {
		sendResponse(true, "payload must be base64", http.StatusBadRequest, w)
		return
	}
	if err := h.store.UpdateVaultItem(r.Context(), userID, id, req.MetaName, payload); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			sendResponse(true, "Not found", http.StatusNotFound, w)
			return
		}
		h.logger.ZL.Error("UpdateVaultItem failed", zap.Error(err))
		sendResponse(true, "Internal server error", http.StatusInternalServerError, w)
		return
	}
	sendResponse(false, "OK", http.StatusOK, w)
}

// DeleteVaultItem удаляет запись.
func (h *Handlers) DeleteVaultItem(w http.ResponseWriter, r *http.Request) {
	userID, ok := getUserIDFromRequest(r)
	if !ok || userID == 0 {
		sendResponse(true, "Unauthorized", http.StatusUnauthorized, w)
		return
	}
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || id <= 0 {
		sendResponse(true, "Invalid id", http.StatusBadRequest, w)
		return
	}
	if err := h.store.DeleteVaultItem(r.Context(), userID, id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			sendResponse(true, "Not found", http.StatusNotFound, w)
			return
		}
		h.logger.ZL.Error("DeleteVaultItem failed", zap.Error(err))
		sendResponse(true, "Internal server error", http.StatusInternalServerError, w)
		return
	}
	sendResponse(false, "OK", http.StatusOK, w)
}
