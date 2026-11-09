<script setup lang="ts">
import { ref, computed } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'

const router = useRouter()
const auth = useAuthStore()

const activeTab = ref('Все')
const search = ref('')
const savedAlbums = ref<number[]>([2])

const tabs = ['Все', 'Альбомы', 'Синглы', 'EP']

const releases = [
  {
    id: 1,
    title: 'Город не спит',
    artist: 'Ночной режим',
    year: '2025',
    type: 'Альбом',
    color: 'cover-purple',
    symbol: 'N',
  },
  {
    id: 2,
    title: 'Последний этаж',
    artist: 'Северный',
    year: '2025',
    type: 'Альбом',
    color: 'cover-orange',
    symbol: 'С',
  },
  {
    id: 3,
    title: 'Без лишних слов',
    artist: 'KULT',
    year: '2025',
    type: 'EP',
    color: 'cover-blue',
    symbol: 'K',
  },
  {
    id: 4,
    title: 'После полуночи',
    artist: 'Молодой город',
    year: '2024',
    type: 'Сингл',
    color: 'cover-green',
    symbol: 'М',
  },
]

const artists = [
  { name: 'Северный', genre: 'Атмосферный рэп', initials: 'С', color: 'artist-one' },
  { name: 'KULT', genre: 'Альтернативный рэп', initials: 'K', color: 'artist-two' },
  { name: 'Ночной режим', genre: 'Underground', initials: 'Н', color: 'artist-three' },
  { name: 'Молодой город', genre: 'Хип-хоп', initials: 'М', color: 'artist-four' },
]

const collections = [
  {
    title: 'Новый русский звук',
    description: 'Свежие релизы и новые имена',
    count: '24 релиза',
    color: 'collection-one',
  },
  {
    title: 'После полуночи',
    description: 'Музыка для ночных прогулок',
    count: '18 релизов',
    color: 'collection-two',
  },
  {
    title: 'За пределами мейнстрима',
    description: 'Неочевидные музыкальные находки',
    count: '31 релиз',
    color: 'collection-three',
  },
]

const filteredReleases = computed(() => {
  return releases.filter((release) => {
    const matchesTab =
        activeTab.value === 'Все' ||
        release.type === activeTab.value.slice(0, -1) ||
        (activeTab.value === 'Альбомы' && release.type === 'Альбом')

    const query = search.value.toLowerCase()
    const matchesSearch =
        release.title.toLowerCase().includes(query) ||
        release.artist.toLowerCase().includes(query)

    return matchesTab && matchesSearch
  })
})

function toggleSaved(id: number) {
  if (savedAlbums.value.includes(id)) {
    savedAlbums.value = savedAlbums.value.filter((item) => item !== id)
  } else {
    savedAlbums.value.push(id)
  }
}

function goTo(path: string) {
  router.push(path)
}
</script>

<template>
  <div class="noiseplace">
    <aside class="sidebar">
      <div class="brand">
        <div class="brand-icon">N</div>
        <span>noise<span class="brand-light">place</span></span>
      </div>

      <div class="nav-group">
        <div class="nav-label">МЕНЮ</div>

        <button class="nav-item active">
          <span class="nav-icon">⌂</span>
          Обзор
        </button>

        <button class="nav-item" @click="goTo('/search')">
          <span class="nav-icon">⌕</span>
          Поиск
        </button>

        <button class="nav-item" @click="goTo('/artists')">
          <span class="nav-icon">◉</span>
          Исполнители
        </button>

        <button class="nav-item" @click="goTo('/releases')">
          <span class="nav-icon">▧</span>
          Релизы
        </button>
      </div>

      <div class="nav-group">
        <div class="nav-label">ТВОЁ ПРОСТРАНСТВО</div>

        <button class="nav-item" @click="goTo('/collection')">
          <span class="nav-icon">▤</span>
          Моя коллекция
        </button>

        <button class="nav-item" @click="goTo('/playlists')">
          <span class="nav-icon">☷</span>
          Подборки
        </button>

        <button class="nav-item" @click="goTo('/favorites')">
          <span class="nav-icon">♡</span>
          Избранное
        </button>
      </div>

      <div class="sidebar-bottom">
        <div class="sidebar-note">
          <div class="note-icon">✳</div>
          <p>Музыка — это больше, чем просто звук.</p>
          <span>Исследуй. Сохраняй. Делись.</span>
        </div>


        <button
            class="profile-button"
            @click="auth.isAuthenticated ? goTo('/profile') : goTo('/login')"
        >
          <div class="avatar">
            {{ auth.user?.username?.charAt(0).toUpperCase() || 'U' }}
          </div>

          <div class="profile-info">
            <strong>
              {{ auth.user?.username || 'Мой профиль' }}
            </strong>
            <span>
      {{ auth.isAuthenticated ? 'Слушатель' : 'Войти в аккаунт' }}
    </span>
          </div>

          <span class="profile-arrow">↗</span>
        </button>
      </div>
    </aside>

    <main class="main-content">
      <div class="page-content">
        <section class="welcome-section">
          <div class="welcome-copy">
            <div class="eyebrow">
              <span class="eyebrow-dot"></span>
              ТВОЁ МУЗЫКАЛЬНОЕ ПРОСТРАНСТВО
            </div>

            <h1>
              Найди свой<br />
              <span>новый звук.</span>
            </h1>

            <p>
              Исследуй музыку, открывай новых исполнителей
              и собирай коллекцию, которая отражает тебя.
            </p>

            <button class="primary-button" @click="goTo('/releases')">
              Исследовать музыку
              <span>↗</span>
            </button>
          </div>

          <div class="welcome-art">
            <div class="art-glow"></div>
            <div class="art-circle circle-back"></div>
            <div class="art-circle circle-front"></div>
            <div class="art-letter">N</div>
            <div class="art-caption">
              <span>DISCOVER YOUR SOUND</span>
              <strong>NOISEPLACE®</strong>
            </div>
            <div class="art-index">001 — 025</div>
          </div>
        </section>

        <section class="section">
          <div class="section-heading">
            <div>
              <div class="section-kicker">СЛУШАЙ И ОТКРЫВАЙ</div>
              <h2>Новые релизы</h2>
            </div>
            <button class="text-link" @click="goTo('/releases')">
              Все релизы <span>↗</span>
            </button>
          </div>

          <div class="tabs">
            <button
                v-for="tab in tabs"
                :key="tab"
                :class="{ selected: activeTab === tab }"
                @click="activeTab = tab"
            >
              {{ tab }}
            </button>
          </div>

          <div class="release-grid">
            <article
                v-for="release in filteredReleases"
                :key="release.id"
                class="release-card"
            >
              <div class="release-cover" :class="release.color">
                <span class="cover-symbol">{{ release.symbol }}</span>
                <span class="cover-brand">NOISEPLACE</span>
                <button
                    class="save-button"
                    :class="{ saved: savedAlbums.includes(release.id) }"
                    :aria-label="savedAlbums.includes(release.id) ? 'Убрать из коллекции' : 'Сохранить в коллекцию'"
                    @click="toggleSaved(release.id)"
                >
                  {{ savedAlbums.includes(release.id) ? '♥' : '♡' }}
                </button>
              </div>

              <div class="release-info">
                <div class="release-title">{{ release.title }}</div>
                <div class="release-meta">
                  {{ release.artist }} <span>·</span> {{ release.year }}
                </div>
              </div>
            </article>
          </div>

          <div v-if="filteredReleases.length === 0" class="empty-state">
            Ничего не найдено. Попробуй изменить запрос.
          </div>
        </section>

        <section class="section">
          <div class="section-heading">
            <div>
              <div class="section-kicker">ЗНАКОМЬСЯ С НОВЫМИ ИМЕНАМИ</div>
              <h2>Исполнители</h2>
            </div>
            <button class="text-link" @click="goTo('/artists')">
              Все исполнители <span>↗</span>
            </button>
          </div>

          <div class="artist-grid">
            <article
                v-for="artist in artists"
                :key="artist.name"
                class="artist-card"
                @click="goTo('/artists')"
            >
              <div class="artist-avatar" :class="artist.color">
                {{ artist.initials }}
              </div>
              <strong>{{ artist.name }}</strong>
              <span>{{ artist.genre }}</span>
            </article>
          </div>
        </section>

        <section class="section collections-section">
          <div class="section-heading">
            <div>
              <div class="section-kicker">МУЗЫКА В КОНТЕКСТЕ</div>
              <h2>Подборки сообщества</h2>
            </div>
            <button class="text-link" @click="goTo('/playlists')">
              Все подборки <span>↗</span>
            </button>
          </div>

          <div class="collection-grid">
            <article
                v-for="(collection, index) in collections"
                :key="collection.title"
                class="collection-card"
                :class="collection.color"
            >
              <div class="collection-top">
                <span>0{{ index + 1 }} / COMMUNITY</span>
                <span>↗</span>
              </div>

              <div class="collection-bottom">
                <div>
                  <h3>{{ collection.title }}</h3>
                  <p>{{ collection.description }}</p>
                </div>
                <span class="collection-count">{{ collection.count }}</span>
              </div>
            </article>
          </div>
        </section>

        <footer class="footer">
          <div class="footer-brand">NOISEPLACE®</div>
          <span>Твоё пространство для музыкальных открытий.</span>
          <span>© 2025 NoisePlace</span>
        </footer>
      </div>
    </main>
  </div>
</template>

<style>
@import url('https://fonts.googleapis.com/css2?family=DM+Sans:wght@400;500;600;700;800&family=Manrope:wght@400;500;600;700;800&display=swap');

:root {
  font-family: 'DM Sans', sans-serif;
  color: #f5f3fa;
  background: #0c0b10;
  font-synthesis: none;
  text-rendering: optimizeLegibility;
}

* {
  box-sizing: border-box;
}

body {
  margin: 0;
  min-width: 360px;
  background: #0c0b10;
}

button,
input {
  font: inherit;
}

button {
  cursor: pointer;
}

.noiseplace {
  min-height: 100vh;
  background: #0c0b10;
}

.sidebar {
  width: 248px;
  position: fixed;
  inset: 0 auto 0 0;
  z-index: 5;
  padding: 30px 20px 18px;
  display: flex;
  flex-direction: column;
  background: #100f15;
  border-right: 1px solid #24212b;
}

.brand {
  display: flex;
  align-items: center;
  gap: 11px;
  padding: 0 9px;
  font: 800 21px 'Manrope', sans-serif;
  letter-spacing: -1px;
  margin-bottom: 56px;
}

.brand-icon {
  width: 32px;
  height: 32px;
  display: grid;
  place-items: center;
  border-radius: 10px;
  color: white;
  background: #8663ee;
  font-size: 19px;
}

.brand-light {
  color: #9b87d8;
}

.nav-group {
  margin-bottom: 36px;
}

.nav-label,
.section-kicker {
  font-size: 10px;
  font-weight: 800;
  letter-spacing: 1.6px;
  color: #777181;
}

.nav-label {
  padding: 0 10px;
  margin-bottom: 14px;
}

.nav-item {
  width: 100%;
  display: flex;
  align-items: center;
  gap: 13px;
  padding: 12px 11px;
  border: 0;
  border-radius: 8px;
  color: #96919f;
  background: transparent;
  text-align: left;
  font-size: 13px;
  margin-bottom: 4px;
  transition: .2s;
}

.nav-item:hover,
.nav-item.active {
  color: white;
  background: #211c30;
}

.nav-item.active {
  box-shadow: inset 2px 0 #a184ff;
}

.nav-icon {
  width: 19px;
  text-align: center;
  font-size: 19px;
  color: #a38af2;
}

.sidebar-bottom {
  margin-top: auto;
}

.sidebar-note {
  padding: 17px;
  border: 1px solid #2a2537;
  border-radius: 11px;
  background: linear-gradient(145deg, #211a32, #15121d);
  margin-bottom: 20px;
}

.note-icon {
  color: #a88aff;
  font-size: 21px;
}

.sidebar-note p {
  font-size: 12px;
  line-height: 1.6;
  margin: 10px 0 6px;
  color: #e9e3f7;
}

.sidebar-note span {
  font-size: 10px;
  color: #91869f;
}

.profile-button {
  display: flex;
  align-items: center;
  width: 100%;
  gap: 11px;
  border: 0;
  border-top: 1px solid #292630;
  padding: 17px 3px 3px;
  color: white;
  background: transparent;
  text-align: left;
}

.avatar {
  width: 34px;
  height: 34px;
  display: grid;
  place-items: center;
  border-radius: 50%;
  background: #66528e;
  font-weight: 700;
}

.profile-info {
  display: flex;
  flex: 1;
  flex-direction: column;
  gap: 3px;
}

.profile-info strong {
  font-size: 12px;
}

.profile-info span {
  color: #85808d;
  font-size: 10px;
}

.profile-arrow {
  color: #8e879b;
}

.main-content {
  margin-left: 248px;
  min-height: 100vh;
}

.topbar {
  height: 75px;
  padding: 0 5%;
  display: flex;
  align-items: center;
  justify-content: space-between;
  border-bottom: 1px solid #24212b;
  background: #0e0d13;
}

.breadcrumbs {
  display: flex;
  align-items: center;
  gap: 12px;
  color: #817b89;
  font-size: 12px;
}

.breadcrumbs strong {
  color: #e9e6ee;
  font-weight: 500;
}

.crumb-divider {
  color: #494451;
}

.topbar-actions {
  display: flex;
  gap: 13px;
  align-items: center;
}

.search-box {
  width: 280px;
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
}

.auth-actions {
  display: flex;
  align-items: center;
  gap: 9px;
}

.auth-login,
.auth-register {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  min-height: 38px;
  padding: 0 15px;
  border-radius: 7px;
  text-decoration: none;
  white-space: nowrap;
  font-size: 11px;
  font-weight: 700;
  transition: .2s;
}

.auth-login {
  color: #eee9f8;
  border: 1px solid #393244;
  background: #15131b;
}

.auth-login:hover {
  border-color: #8065b8;
  background: #211b2d;
}

.auth-register {
  color: #fff;
  border: 1px solid #8464d7;
  background: #8464d7;
}

.auth-register:hover {
  background: #997bf0;
  border-color: #997bf0;
}

.page-content {
  width: 100%;
  max-width: 1440px;
  padding: 37px 5% 20px;
  margin: 0 auto;
}

.welcome-section {
  min-height: 320px;
  position: relative;
  display: flex;
  align-items: center;
  overflow: hidden;
  padding: 35px 48px;
  border: 1px solid #342944;
  border-radius: 13px;
  background:
      radial-gradient(ellipse at 80% 50%, #3b275d 0%, transparent 42%),
      linear-gradient(110deg, #1b1526, #17121f 70%, #251a35);
}

.welcome-copy {
  position: relative;
  z-index: 2;
  max-width: 470px;
}

.eyebrow {
  display: flex;
  align-items: center;
  gap: 9px;
  color: #b6a4dc;
  font-size: 9px;
  font-weight: 700;
  letter-spacing: 1.7px;
}

.eyebrow-dot {
  width: 7px;
  height: 7px;
  border-radius: 50%;
  background: #a68aff;
  box-shadow: 0 0 12px #a68aff;
}

.welcome-section h1 {
  font: 800 clamp(38px, 5vw, 60px)/1.08 'Manrope', sans-serif;
  letter-spacing: -3px;
  margin: 25px 0 15px;
}

.welcome-section h1 span {
  color: #a78aff;
}

.welcome-copy > p {
  max-width: 370px;
  color: #aaa2b5;
  font-size: 13px;
  line-height: 1.8;
  margin-bottom: 23px;
}

.primary-button {
  padding: 13px 18px;
  display: inline-flex;
  align-items: center;
  gap: 23px;
  color: white;
  background: #8464d7;
  border: 0;
  border-radius: 6px;
  font-size: 11px;
  font-weight: 700;
  transition: .2s;
}

.primary-button:hover {
  background: #997bf0;
  transform: translateY(-2px);
}

.welcome-art {
  position: absolute;
  width: 390px;
  height: 320px;
  right: 35px;
  top: 0;
}

.art-glow {
  position: absolute;
  width: 260px;
  height: 260px;
  right: 45px;
  top: 25px;
  background: #8052c9;
  opacity: .2;
  filter: blur(80px);
}

.art-circle {
  position: absolute;
  border-radius: 50%;
  transform: rotate(-20deg);
}

.circle-back {
  width: 225px;
  height: 225px;
  top: 47px;
  right: 65px;
  border: 1px solid #b49aff;
  opacity: .6;
}

.circle-front {
  width: 185px;
  height: 185px;
  top: 68px;
  right: 85px;
  background: linear-gradient(145deg, #b09aff, #60418e 55%, #231832);
  box-shadow: 0 15px 80px #9061db55;
}

.art-letter {
  position: absolute;
  top: 70px;
  right: 130px;
  z-index: 1;
  font: 800 145px 'Manrope', sans-serif;
  color: #fff;
  text-shadow: 0 8px 40px #302044;
}

.art-caption {
  position: absolute;
  bottom: 25px;
  right: 15px;
  display: flex;
  flex-direction: column;
  gap: 5px;
  text-align: right;
  letter-spacing: 2px;
}

.art-caption span {
  color: #a89abf;
  font-size: 8px;
}

.art-caption strong {
  font-size: 12px;
}

.art-index {
  position: absolute;
  left: 0;
  bottom: 25px;
  color: #9a88b5;
  font-size: 9px;
  letter-spacing: 2px;
}

.section {
  margin-top: 49px;
}

.section-heading {
  display: flex;
  align-items: end;
  justify-content: space-between;
  margin-bottom: 22px;
}

.section-heading h2 {
  margin: 7px 0 0;
  font: 700 24px 'Manrope', sans-serif;
  letter-spacing: -.7px;
}

.text-link {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 7px 0;
  color: #a992e6;
  background: transparent;
  border: 0;
  font-size: 11px;
}

.text-link:hover {
  color: white;
}

.tabs {
  display: flex;
  gap: 7px;
  margin-bottom: 18px;
}

.tabs button {
  padding: 7px 14px;
  color: #938c9d;
  border: 1px solid #2a2731;
  border-radius: 5px;
  background: #14121a;
  font-size: 10px;
}

.tabs button.selected {
  color: white;
  background: #34274e;
  border-color: #624b8e;
}

.release-grid {
  display: grid;
  grid-template-columns: repeat(4, minmax(0, 1fr));
  gap: 20px;
}

.release-card {
  min-width: 0;
}

.release-cover {
  aspect-ratio: 1;
  position: relative;
  display: grid;
  place-items: center;
  overflow: hidden;
  border-radius: 8px;
  transition: transform .2s;
}

.release-card:hover .release-cover {
  transform: translateY(-4px);
}

.cover-purple {
  background: radial-gradient(circle at 60% 35%, #bb8aff, #653b96 40%, #25152f);
}

.cover-orange {
  background: radial-gradient(circle at 30% 70%, #ffca80, #a94e32 45%, #351c25);
}

.cover-blue {
  background: radial-gradient(circle at 65% 40%, #81c9df, #285c83 50%, #101c34);
}

.cover-green {
  background: radial-gradient(circle at 50% 45%, #c4d58c, #496f55 50%, #172a28);
}

.cover-symbol {
  font: 800 clamp(65px, 8vw, 110px) 'Manrope', sans-serif;
  color: #ffffffc9;
  text-shadow: 0 5px 30px #0005;
}

.cover-brand {
  position: absolute;
  left: 12px;
  bottom: 11px;
  font-size: 8px;
  letter-spacing: 2px;
  color: #ffffffa8;
}

.save-button {
  position: absolute;
  top: 10px;
  right: 10px;
  width: 30px;
  height: 30px;
  border: 1px solid #ffffff45;
  border-radius: 50%;
  color: white;
  background: #0005;
  font-size: 16px;
}

.save-button.saved {
  color: #e4b7ff;
  background: #4b326c;
}

.release-info {
  padding: 12px 2px;
}

.release-title {
  overflow: hidden;
  color: #f0edf5;
  font-size: 12px;
  font-weight: 700;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.release-meta {
  margin-top: 5px;
  color: #898391;
  font-size: 10px;
}

.release-meta span {
  padding: 0 4px;
  color: #665d73;
}

.empty-state {
  padding: 30px;
  color: #91899c;
  text-align: center;
  border: 1px dashed #383143;
  border-radius: 8px;
  font-size: 12px;
}

.artist-grid {
  display: grid;
  grid-template-columns: repeat(4, minmax(0, 1fr));
  gap: 13px;
}

.artist-card {
  display: flex;
  flex-direction: column;
  align-items: center;
  padding: 21px 10px;
  border: 1px solid #28242f;
  border-radius: 9px;
  background: #131119;
  cursor: pointer;
  transition: .2s;
}

.artist-card:hover {
  border-color: #7657ae;
  background: #1b1725;
  transform: translateY(-3px);
}

.artist-avatar {
  width: 78px;
  height: 78px;
  display: grid;
  place-items: center;
  margin-bottom: 14px;
  border-radius: 50%;
  font: 700 30px 'Manrope', sans-serif;
  color: white;
}

.artist-one {
  background: linear-gradient(145deg, #b17b72, #432c45);
}

.artist-two {
  background: linear-gradient(145deg, #5e9db0, #243449);
}

.artist-three {
  background: linear-gradient(145deg, #a181ce, #3a294d);
}

.artist-four {
  background: linear-gradient(145deg, #9a9b66, #39472f);
}

.artist-card strong {
  font-size: 12px;
}

.artist-card > span {
  margin-top: 5px;
  color: #8c8496;
  font-size: 10px;
}

.collection-grid {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 15px;
}

.collection-card {
  min-height: 190px;
  display: flex;
  flex-direction: column;
  justify-content: space-between;
  padding: 17px;
  border: 1px solid #393044;
  border-radius: 9px;
  transition: .2s;
}

.collection-card:hover {
  transform: translateY(-3px);
  border-color: #9a7cce;
}

.collection-one {
  background: linear-gradient(145deg, #38274c, #21182d);
}

.collection-two {
  background: linear-gradient(145deg, #49302e, #24191f);
}

.collection-three {
  background: linear-gradient(145deg, #253d42, #17262e);
}

.collection-top {
  display: flex;
  justify-content: space-between;
  color: #c0afda;
  font-size: 9px;
  letter-spacing: 1px;
}

.collection-bottom h3 {
  margin: 0 0 7px;
  font: 700 16px 'Manrope', sans-serif;
}

.collection-bottom p {
  margin: 0;
  color: #b4a9bd;
  font-size: 10px;
}

.collection-count {
  display: inline-block;
  margin-top: 14px;
  color: #d2c0e8;
  font-size: 9px;
}

.footer {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 15px;
  margin-top: 60px;
  padding: 20px 0;
  border-top: 1px solid #29252f;
  color: #777180;
  font-size: 10px;
}

.footer-brand {
  color: #b49be9;
  font-weight: 800;
  letter-spacing: 2px;
}

@media (max-width: 1050px) {
  .sidebar {
    width: 205px;
  }

  .main-content {
    margin-left: 205px;
  }

  .release-grid {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }

  .welcome-art {
    right: -100px;
    opacity: .65;
  }
}

@media (max-width: 700px) {
  .sidebar {
    position: static;
    width: 100%;
    height: auto;
    padding: 15px 18px;
    border-right: 0;
    border-bottom: 1px solid #29252f;
  }

  .brand {
    margin: 0;
  }

  .nav-group,
  .sidebar-bottom {
    display: none;
  }

  .main-content {
    margin-left: 0;
  }

  .topbar {
    height: auto;
    min-height: 65px;
    padding: 12px 18px;
  }

  .breadcrumbs {
    display: none;
  }

  .topbar-actions {
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

  .auth-login,
  .auth-register {
    min-height: 34px;
    padding: 0 11px;
    font-size: 10px;
  }

  .page-content {
    padding: 20px 16px;
  }

  .welcome-section {
    min-height: 390px;
    align-items: start;
    padding: 30px 23px;
  }

  .welcome-section h1 {
    font-size: 43px;
  }

  .welcome-art {
    top: 180px;
    right: -95px;
    opacity: .35;
  }

  .release-grid {
    grid-template-columns: repeat(2, minmax(0, 1fr));
    gap: 13px;
  }

  .artist-grid {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }

  .collection-grid {
    grid-template-columns: 1fr;
  }

  .section-heading h2 {
    font-size: 20px;
  }

  .text-link {
    font-size: 10px;
  }

  .footer {
    flex-wrap: wrap;
  }
}
</style>
