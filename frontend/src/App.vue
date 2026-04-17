<template>
  <UApp>
    <div v-if="!sessionLoaded" class="flex items-center justify-center min-h-screen bg-default">
      <UIcon name="i-lucide-loader-circle" class="w-8 h-8 animate-spin text-primary" />
    </div>
    <div v-else class="grid grid-rows-[auto_1fr_auto] min-h-screen w-full bg-default">
      <header class="flex items-center justify-between px-6 py-3 border-b border-default bg-default">
        <div class="flex items-center gap-3">
          <h1 class="text-xl font-bold">CUPS 打印</h1>
          <span v-if="session" class="text-sm text-muted">{{ session.username }}</span>
        </div>
        <div class="flex items-center gap-2">
          <UButton v-if="isAdmin" :variant="route.path === '/print' ? 'solid' : 'ghost'" size="sm" @click="router.push('/print')">打印</UButton>
          <UButton v-if="isAdmin" :variant="route.path === '/admin' ? 'solid' : 'ghost'" size="sm" @click="router.push('/admin')">管理</UButton>
          <UButton v-if="session" variant="outline" size="sm" @click="logout">登出</UButton>
        </div>
      </header>
      <div class="overflow-auto relative">
        <router-view :session="session" @login-success="onLogin" @logout="onLogout" />
      </div>
      <footer class="px-6 py-3 border-t border-default bg-default text-sm text-muted text-center">
        Powered by <a href="https://github.com/hanxi/cups-web" target="_blank" class="text-primary hover:underline">cups-web</a>
      </footer>
    </div>
  </UApp>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { clearSessionCache, updateSessionCache } from './router'

const router = useRouter()
const route = useRoute()

const session = ref(null)
const sessionLoaded = ref(false)

const isAdmin = computed(() => session.value?.role === 'admin')

async function loadSession() {
  try {
    const resp = await fetch('/api/session', { credentials: 'include' })
    if (resp.ok) {
      const data = await resp.json()
      session.value = data
      updateSessionCache(data)
      router.push('/print')
    } else {
      session.value = null
      router.push('/login')
    }
  } catch (e) {
    session.value = null
  } finally {
    sessionLoaded.value = true
  }
}

function onLogin() {
  loadSession()
}

function onLogout() {
  session.value = null
  clearSessionCache()
  router.push('/login')
}

async function logout() {
  try {
    await fetch('/api/logout', { method: 'POST', credentials: 'include' })
  } catch (e) {
    // ignore errors
  }
  onLogout()
}

onMounted(() => loadSession())
</script>
