<script>
import {checkLoginStatus} from "../store/auth";

export default {
  data: function() {
    return {
      errormsg: null,
      loading: false,
      username: "",
      password: "",
    }
  },
  methods: {
    async login() {
      this.loading = true;
      this.errormsg = null;
      try {
        let response = await this.$axios.post("/login", {
          username: this.username,
          password: this.password,
        });
        console.log(response.data);
      } catch (e) {
        this.errormsg = "Login failed. Please check your credentials.";
      }
      this.loading = false;
    },
  },
  mounted() {
    this. checkLoginStatus()
  }
};
</script>

<template>
  <div class="login-container">
    <h1 class="h2">Login</h1>

    <ErrorMsg v-if="errormsg" :msg="errormsg"></ErrorMsg>

    <form @submit.prevent="login" class="login-form">
      <div class="mb-3">
        <label for="username" class="form-label">Username</label>
        <input
            type="text"
            id="username"
            v-model="username"
            class="form-control"
            placeholder="Enter your username"
            required
        />
      </div>
      <div class="mb-3">
        <label for="password" class="form-label">Password</label>
        <input
            type="password"
            id="password"
            v-model="password"
            class="form-control"
            placeholder="Enter your password"
            required
        />
      </div>

      <div class="d-flex justify-content-between align-items-center">
        <button type="submit" class="btn btn-primary" :disabled="loading">
          {{ loading ? "Logging in..." : "Login" }}
        </button>
      </div>
    </form>
  </div>
</template>

<style scoped>
.login-container {
  max-width: 400px;
  margin: 50px auto;
  padding: 20px;
  border: 1px solid #ccc;
  border-radius: 5px;
  background: #fff;
  box-shadow: 0px 4px 6px rgba(0, 0, 0, 0.1);
}

.login-form .form-label {
  font-weight: 600;
}
</style>
