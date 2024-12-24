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
        </div>
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
import {getId, getToken} from "../store/auth";
import {initializeWebSocket} from "../services/socket";
import {getProfileImage} from "../store/helpers";

export default {
  data() {
    return {
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
    };
  },
  beforeDestroy() {
    if (this.socket) {
      this.socket.close();
    }
  },
  methods: {
    getProfileImage,
    initializeSocket() {
      if (this.groupInfo && this.groupInfo.ID) {
        this.socket = initializeWebSocket(this, this.groupInfo.ID);
      } else {
        this.socket = initializeWebSocket(this, 0);
      }
    },
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
          message: message.message,
          isPhoto: message.is_photo,
          isSent: message.user_id == getId(),
          username: message.username ?? "",
          createdAt: message.createdAt ?? ""
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
    sendMessage() {
      if (this.newMessage.trim()) {
        const isGroup = !!this.groupInfo.ID;
        const messageToSend = {
          message: this.newMessage,
          isSent: true,
          isGroup: isGroup,
          createdAt: Date(),
        };

        this.messages.push(messageToSend);

        this.newMessage = "";

        this.sendMessageToBackend(messageToSend);

        this.$nextTick(() => {
          const chatMessages = this.$refs.chatMessages;
          chatMessages.scrollTop = chatMessages.scrollHeight;
        });
      }
    },
    async sendMessageToBackend(message) {
      try {
        const response = await this.$axios.post('/send-message', {
          text: message.message,
          toUserId: this.userInfo.ID,
          isGroup:message.isGroup,
          groupId:this.groupInfo.ID
        }, {
          headers: {
            'Authorization': `Bearer ${getToken()}`
          },
        });

        console.log('Message sent:', response.data);
      } catch (error) {
        console.error('Error sending message:', error);
        this.messages.pop();
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

        const messageToSend = {
          message: this.photoPath,
          isSent: true,
          isGroup: isGroup,
          createdAt: Date(),
          isPhoto: true
        };

        this.messages.push(messageToSend);

        this.$nextTick(() => {
          const chatMessages = this.$refs.chatMessages;
          chatMessages.scrollTop = chatMessages.scrollHeight;
        });
      } catch (error) {
        console.error("Error creating group:", error);
      }
    },
  },
  mounted() {
    if (this.$route.query.entity) {
      try {
        const entity = JSON.parse(this.$route.query.entity);
        const isGroup = this.$route.query.isGroup === "true";

        if (isGroup) {
          this.groupInfo = entity;
        } else {
          this.userInfo = entity;
        }
      } catch (error) {
        console.error('Error parsing conversation data:', error);
      }
    }
  },
  watch: {
    userInfo(newValue) {
      if (newValue && newValue.ID) {
        this.getMessages(newValue.ID);
        this.initializeSocket();
      }
    },
    groupInfo(newValue) {
      if (newValue && newValue.ID) {
        this.getMessages(newValue.ID);
        this.initializeSocket();
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
  font-size: 0.8em;
  color: gray;
  text-align: left;
  margin-top: 5px;
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
</style>
