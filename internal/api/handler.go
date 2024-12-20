package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"project/cmd/database/model"
	"project/internal/helpers"
	"project/internal/service"
	"time"
)

type Handler struct {
	services *service.Service
}

func NewHandler(services *service.Service) *Handler {
	return &Handler{services: services}
}

func (h *Handler) doLogin(w http.ResponseWriter, r *http.Request) {
	var input model.User
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		helpers.HandleError(w, helpers.NewAPIError(err.Error(), http.StatusBadRequest))
		return
	}

	user, err := h.services.CreateOrGetUser(&input)
	if err != nil {
		helpers.HandleError(w, helpers.NewAPIError("Failed to retrieve token", http.StatusBadRequest))
		return
	}

	token, err := helpers.GenerateSessionToken(user)
	if err != nil {
		helpers.HandleError(w, helpers.NewAPIError("Failed to retrieve token", http.StatusBadRequest))
		return
	}

	res := map[string]interface{}{
		"token": token,
		"id":    user.ID,
	}

	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(res)
	if err != nil {
		helpers.HandleError(w, helpers.NewAPIError("Failed to encode response", http.StatusBadRequest))
		return
	}
}

func (h *Handler) getConversations(w http.ResponseWriter, r *http.Request) {
	userId := r.Context().Value("userId").(uint)

	info, err := h.services.GetConversations(userId)
	if err != nil {
		helpers.HandleError(w, helpers.NewAPIError(err.Error(), http.StatusBadRequest))
		return
	}

	res := map[string]interface{}{
		"result": info,
	}

	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(res)
	if err != nil {
		helpers.HandleError(w, helpers.NewAPIError(err.Error(), http.StatusBadRequest))
		return
	}
}

func (h *Handler) getUsers(w http.ResponseWriter, r *http.Request) {
	userId := r.Context().Value("userId").(uint)

	query := r.URL.Query().Get("search")
	if query == "" {
		helpers.HandleError(w, helpers.NewAPIError("Search query is required", http.StatusBadRequest))
		return
	}

	users, err := h.services.GetUsers(query, userId)
	if err != nil {
		helpers.HandleError(w, helpers.NewAPIError(err.Error(), http.StatusBadRequest))
		return
	}

	res := map[string][]model.User{"users": *users}

	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(res)
	if err != nil {
		helpers.HandleError(w, helpers.NewAPIError(err.Error(), http.StatusBadRequest))
		return
	}
}

func (h *Handler) getMessages(w http.ResponseWriter, r *http.Request) {
	var convUserId helpers.GetMessagesRequest
	if err := json.NewDecoder(r.Body).Decode(&convUserId); err != nil {
		helpers.HandleError(w, helpers.NewAPIError(err.Error(), http.StatusBadRequest))
		return
	}

	userId := r.Context().Value("userId").(uint)

	messages, err := h.services.GetMessages(userId, &convUserId)
	if err != nil {
		helpers.HandleError(w, helpers.NewAPIError(err.Error(), http.StatusBadRequest))
		return
	}

	res := map[string][]helpers.MessagesResponse{"messages": *messages}

	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(res)
	if err != nil {
		helpers.HandleError(w, helpers.NewAPIError(err.Error(), http.StatusBadRequest))
		return
	}
}

func (h *Handler) sendMessage(w http.ResponseWriter, r *http.Request) {
	var input helpers.SendMessageRequest
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		helpers.HandleError(w, helpers.NewAPIError(err.Error(), http.StatusUnprocessableEntity))
		return
	}

	userId := r.Context().Value("userId").(uint)

	result, err := h.services.SendMessage(userId, &input)
	if err != nil {
		helpers.HandleError(w, helpers.NewAPIError(err.Error(), http.StatusBadRequest))
		return
	}

	err = h.webSocket(&input, userId)
	if err != nil {
		helpers.HandleError(w, helpers.NewAPIError(err.Error(), http.StatusBadRequest))
		return
	}

	res := map[string]bool{"success": result}

	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(res)
	if err != nil {
		helpers.HandleError(w, helpers.NewAPIError(err.Error(), http.StatusBadRequest))
		return
	}
}

func (h *Handler) createGroup(w http.ResponseWriter, r *http.Request) {
	var input helpers.CreateGroupRequest
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		helpers.HandleError(w, helpers.NewAPIError(err.Error(), http.StatusUnprocessableEntity))
		return
	}
	userId := r.Context().Value("userId").(uint)

	result, err := h.services.CreateGroups(&input, userId)
	if err != nil {
		helpers.HandleError(w, helpers.NewAPIError(err.Error(), http.StatusUnprocessableEntity))
		return
	}

	res := map[string]bool{"success": result}

	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(res)
	if err != nil {
		helpers.HandleError(w, helpers.NewAPIError(err.Error(), http.StatusBadRequest))
		return
	}
}

func (h *Handler) webSocket(input *helpers.SendMessageRequest, userId uint) error {
	var fromUsername string
	// Проверяем, является ли сообщение групповым
	if input.IsGroup {
		// Получаем ID участников группы
		groupMembers, err := h.services.GetGroupMembers(input.GroupId)
		if err != nil {
			return err
		}

		// Отправляем сообщение каждому участнику группы
		for _, member := range *groupMembers {
			if member.UserID == userId {
				fromUsername = member.User.Username
				continue
			}

			// Проверяем, есть ли активное WebSocket-соединение для участника
			memberConn, exists := groupConn[input.GroupId][member.UserID]
			if exists {
				fmt.Println(member.User.Username, "asdassssssssssssssss")

				message := map[string]interface{}{
					"message":   input.Text,
					"sender":    userId,
					"groupId":   input.GroupId,
					"username":  fromUsername,
					"createdAt": time.Now(),
				}
				err = memberConn.WriteJSON(message)
				if err != nil {
					// Если отправка не удалась, логируем ошибку
					delete(connections, member.UserID)
				}
			}
		}
	} else {
		// Проверяем, есть ли активное WebSocket-соединение для получателя
		receiverConn, exists := connections[input.ToUserId]
		if exists {
			message := map[string]interface{}{
				"message":   input.Text,
				"sender":    userId,
				"createdAt": time.Now(),
			}
			err := receiverConn.WriteJSON(message)
			if err != nil {
				// Если отправка не удалась, логируем ошибку
				delete(connections, input.ToUserId)
			}
		}
	}

	return nil
}
