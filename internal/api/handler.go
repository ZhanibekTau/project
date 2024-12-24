package api

import (
	"encoding/json"
	"net/http"
	"project/cmd/database/model"
	"project/internal/helpers"
	"project/internal/service"
	"strconv"
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
	userId := r.Context().Value("userId").(uint)

	err := r.ParseMultipartForm(10 << 20) // Limit file size to 10MB
	if err != nil {
		helpers.HandleError(w, helpers.NewAPIError("Unable to parse form data", http.StatusUnprocessableEntity))
		return
	}

	groupName := r.FormValue("groupName")
	if groupName == "" {
		helpers.HandleError(w, helpers.NewAPIError("Group name is required", http.StatusBadRequest))
		return
	}

	selectedUsers := r.FormValue("selectedUsers")
	if selectedUsers == "" {
		helpers.HandleError(w, helpers.NewAPIError("Selected users are required", http.StatusBadRequest))
		return
	}

	var users []model.User
	err = json.Unmarshal([]byte(selectedUsers), &users)
	if err != nil {
		helpers.HandleError(w, helpers.NewAPIError("Invalid selected users data", http.StatusBadRequest))
		return
	}

	file, header, err := r.FormFile("profilePhoto")
	if err != nil && err != http.ErrMissingFile {
		helpers.HandleError(w, helpers.NewAPIError("Error retrieving file", http.StatusBadRequest))
		return
	}

	filepath, err := helpers.SaveUploadedFile(file, header, "webui/public/images/group", userId)
	if err != nil {
		helpers.HandleError(w, helpers.NewAPIError(err.Error(), http.StatusBadRequest))
		return
	}

	result, err := h.services.CreateGroups(&helpers.CreateGroupRequest{
		GroupName:      groupName,
		GroupPhotoPath: filepath,
		Users:          users,
	}, userId)
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
	message := map[string]interface{}{}

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
				if input.PhotoPath != "" {
					message = map[string]interface{}{
						"message":   input.PhotoPath,
						"isPhoto":   true,
						"sender":    userId,
						"groupId":   input.GroupId,
						"username":  fromUsername,
						"createdAt": time.Now(),
					}
				} else {
					message = map[string]interface{}{
						"message":   input.Text,
						"isPhoto":   false,
						"sender":    userId,
						"groupId":   input.GroupId,
						"username":  fromUsername,
						"createdAt": time.Now(),
					}
				}

				err = memberConn.WriteJSON(message)
				if err != nil {
					// Если отправка не удалась, логируем ошибку
					delete(connections, member.UserID)
				}
			}
		}
	} else {
		receiverConn, exists := connections[input.ToUserId]
		if exists {
			if input.PhotoPath != "" {
				message = map[string]interface{}{
					"message":   input.PhotoPath,
					"isPhoto":   true,
					"sender":    userId,
					"createdAt": time.Now(),
				}
			} else {
				message = map[string]interface{}{
					"message":   input.Text,
					"isPhoto":   false,
					"sender":    userId,
					"createdAt": time.Now(),
				}
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

func (h *Handler) uploadProfilePicture(w http.ResponseWriter, r *http.Request) {
	userId := r.Context().Value("userId").(uint) // Получение ID пользователя из контекста

	// Парсим загруженный файл
	file, header, err := r.FormFile("profile_picture")
	if err != nil {
		helpers.HandleError(w, helpers.NewAPIError(err.Error(), http.StatusUnprocessableEntity))
		return
	}
	defer file.Close()

	filepath, err := helpers.SaveUploadedFile(file, header, "webui/public/images/profile", userId)
	if err != nil {
		helpers.HandleError(w, helpers.NewAPIError(err.Error(), http.StatusUnprocessableEntity))
		return
	}

	result, err := h.services.UpdateUserProfile(userId, filepath)
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

func (h *Handler) leaveGroup(w http.ResponseWriter, r *http.Request) {
	var input helpers.GroupRequest
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		helpers.HandleError(w, helpers.NewAPIError(err.Error(), http.StatusUnprocessableEntity))
		return
	}

	userId := r.Context().Value("userId").(uint)

	result, err := h.services.LeaveGroup(userId, &input)
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

func (h *Handler) addUsersToGroup(w http.ResponseWriter, r *http.Request) {
	var input helpers.AddUsersToGroup
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		helpers.HandleError(w, helpers.NewAPIError(err.Error(), http.StatusUnprocessableEntity))
		return
	}

	userId := r.Context().Value("userId").(uint)

	result, err := h.services.AddUsersToGroup(userId, &input)
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

func (h *Handler) sendPhoto(w http.ResponseWriter, r *http.Request) {
	userId, ok := r.Context().Value("userId").(uint)
	if !ok {
		helpers.HandleError(w, helpers.NewAPIError("Invalid user ID", http.StatusUnauthorized))
		return
	}

	if err := r.ParseMultipartForm(10 << 20); err != nil {
		helpers.HandleError(w, helpers.NewAPIError("Unable to parse form data", http.StatusUnprocessableEntity))
		return
	}

	isGroup, err := strconv.ParseBool(r.FormValue("isGroup"))
	if err != nil {
		helpers.HandleError(w, helpers.NewAPIError("Invalid 'isGroup' value", http.StatusUnprocessableEntity))
		return
	}

	groupId, err := parseOptionalInt(r.FormValue("groupId"))
	if err != nil {
		helpers.HandleError(w, helpers.NewAPIError("Invalid 'groupId' value", http.StatusUnprocessableEntity))
		return
	}

	toUserId, err := parseOptionalInt(r.FormValue("toUserId"))
	if err != nil {
		helpers.HandleError(w, helpers.NewAPIError("Invalid 'toUserId' value", http.StatusUnprocessableEntity))
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil && err != http.ErrMissingFile {
		helpers.HandleError(w, helpers.NewAPIError("Error retrieving file", http.StatusBadRequest))
		return
	}

	filepath, err := helpers.SaveUploadedFile(file, header, "webui/public/images/conversation", userId)
	if err != nil {
		helpers.HandleError(w, helpers.NewAPIError("Error saving file", http.StatusBadRequest))
		return
	}

	payload := helpers.SendMessageRequest{
		ToUserId:  uint(toUserId),
		IsGroup:   isGroup,
		GroupId:   uint(groupId),
		PhotoPath: filepath,
	}

	result, err := h.services.SendMessage(userId, &payload)
	if err != nil {
		helpers.HandleError(w, helpers.NewAPIError("Failed to send message", http.StatusBadRequest))
		return
	}

	if err = h.webSocket(&payload, userId); err != nil {
		helpers.HandleError(w, helpers.NewAPIError("Failed to send WebSocket notification", http.StatusBadRequest))
		return
	}
	var res map[string]string

	if result {
		res = map[string]string{
			"photoPath": filepath,
		}
	}

	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(res)
	if err != nil {
		helpers.HandleError(w, helpers.NewAPIError(err.Error(), http.StatusBadRequest))
		return
	}
}

func parseOptionalInt(value string) (int, error) {
	if value == "undefined" || value == "" {
		return 0, nil
	}
	return strconv.Atoi(value)
}
