<template xmlns="http://www.w3.org/1999/html">
  <link href="https://cdnjs.cloudflare.com/ajax/libs/font-awesome/6.0.0-beta3/css/all.min.css" rel="stylesheet">

  <div class="conversation-page">
    <!-- Header -->
    <header class="chat-header">
      <div class="back-button" @click="goBack">
        ←
      </div>
      <div class="chat-title">
        {{ userInfo.Username ?? groupInfo.Name }}
      </div>
      <div v-if="groupInfo && groupInfo.Name">
        <div class="add-user-button" @click="addUsers">
          +
        </div>
      </div>
    </header>

    <!-- Chat messages container -->
    <div class="chat-messages" ref="chatMessages">
      <div
          v-for="(message, index) in messages"
          :key="index"
          :class="['chat-message', message.isSent ? 'sent' : 'received']"
          @contextmenu.prevent="openContextMenu($event, message, index)"
      >
        <!-- Name at the top -->
        <div class="message-header">
          <span class="message-sender">{{ message.username }}</span>
        </div>

        <!-- Message content -->
        <div class="message-content">
          <div v-if="message.isPhoto">
            <img
                :src="getProfileImage(message.message)"
                alt="Sent Photo"
            />
          </div>
          <div v-else>
            {{ message.message }}
          </div>
        </div>

        <!-- Time at the bottom left -->
        <div class="message-footer">
          <span class="message-time">{{ formatTime(message.createdAt) }}</span>
          <span>{{message.emoji}}</span>
          
          <div v-if="message.isRead">
            <i class="fas fa-check-double check-icon"></i>
          </div>
          <div v-else-if="message.isReceived">
            <i class="fas fa-check check-icon"></i>
          </div>
        </div>
      </div>
    </div>

    <div
        v-if="showContextMenu"
        class="context-menu"
        :style="{ top: contextMenuPosition.y + 'px', left: contextMenuPosition.x + 'px' }"
    >
      <ul>
        <li @click="deleteMessage(selectedMessage)">Delete</li>
        <li @click="commentMessage(selectedMessage)">Comment</li>
        <li @click="forwardMessage(selectedMessage)">Forward</li>
      </ul>
    </div>

    <div v-if="showEmojiPicker" class="emoji-picker">
      <div v-for="emoji in emojis" :key="emoji" class="emoji" @click="selectEmoji(emoji, selectedMessage.id)">
        {{ emoji }}
      </div>
    </div>

    <!-- Input box -->
    <div class="chat-input">
      <input
          v-model="newMessage"
          type="text"
          placeholder="Type a message..."
          @keyup.enter="sendMessage"
      />
      <button @click="sendMessage">Send</button>
      <input
          type="file"
          ref="fileInput"
          @change="uploadAndSendPhoto"
          accept="image/*"
          style="display: none;"
      />
      <button @click="triggerFileInput">
        <i class="fas fa-paperclip"></i>
      </button>
    </div>

    <div v-if="isModalVisible" class="modal">
      <form @submit.prevent="addNewUsers">

        <div class="modal-content">
          <h3>Invite Users</h3>
          <div class="user-finder">
            <input
                type="text"
                v-model="searchQuery"
                placeholder="Search for a user to start a conversation"
                class="search-bar"
                @input="findUsers"
            />
          </div>

          <div v-if="userSearchResult.length" class="search-results">
            <ul>
              <li v-for="user in userSearchResult" :key="user.ID" class="search-item">
                <img :src="getProfileImage(user.ProfilePhotoURL)" alt="Profile" class="profile-photo" />
                <div class="user-details">
                  <h5>{{ user.Username }}</h5>
                  <button type="button" @click="addUserToGroup(user)">Add to Group</button>
                </div>
              </li>
            </ul>
          </div>

          <div v-if="selectedUsers.length" class="selected-users">
            <h4>Selected Users</h4>
            <ul>
              <li v-for="user in selectedUsers" :key="user.ID" class="selected-user-item">
                <img :src="getProfileImage(user.ProfilePhotoURL)" alt="Profile" class="profile-photo" />
                <div class="user-details">
                  <h5>{{ user.Username }}</h5>
                  <button type="button" @click="removeUserFromGroup(user)">Remove</button>
                </div>
              </li>
            </ul>
          </div>

          <button @click="closeModal">Close</button>

          <button type="submit">Update Group</button>
        </div>
      </form>
    </div>
  </div>
</template>

<script>
import {getId, getToken, getUsername} from "../store/auth";
import {getProfileImage} from "../store/helpers";

export default {
  data() {
    return {
      authUsername:"",
      photoPath:"",
      socket: null,
      userInfo: {},
      groupInfo: {},
      messages: [],
      newMessage: "",
      isModalVisible: false,
      searchQuery: '',
      userSearchResult: [],
      selectedUsers: [],
      showContextMenu: false,
      contextMenuPosition: { x: 0, y: 0 },
      selectedMessage: null,
      showEmojiPicker: false,
      emojis: ["😀", "😂", "😍", "😎", "😢", "😡", "👍", "👎", "🔥", "❤️", "💯"],
      selectedEmoji: "",
      selectedIndex: null,
    };
  },
  beforeDestroy() {
    if (this.socket) {
      this.socket.close();
    }
  },
  methods: {
    getProfileImage,
    formatTime(timestamp) {
      const date = new Date(timestamp);
      return date.toLocaleTimeString([], { hour: "2-digit", minute: "2-digit" });
    },
    goBack() {
      this.$router.push("/");
    },
    async getMessages(id) {
      const isGroup = this.groupInfo.ID === id;

      try {
        let response = await this.$axios.post('/get-messages', {
          id: id,
          isGroup: isGroup
        }, {
          headers: {
            'Authorization': `Bearer ${getToken()}`
          }
        });
        this.messages = response.data['messages'].map(message => ({
          id: message.message_id,
          message: message.message,
          isPhoto: message.is_photo,
          isSent: message.user_id == getId(),
          username: message.username ?? "",
          createdAt: message.createdAt ?? "",
          isReceived:true,
          isRead:message.is_read,
          emoji:message.emoji,
        }));
        console.log(this.messages)
      } catch (error) {
        if (error.response) {
          if (error.response.status === 401) {
            console.error('Unauthorized: Token may be invalid or expired.');
            localStorage.removeItem('authToken');
          } else {
            console.error('Error fetching messages:', error.response.data);
          }
        } else if (error.request) {
          console.error('No response received:', error.request);
        } else {
          console.error('Error setting up request:', error.message);
        }
      }
    },
    async sendMessage() {
      if (this.newMessage.trim()) {
        const isGroup = !!this.groupInfo.ID;
        let isReceived = false;

        try {
          const response = await this.$axios.post('/send-message', {
            text: this.newMessage,
            toUserId: this.userInfo.ID,
            isGroup:isGroup,
            groupId:this.groupInfo.ID
          }, {
            headers: {
              'Authorization': `Bearer ${getToken()}`
            },
          });

          if (response.data) {
            isReceived = true
          }
          console.log('Message sent:', response.data);
        } catch (error) {
          console.error('Error sending message:', error);
          this.messages.pop();
        }



        const messageToSend = {
          username:this.authUsername,
          message: this.newMessage,
          isSent: true,
          isGroup: isGroup,
          createdAt: Date(),
          isReceived:isReceived,
        };

        this.messages.push(messageToSend);

        this.newMessage = "";


        this.$nextTick(() => {
          const chatMessages = this.$refs.chatMessages;
          chatMessages.scrollTop = chatMessages.scrollHeight;
        });
      }
    },
    addUserToGroup(user) {
      if (!this.selectedUsers.some(u => u.ID === user.ID)) {
        this.selectedUsers.push(user);
      }
    },
    removeUserFromGroup(user) {
      this.selectedUsers = this.selectedUsers.filter(u => u.ID !== user.ID);
    },
    addUsers() {
      this.isModalVisible = true;
    },
    closeModal() {
      this.isModalVisible = false;
    },
    async findUsers() {
      if (!this.searchQuery) {
        this.userSearchResult = [];
        return;
      }

      try {
        let response = await this.$axios.get(`/get-users?search=${this.searchQuery}`, {
          headers: {
            'Authorization': `Bearer ${getToken()}`
          }
        });

        this.userSearchResult = response.data.users;
      } catch (error) {
        console.error("Error fetching users:", error);
        this.userSearchResult = [];
      }
    },
    async addNewUsers() {
      try {
        const response = await this.$axios.post(
            "/update-group",
            {
              groupId: this.groupInfo.ID,
              users: this.selectedUsers
            },
            {
              headers: {
                Authorization: `Bearer ${getToken()}`,
              },
            }
        );
        this.isModalVisible = false;
        alert("User added")
      } catch (error) {
        console.error("Error add users to group:", error);
      }
    },
    triggerFileInput() {
      this.$refs.fileInput.click();
    },
    async uploadAndSendPhoto(event) {
      const file = event.target.files[0];
      const isGroup = !!this.groupInfo.ID;

      if (!file) {
        alert('No file selected!');
        return;
      }

      const formData = new FormData();

      formData.append('file', file);
      formData.append('isGroup', isGroup);
      formData.append('groupId', this.groupInfo.ID);
      formData.append('toUserId', this.userInfo.ID);

      try {
        const response = await this.$axios.post("/send-photo", formData, {
          headers: {
            Authorization: `Bearer ${getToken()}`,
            "Content-Type": "multipart/form-data",
          },
        });

        this.photoPath = response.data['photoPath']
      } catch (error) {
        console.error("Error creating group:", error);
      }
    },
    markAsRead() {
      const isGroup = !!this.groupInfo.ID;

      try {
        const response = this.$axios.post(
            "/mark-as-read",
            {
              groupId: this.groupInfo.ID,
              toUserId: this.userInfo.ID,
              isGroup:isGroup,
            },
            {
              headers: {
                Authorization: `Bearer ${getToken()}`,
              },
            }
        );
      } catch (error) {
        console.error("Error creating group:", error);
      }
    },
    openContextMenu(event, message, index) {
      this.selectedIndex = index;
      event.preventDefault();
      this.showContextMenu = true;
      this.contextMenuPosition = { x: event.clientX, y: event.clientY };
      this.selectedMessage = message;

      document.addEventListener("click", this.closeContextMenu);
    },
    closeContextMenu() {
      this.showContextMenu = false;
      document.removeEventListener("click", this.closeContextMenu);
    },
    async deleteMessage(message) {
      try {
        const response = await this.$axios.post(
            "/delete-message",
            {
              id: message.id,
            },
            {
              headers: {
                Authorization: `Bearer ${getToken()}`,
              },
            }
        );

        if(response.data) {
          alert("Message deleted")
          window.location.reload()
        }

      } catch (error) {
        console.error("Error creating group:", error);
      }
    },
    commentMessage(message) {
      this.showEmojiPicker = true;
      this.showContextMenu = false;
    },
    forwardMessage(message) {
      this.showContextMenu = false;
      const recipient = prompt("Enter recipient username or ID:");
      if (recipient) {
        this.$axios.post(`/api/messages/${message.id}/forward`, { recipient })
            .then(() => {
              console.log("Message forwarded");
            })
            .catch(error => {
              console.error("Failed to forward message:", error);
            });
      }
    },
    selectEmoji(emoji, messageId) {
      console.log("Selected emoji:", emoji, messageId);
      this.selectedEmoji = emoji;
      this.showEmojiPicker = false;

      try {
        const response = this.$axios.post(
            "/comment-message",
            {
              messageId: messageId,
              emoji: emoji,
            },
            {
              headers: {
                Authorization: `Bearer ${getToken()}`,
              },
            }
        );
      } catch (error) {
        console.error("Error creating group:", error);
      }
    },
    getAuthUsername() {
      this.authUsername = getUsername();
    }
  },
  mounted() {
    if (this.$route.query.entity) {
      this.getAuthUsername();
      try {
        const entity = JSON.parse(this.$route.query.entity);
        const isGroup = this.$route.query.isGroup === "true";

        if (isGroup) {
          this.groupInfo = entity;
        } else {
          this.userInfo = entity;
        }

        this.markAsRead();
      } catch (error) {
        console.error('Error parsing conversation data:', error);
      }
    }
  },
  watch: {
    userInfo(newValue) {
      if (newValue && newValue.ID) {
        this.getMessages(newValue.ID);
      }
    },
    groupInfo(newValue) {
      if (newValue && newValue.ID) {
        this.getMessages(newValue.ID);
      }
    },
  },
};
</script>

<style scoped>
/* General styles */
.conversation-page {
  display: flex;
  flex-direction: column;
  height: 100vh;
  background-color: #f5f8fb;
}

/* Header styles */
.chat-header {
  display: flex;
  flex-direction: row;
  align-items: center;
  z-index: 1;
  background-color: #0088cc;
  color: #fff;
  font-size: 25px;
}

.back-button {
  cursor: pointer;
  margin-right: 15px;
  font-size: 25px;
  font-weight: bold;
}

.chat-title {
  flex-grow: 1;
  text-align: center;
  padding: 5px;
}

/* Messages styles */
.chat-messages {
  flex: 1;
  overflow-y: auto;
  padding: 15px;
  display: flex;
  flex-direction: column;
}

/* Input box styles */
.chat-input {
  display: flex;
  padding: 10px;
  border-top: 1px solid #ddd;
  background-color: #fff;
  position: sticky;
  bottom: 0;
}

.chat-input input {
  flex: 1;
  padding: 10px;
  border: 1px solid #ccc;
  border-radius: 20px;
  outline: none;
}

.chat-input button {
  background-color: #0088cc;
  color: white;
  border: none;
  border-radius: 20px;
  padding: 10px 15px;
  margin-left: 10px;
  cursor: pointer;
}

.chat-input button:hover {
  background-color: #005f99;
}

.chat-message {
  padding: 10px;
  margin: 5px 0;
  border-radius: 8px;
  max-width: 60%;
  word-wrap: break-word;
  position: relative;
}

.chat-message.sent {
  background-color: #d1e7dd;
  margin-left: auto;
  text-align: right;
}

.chat-message.received {
  background-color: #f8d7da;
  margin-right: auto;
}

.message-header {
  font-weight: bold;
  margin-bottom: 5px;
}

.message-footer {
  display: flex;
  align-items: center;
  justify-content: space-between; /* Размещаем элементы по краям */
  padding: 5px 10px;
  font-size: 12px;
  color: gray;
}

.check-icon {
  color: gray;          /* Цвет иконки */
  margin-left: 5px;     /* Отступ слева */
  font-size: 14px;      /* Размер иконки */
}

.modal {
  position: fixed;
  top: 0;
  left: 0;
  width: 100%;
  height: 100%;
  background: rgba(0, 0, 0, 0.5);
  display: flex;
  justify-content: center;
  align-items: center;
  z-index: 1000;
}

.modal-content {
  background: white;
  padding: 20px;
  border-radius: 8px;
  max-width: 500px;
  width: 100%;
}

.search-results, .selected-users {
  margin-top: 20px;
}

.context-menu {
  position: absolute;
  z-index: 1000;
  background: #fff;
  border: 1px solid #ddd;
  box-shadow: 0 2px 10px rgba(0, 0, 0, 0.1);
  border-radius: 4px;
}

.context-menu ul {
  list-style: none;
  padding: 0;
  margin: 0;
}

.context-menu ul li {
  padding: 8px 12px;
  cursor: pointer;
  transition: background 0.2s;
}

.context-menu ul li:hover {
  background: #f5f5f5;
}
</style>
