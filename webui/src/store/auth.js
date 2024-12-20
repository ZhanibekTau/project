import { reactive } from 'vue';

export const authState = reactive({
    isLoggedIn: false, // Tracks if the user is logged in
});

export function checkLoginStatus() {
    return !!localStorage.getItem('authToken');
}

export function logIn(token, id) {
    localStorage.setItem('authToken', token);
    localStorage.setItem('id', id);
    authState.isLoggedIn = true;
}

export function logOut() {
    localStorage.removeItem('authToken');
    localStorage.removeItem('id');
    authState.isLoggedIn = false;
}

export function getToken() {
    return localStorage.getItem('authToken');
}

export function getId() {
    return localStorage.getItem('id');
}
