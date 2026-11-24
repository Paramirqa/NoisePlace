
import { createRouter, createWebHistory } from 'vue-router'
import { useAuthStore } from '../stores/auth'

import LoginView from '../views/LoginView.vue'
import RegisterView from '../views/RegisterView.vue'
import HomeView from '../views/HomeView.vue'
import ProfileView from '@/views/ProfileView.vue'

const router = createRouter({
    history: createWebHistory(import.meta.env.BASE_URL),

    routes: [
        {
            path: '/',
            name: 'home',
            component: HomeView
        },
        {
            path: '/login',
            name: 'login',
            component: LoginView
        },
        {
            path: '/register',
            name: 'register',
            component: RegisterView
        },
        {
            path: '/profile',
            name: 'profile',
            component: ProfileView,
            meta: { requiresAuth: true }
        }

        // Пример будущего приватного маршрута:
        // {
        //   path: '/profile',
        //   name: 'profile',
        //   component: ProfileView,
        //   meta: { requiresAuth: true }
        // }
    ]
})

router.beforeEach(async (to) => {
    const auth = useAuthStore()

    if (auth.token && !auth.user) {
        try {
            await auth.fetchMe()
        } catch {
            // Токен недействителен или истёк.
        }
    }

    if (to.meta.requiresAuth && !auth.isAuthenticated) {
        return {
            name: 'login',
            query: { redirect: to.fullPath }
        }
    }

    return true
})

export default router
