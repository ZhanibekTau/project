<template>
  <div class="conversation-page">
    <!-- Header -->
    <header class="chat-header">
      <div class="back-button" @click="goBack">
        ←
      </div>
      <div class="chat-title">
        {{ userInfo.Username }}
      </div>
    </header>

    <!-- Chat messages container -->
    <div class="chat-messages" ref="chatMessages">
      <div
          v-for="(message, index) in messages"
          :key="index"
          :class="['chat-message', message.isSent ? 'sent' : 'received']"
      >
        {{ message.message }}
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
import {getToken} from "../store/auth";
import {initializeWebSocket} from "../services/socket";

export default {
  data() {
    return {
      socket: null,
      userInfo: {},
      messages: [],
      newMessage: "",
    };
  },
  created() {
    this.socket = initializeWebSocket(this);
  },
  beforeDestroy() {
    if (this.socket) {
      this.socket.close();
    }
  },
  methods: {
    goBack() {
      this.$router.push("/");
    },
    async getMessages(userId) {
      try {
        let response = await this.$axios.post('/get-messages', {
          userId:  userId
        }, {
          headers: {
            'Authorization': `Bearer ${getToken()}`
          }
        });

          this.messages = response.data['messages'].map(message => {
            return {
              message: message.message,
              isSent: message.user_id !== this.userInfo.ID
            };
          });
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
        const messageToSend = { message: this.newMessage, isSent: true };

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
        }, {
          headers: {
            'Authorization': `Bearer ${getToken()}`
          }
        });

        console.log('Message sent:', response.data);
      } catch (error) {
        console.error('Error sending message:', error);
        this.messages.pop();
      }
    },
  },
  mounted() {
    if (this.$route.query.user) {
      try {
        this.userInfo = JSON.parse(this.$route.query.user);
      } catch (error) {
        console.error('Error parsing user data:', error);
      }
    }
  },
  watch: {
    userInfo(newValue) {
      if (newValue && newValue.ID) {
        this.getMessages(newValue.ID);
      }
    }
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

.chat-message {
  max-width: 70%;
  padding: 10px;
  margin: 5px 0;
  border-radius: 8px;
  word-wrap: break-word;
}

.chat-message.sent {
  align-self: flex-end;
  background-color: #0088cc;
  color: white;
}

.chat-message.received {
  align-self: flex-start;
  background-color: #e5e5ea;
  color: black;
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
</style>
