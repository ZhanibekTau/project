import { reactive } from 'vue';

export const authState = reactive({
    isLoggedIn: false, // Tracks if the user is logged in
});

export function checkLoginStatus() {
    return !!localStorage.getItem('authToken');
}

export function logIn(token) {
    localStorage.setItem('authToken', token);
    authState.isLoggedIn = true;
}

export function logOut() {
    localStorage.removeItem('authToken');
    authState.isLoggedIn = false;
}
