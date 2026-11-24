<script setup lang="ts">
import { useRoute } from 'vue-router'
import Header from '@/components/Header.vue'

const route = useRoute()

const hideHeaderRoutes = ['/login', '/register']
</script>

<template>
  <Header v-if="!hideHeaderRoutes.includes(route.path)" />
  <RouterView />
</template>
