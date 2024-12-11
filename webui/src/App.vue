  <script setup>
  import { RouterLink, RouterView } from 'vue-router'
  </script>
  <script>
  import {checkLoginStatus, logIn} from "./store/auth";

  export default {
    data: function() {
      return {
        errormsg: null,
        loading: false,
        username: "",
        password: "",
        checkLogin: false,
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
          logIn(response.data['token'])
          window.location.reload();
        } catch (e) {
          this.errormsg = "Login failed. Please check your credentials.";
        }
        this.loading = false;
      },
    },
    mounted() {
      this.checkLogin = checkLoginStatus()
    }
  }
  </script>

  <template>
    <div class="login-container" v-if="!checkLogin">
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
    <div v-else>
      <header class="navbar navbar-dark sticky-top bg-dark flex-md-nowrap p-0 shadow">
        <a class="navbar-brand col-md-3 col-lg-2 me-0 px-3 fs-6" href="#/">Project</a>
        <button class="navbar-toggler position-absolute d-md-none collapsed" type="button" data-bs-toggle="collapse" data-bs-target="#sidebarMenu" aria-controls="sidebarMenu" aria-expanded="false" aria-label="Toggle navigation">
          <span class="navbar-toggler-icon"></span>
        </button>
      </header>

      <div class="container-fluid">
        <div class="row">
          <nav id="sidebarMenu" class="col-md-3 col-lg-2 d-md-block bg-light sidebar collapse">
            <div class="position-sticky pt-3 sidebar-sticky">
              <h6 class="sidebar-heading d-flex justify-content-between align-items-center px-3 mt-4 mb-1 text-muted text-uppercase">
                <span>General</span>
              </h6>
              <ul class="nav flex-column">
                <li class="nav-item">
                  <RouterLink to="/" class="nav-link">
                    <svg class="feather"><use href="/feather-sprite-v4.29.0.svg#home"/></svg>
                    Home
                  </RouterLink>
                </li>
                <li class="nav-item">
                  <RouterLink to="/link1" class="nav-link">
                    <svg class="feather"><use href="/feather-sprite-v4.29.0.svg#layout"/></svg>
                    Menu item 1
                  </RouterLink>
                </li>
                <li class="nav-item">
                  <RouterLink to="/link2" class="nav-link">
                    <svg class="feather"><use href="/feather-sprite-v4.29.0.svg#key"/></svg>
                    Menu item 2
                  </RouterLink>
                </li>
              </ul>

              <h6 class="sidebar-heading d-flex justify-content-between align-items-center px-3 mt-4 mb-1 text-muted text-uppercase">
                <span>Secondary menu</span>
              </h6>
              <ul class="nav flex-column">
                <li class="nav-item">
                  <RouterLink :to="'/some/' + 'variable_here' + '/path'" class="nav-link">
                    <svg class="feather"><use href="/feather-sprite-v4.29.0.svg#file-text"/></svg>
                    Item 1
                  </RouterLink>
                </li>
              </ul>
            </div>
          </nav>

          <main class="col-md-9 ms-sm-auto col-lg-10 px-md-4">
            <RouterView />
          </main>
        </div>
      </div>
    </div>
  </template>

  <style>
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
