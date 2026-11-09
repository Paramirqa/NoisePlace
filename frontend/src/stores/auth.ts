import { defineStore } from 'pinia'
import { ref } from 'vue'

export interface User {
    id: number | string
    username: string
    email: string
}

const API_URL = import.meta.env.VITE_API_URL ?? ''

export const useAuthStore = defineStore('auth', () => {
    const token = ref<string | null>(sessionStorage.getItem('access_token'))
    const user = ref<User | null>(null)
    const isAuthenticated = ref(Boolean(token.value))
    const loading = ref(false)

    async function register(
        username: string,
        email: string,
        password: string
    ) {
        loading.value = true

        try {
            const response = await fetch(`${API_URL}/api/v1/auth/register`, {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json'
                },
                body: JSON.stringify({
                    username,
                    email,
                    password
                })
            })

            const data = await response.json().catch(() => ({}))

            if (!response.ok) {
                throw new Error(data.error || 'Ошибка регистрации')
            }

            try {
                await login(email, password)
            } catch {
                throw new Error(
                    'Аккаунт создан, но автоматический вход не удался. Попробуйте войти вручную.'
                )
            }

            return data
        } finally {
            loading.value = false
        }
    }

    async function login(email: string, password: string) {
        loading.value = true

        try {
            const response = await fetch(`${API_URL}/api/v1/auth/login`, {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json'
                },
                body: JSON.stringify({
                    email,
                    password
                })
            })

            const data = await response.json().catch(() => ({}))

            if (!response.ok) {
                throw new Error(data.error || 'Неверный email или пароль')
            }

            if (!data.access_token) {
                throw new Error('Сервер не вернул токен авторизации')
            }

            token.value = data.access_token
            sessionStorage.setItem('access_token', data.access_token)

            const userResponse = await fetch(`${API_URL}/api/v1/users/me`, {
                headers: {
                    Authorization: `Bearer ${token.value}`
                }
            })

            const userData = await userResponse.json().catch(() => ({}))

            if (!userResponse.ok) {
                token.value = null
                sessionStorage.removeItem('access_token')
                throw new Error(userData.error || 'Не удалось получить данные пользователя')
            }

            user.value = userData
            isAuthenticated.value = true

            return userData
        } finally {
            loading.value = false
        }
    }

    function logout() {
        token.value = null
        user.value = null
        isAuthenticated.value = false
        sessionStorage.removeItem('access_token')
    }

    return {
        token,
        user,
        isAuthenticated,
        loading,
        register,
        login,
        logout
    }
})
