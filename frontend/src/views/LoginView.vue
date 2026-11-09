<template>
  <main class="auth-page">
    <div class="auth-background"></div>

    <section class="auth-card">
      <router-link to="/" class="logo">
        Noise<span>Place</span>
      </router-link>

      <div class="auth-heading">
        <h1>С возвращением</h1>
        <p>Войди в аккаунт и продолжай открывать музыку</p>
      </div>

      <form @submit.prevent="handleLogin">
        <div class="form-group">
          <label for="email">Электронная почта</label>
          <input
              id="email"
              v-model.trim="email"
              type="email"
              placeholder="example@email.com"
              autocomplete="email"
              required
          />
        </div>

        <div class="form-group">
          <label for="password">Пароль</label>
          <input
              id="password"
              v-model="password"
              type="password"
              placeholder="Введи свой пароль"
              autocomplete="current-password"
              required
          />
        </div>

        <p v-if="errorMessage" class="message error">
          {{ errorMessage }}
        </p>

        <button type="submit" :disabled="auth.loading">
          {{ auth.loading ? 'Входим...' : 'Войти' }}
        </button>
      </form>

      <div class="auth-footer">
        Нет аккаунта?
        <router-link to="/register">Зарегистрироваться</router-link>
      </div>
    </section>
  </main>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '../stores/auth'

const router = useRouter()
const auth = useAuthStore()

const email = ref('')
const password = ref('')
const errorMessage = ref('')

async function handleLogin() {
  errorMessage.value = ''

  try {
    await auth.login(email.value, password.value)

    await router.push('/')
  } catch (error) {
    errorMessage.value =
        error instanceof Error
            ? error.message
            : 'Не удалось войти в аккаунт'
  }
}
</script>

<style scoped>
.auth-page {
  min-height: 100vh;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 40px 20px;
  background: #0b0912;
  color: #fff;
  position: relative;
  overflow: hidden;
  font-family: Inter, Arial, sans-serif;
}

.auth-background {
  position: absolute;
  width: 600px;
  height: 600px;
  top: -300px;
  left: 50%;
  transform: translateX(-50%);
  background: #7839d6;
  opacity: 0.15;
  filter: blur(150px);
  pointer-events: none;
}

.auth-card {
  position: relative;
  width: 100%;
  max-width: 440px;
  padding: 40px;
  background: rgba(24, 20, 35, 0.95);
  border: 1px solid #342947;
  border-radius: 20px;
  box-shadow: 0 20px 80px rgba(0, 0, 0, 0.4);
}

.logo {
  display: block;
  margin-bottom: 35px;
  text-align: center;
  font-size: 30px;
  font-weight: 800;
  color: #fff;
  text-decoration: none;
  letter-spacing: -1px;
}

.logo span {
  color: #a879ff;
}

.auth-heading {
  margin-bottom: 30px;
  text-align: center;
}

.auth-heading h1 {
  margin-bottom: 10px;
  font-size: 26px;
}

.auth-heading p {
  color: #a39bad;
  font-size: 14px;
  line-height: 1.5;
}

.form-group {
  margin-bottom: 20px;
}

label {
  display: block;
  margin-bottom: 9px;
  color: #d9d2e5;
  font-size: 14px;
}

input {
  box-sizing: border-box;
  width: 100%;
  padding: 14px 15px;
  background: #100e18;
  border: 1px solid #393047;
  border-radius: 9px;
  color: #fff;
  outline: none;
  font-size: 14px;
  transition: border-color 0.2s, box-shadow 0.2s;
}

input:focus {
  border-color: #a879ff;
  box-shadow: 0 0 0 3px rgba(168, 121, 255, 0.12);
}

input::placeholder {
  color: #70697d;
}

button {
  width: 100%;
  padding: 15px;
  margin-top: 8px;
  background: linear-gradient(90deg, #8248dc, #a76aff);
  border: none;
  border-radius: 9px;
  color: white;
  font-size: 15px;
  font-weight: 700;
  cursor: pointer;
  transition: opacity 0.2s, transform 0.2s;
}

button:hover:not(:disabled) {
  opacity: 0.9;
  transform: translateY(-1px);
}

button:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

.message {
  margin-bottom: 15px;
  font-size: 13px;
}

.error {
  color: #ff7185;
}

.auth-footer {
  margin-top: 25px;
  text-align: center;
  color: #a39bad;
  font-size: 14px;
}

.auth-footer a {
  color: #b58aff;
  text-decoration: none;
  margin-left: 4px;
}

.auth-footer a:hover {
  text-decoration: underline;
}

@media (max-width: 480px) {
  .auth-card {
    padding: 30px 22px;
  }

  .auth-heading h1 {
    font-size: 23px;
  }
}
</style>