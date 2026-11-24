<script setup lang="ts">
import { ref } from 'vue'
import { useAuthStore } from '@/stores/auth'
import { useRouter } from 'vue-router'

const auth = useAuthStore()
const router = useRouter()

const search = ref('')

function handleLogout() {
  auth.logout()
  router.push('/')
}

function handleSearch() {
  const query = search.value.trim()

  if (query) {
    router.push({
      path: '/search',
      query: { q: query },
    })
  }
}
</script>

<template>
  <header class="header">
    <div class="header-container">
      <RouterLink to="/" class="logo">
        <span class="logo-icon">N</span>
        <span>noise<span class="logo-accent">place</span></span>
      </RouterLink>

      <nav class="navigation">
        <RouterLink to="/" class="nav-link">
          Главная
        </RouterLink>
        <RouterLink to="/artists" class="nav-link">
          Исполнители
        </RouterLink>
        <RouterLink to="/releases" class="nav-link">
          Релизы
        </RouterLink>
        <RouterLink to="/playlists" class="nav-link">
          Подборки
        </RouterLink>
      </nav>

      <div class="header-actions">
        <label class="search-box">
          <span>⌕</span>
          <input
              v-model="search"
              placeholder="Найти релиз или исполнителя..."
              @keydown.enter="handleSearch"
          />
          <kbd>⌘ K</kbd>
        </label>

        <button class="notification-button" aria-label="Уведомления">
          ♧
        </button>

        <div class="auth-actions">
          <template v-if="!auth.isAuthenticated">
            <RouterLink to="/login" class="login-button">
              Войти
            </RouterLink>

            <RouterLink to="/register" class="register-button">
              Регистрация
            </RouterLink>
          </template>

          <template v-else>
            <RouterLink to="/profile" class="username">
              {{ auth.user?.username || 'Мой профиль' }}
            </RouterLink>

            <button class="logout-button" @click="handleLogout">
              Выйти
            </button>
          </template>
        </div>
      </div>
    </div>
  </header>
</template>

<style scoped>
.header {
  width: 100%;
  background: #0e0d13;
  border-bottom: 1px solid #24212b;
  position: sticky;
  top: 0;
  z-index: 100;
}

.header-container {
  min-height: 75px;
  padding: 0 5%;
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 20px;
}

.logo {
  display: flex;
  align-items: center;
  gap: 10px;
  color: white;
  font: 800 21px 'Manrope', sans-serif;
  text-decoration: none;
  white-space: nowrap;
}

.logo-icon {
  width: 32px;
  height: 32px;
  display: grid;
  place-items: center;
  border-radius: 10px;
  background: #8663ee;
  font-size: 19px;
}

.logo-accent {
  color: #9b87d8;
}

.navigation {
  display: flex;
  align-items: center;
  gap: 18px;
}

.nav-link {
  color: #96919f;
  text-decoration: none;
  font-size: 12px;
  white-space: nowrap;
}

.nav-link:hover,
.nav-link.router-link-active {
  color: white;
}

.header-actions {
  display: flex;
  align-items: center;
  gap: 10px;
}

.search-box {
  width: 240px;
  height: 38px;
  display: flex;
  align-items: center;
  gap: 9px;
  padding: 0 11px;
  border: 1px solid #2b2733;
  border-radius: 7px;
  background: #15131b;
  color: #918a9e;
}

.search-box input {
  width: 100%;
  min-width: 0;
  border: 0;
  outline: none;
  color: white;
  background: transparent;
  font-size: 11px;
}

.search-box kbd {
  white-space: nowrap;
  color: #777180;
  font-size: 10px;
}

.notification-button {
  width: 36px;
  height: 36px;
  border: 1px solid #2b2733;
  border-radius: 7px;
  background: #15131b;
  color: #b6a4e8;
  font-size: 18px;
  cursor: pointer;
}

.auth-actions {
  display: flex;
  align-items: center;
  gap: 9px;
}

.login-button,
.register-button,
.logout-button {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  min-height: 38px;
  padding: 0 13px;
  border-radius: 7px;
  text-decoration: none;
  white-space: nowrap;
  font-size: 11px;
  font-weight: 700;
  cursor: pointer;
}

.login-button {
  color: #eee9f8;
  border: 1px solid #393244;
  background: #15131b;
}

.register-button {
  color: white;
  border: 1px solid #8464d7;
  background: #8464d7;
}

.username {
  color: #eee9f8;
  font-size: 12px;
  font-weight: 700;
  text-decoration: none;
  white-space: nowrap;
}

.logout-button {
  color: #d1c5ff;
  background: transparent;
  border: 1px solid #65509b;
}

@media (max-width: 1100px) {
  .navigation {
    display: none;
  }

  .header-container {
    flex-wrap: wrap;
    padding-top: 12px;
    padding-bottom: 12px;
  }
}

@media (max-width: 700px) {
  .header-container {
    gap: 12px;
  }

  .header-actions {
    width: 100%;
    flex-wrap: wrap;
  }

  .search-box {
    width: 100%;
    flex: 1 1 100%;
  }

  .auth-actions {
    margin-left: auto;
  }
}
</style>
