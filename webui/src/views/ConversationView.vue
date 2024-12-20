<template>
  <div class="conversation-page">
    <!-- Header -->
    <header class="chat-header">
      <div class="back-button" @click="goBack">
        ←
      </div>
      <div class="chat-title">
        {{ userInfo.Username ?? groupInfo.Name }}
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
          {{ message.message }}
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
    </div>
  </div>
</template>

<script>
import {getId, getToken} from "../store/auth";
import {initializeWebSocket} from "../services/socket";

export default {
  data() {
    return {
      socket: null,
      userInfo: {},
      groupInfo: {},
      messages: [],
      newMessage: "",
    };
  },
  beforeDestroy() {
    if (this.socket) {
      this.socket.close();
    }
  },
  methods: {
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
          isSent: message.user_id == getId(),
          username: message.username ?? "",
          createdAt: message.createdAt ?? ""
        }));

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
</style>
