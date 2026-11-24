
<script setup lang="ts">
import { computed, ref } from 'vue'
import { useAuthStore } from '@/stores/auth'

const authStore = useAuthStore()

const isEditing = ref(false)
const username = ref(authStore.user?.username ?? '')
const description = ref('')

const user = computed(() => authStore.user)

function startEditing() {
  username.value = user.value?.username ?? ''
  isEditing.value = true
}

function cancelEditing() {
  isEditing.value = false
  username.value = user.value?.username ?? ''
}
</script>

<template>
  <main class="profile-page">
    <section class="profile-container">
      <div class="profile-header">
        <div class="avatar">
          {{ user?.username?.charAt(0).toUpperCase() || '?' }}
        </div>

        <div class="profile-info">
          <span class="profile-label">MY PROFILE</span>
          <h1>{{ user?.username || 'Пользователь' }}</h1>
          <p>{{ user?.email }}</p>
          <span class="profile-status">Участник NoisePlace</span>
        </div>

        <button class="edit-button" @click="startEditing">
          Редактировать профиль
        </button>
      </div>

      <div class="profile-divider"></div>

      <section class="about-section">
        <h2>О себе</h2>
        <p>{{ description || 'Здесь пока ничего нет.' }}</p>
      </section>

      <section class="library-section">
        <div class="section-heading">
          <div>
            <span class="section-label">YOUR MUSIC SPACE</span>
            <h2>Моя библиотека</h2>
          </div>
        </div>

        <div class="empty-library">
          <div class="music-icon">♫</div>
          <h3>Твоя музыкальная история начинается здесь</h3>
          <p>
            Создавай плейлисты, сохраняй любимые треки
            и делись музыкой с сообществом.
          </p>
        </div>
      </section>

      <div v-if="isEditing" class="modal-overlay">
        <div class="edit-modal">
          <h2>Редактировать профиль</h2>

          <label>Имя пользователя</label>
          <input v-model="username" type="text" />

          <label>О себе</label>
          <textarea v-model="description"></textarea>

          <div class="modal-actions">
            <button class="cancel-button" @click="cancelEditing">
              Отмена
            </button>
            <button class="save-button" disabled>
              Сохранение пока недоступно
            </button>
          </div>
        </div>
      </div>
    </section>
  </main>
</template>

<style scoped>
.profile-page {
  min-height: 100vh;
  background: #101014;
  color: #f5f3f7;
  padding: 55px 7%;
}

.profile-container {
  max-width: 1100px;
  margin: 0 auto;
}

.profile-header {
  display: flex;
  align-items: center;
  gap: 28px;
  flex-wrap: wrap;
}

.avatar {
  width: 115px;
  height: 115px;
  border-radius: 50%;
  background: linear-gradient(135deg, #a855f7, #5926a5);
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 44px;
  font-weight: 700;
}

.profile-info {
  flex: 1;
}

.profile-label,
.section-label {
  color: #a78bfa;
  font-size: 11px;
  letter-spacing: 2px;
  font-weight: 700;
}

.profile-info h1 {
  font-size: 32px;
  margin: 8px 0;
}

.profile-info p {
  color: #9994a3;
  margin: 0 0 12px;
}

.profile-status {
  color: #bba3ee;
  font-size: 13px;
}

button {
  cursor: pointer;
  border: none;
  border-radius: 9px;
  padding: 12px 19px;
  font-weight: 600;
}

.edit-button {
  background: #25212f;
  color: white;
  border: 1px solid #514064;
}

.edit-button:hover {
  background: #342944;
}

.profile-divider {
  height: 1px;
  background: #292630;
  margin: 38px 0;
}

.about-section h2,
.library-section h2 {
  font-size: 21px;
  margin-bottom: 15px;
}

.about-section p {
  color: #9994a3;
}

.library-section {
  margin-top: 55px;
}

.empty-library {
  margin-top: 22px;
  border: 1px dashed #42364f;
  border-radius: 15px;
  background: #17151d;
  text-align: center;
  padding: 55px 20px;
}

.music-icon {
  color: #b78af6;
  font-size: 42px;
}

.empty-library h3 {
  font-size: 18px;
}

.empty-library p {
  color: #9994a3;
  max-width: 430px;
  margin: 12px auto 0;
  line-height: 1.6;
}

.modal-overlay {
  position: fixed;
  inset: 0;
  background: #000b;
  display: grid;
  place-items: center;
  padding: 20px;
  z-index: 10;
}

.edit-modal {
  background: #1c1922;
  border: 1px solid #42364f;
  border-radius: 16px;
  padding: 28px;
  width: 100%;
  max-width: 460px;
}

.edit-modal h2 {
  margin-top: 0;
}

.edit-modal label {
  display: block;
  margin: 18px 0 8px;
  color: #c5bdd0;
}

.edit-modal input,
.edit-modal textarea {
  width: 100%;
  box-sizing: border-box;
  padding: 12px;
  border: 1px solid #443b50;
  border-radius: 8px;
  background: #111016;
  color: white;
}

.edit-modal textarea {
  min-height: 90px;
  resize: vertical;
}

.modal-actions {
  display: flex;
  flex-wrap: wrap;
  gap: 10px;
  margin-top: 22px;
}

.cancel-button {
  background: #302a38;
  color: white;
}

.save-button {
  background: #7650b4;
  color: white;
  opacity: 0.55;
  cursor: not-allowed;
}

@media (max-width: 600px) {
  .profile-page {
    padding: 35px 20px;
  }

  .profile-header {
    align-items: flex-start;
  }

  .avatar {
    width: 80px;
    height: 80px;
    font-size: 30px;
  }

  .profile-info h1 {
    font-size: 24px;
  }
}
</style>
