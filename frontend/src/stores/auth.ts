
import { defineStore } from 'pinia'
import { ref } from 'vue'

export interface User {
    id: number | string
    username: string
    email: string
}

const API_URL =
    import.meta.env.VITE_API_URL ||
    'http://localhost:8080/api/v1'

async function request<T>(
    endpoint: string,
    options: RequestInit = {},
    token?: string
): Promise<T> {
    const headers = new Headers(options.headers)

    headers.set('Content-Type', 'application/json')

    if (token) {
        headers.set('Authorization', `Bearer ${token}`)
    }

    console.log('API_URL:', API_URL)
    console.log('ENDPOINT:', endpoint)
    console.log('FULL URL:', `${API_URL}${endpoint}`)

    const response = await fetch(`${API_URL}${endpoint}`, {
        ...options,
        headers
    })

    const data = await response.json().catch(() => ({}))

    if (!response.ok) {
        throw new Error(
            data.error ||
            data.message ||
            `Ошибка сервера: ${response.status}`
        )
    }

    return data as T
}

export const useAuthStore = defineStore('auth', () => {
    const token = ref<string | null>(
        sessionStorage.getItem('access_token')
    )

    const user = ref<User | null>(null)
    const isAuthenticated = ref(Boolean(token.value))
    const loading = ref(false)

    function clearAuth() {
        token.value = null
        user.value = null
        isAuthenticated.value = false
        sessionStorage.removeItem('access_token')
    }

    async function fetchMe() {
        if (!token.value) {
            clearAuth()
            return null
        }

        try {
            const userData = await request<User>(
                '/users/me',
                { method: 'GET' },
                token.value
            )

            user.value = userData
            isAuthenticated.value = true

            return userData
        } catch (error) {
            clearAuth()
            throw error
        }
    }

    async function register(
        username: string,
        email: string,
        password: string
    ) {
        loading.value = true

        try {
            await request<User>('/auth/register', {
                method: 'POST',
                body: JSON.stringify({
                    username,
                    email,
                    password
                })
            })

            // Backend не возвращает токен при регистрации.
            // Поэтому после создания аккаунта выполняем вход.
            return await loginInternal(email, password)
        } finally {
            loading.value = false
        }
    }

    async function loginInternal(email: string, password: string) {
        const data = await request<{
            access_token: string
            token_type: string
            expires_in: number
        }>('/auth/login', {
            method: 'POST',
            body: JSON.stringify({
                email,
                password
            })
        })

        if (!data.access_token) {
            throw new Error('Сервер не вернул токен авторизации')
        }

        token.value = data.access_token
        sessionStorage.setItem('access_token', data.access_token)
        isAuthenticated.value = true

        try {
            return await fetchMe()
        } catch (error) {
            clearAuth()
            throw error
        }
    }

    async function login(email: string, password: string) {
        loading.value = true

        try {
            return await loginInternal(email, password)
        } finally {
            loading.value = false
        }
    }

    function logout() {
        clearAuth()
    }

    return {
        token,
        user,
        isAuthenticated,
        loading,
        register,
        login,
        logout,
        fetchMe
    }
})
