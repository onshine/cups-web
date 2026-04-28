<template>
  <div ref="container" class="w-full h-full flex items-center justify-center">
    <div v-if="loading" class="text-center text-muted">
      <UIcon name="i-lucide-loader-circle" class="w-6 h-6 animate-spin" />
    </div>
    <div v-else-if="error" class="text-center text-muted text-sm p-4">
      <p>PDF 预览加载失败</p>
    </div>
    <div v-show="!loading && !error" class="relative w-full h-full flex items-center justify-center">
      <canvas ref="canvas" class="max-w-full max-h-full" />
      <div v-if="totalPages > 1 && !loading && !error" class="absolute bottom-2 left-1/2 -translate-x-1/2 flex flex-nowrap items-center gap-1 sm:gap-2 px-2 sm:px-3 py-1 bg-black/40 rounded-full w-max max-w-[90%] whitespace-nowrap" style="backdrop-filter: blur(4px)">
        <UButton size="xs" variant="ghost" color="white" icon="i-lucide-chevron-left" :disabled="currentPage <= 1" class="flex-shrink-0" @click="prevPage" />
        <span class="text-xs text-white whitespace-nowrap flex-shrink-0">{{ currentPage }} / {{ totalPages }}</span>
        <UButton size="xs" variant="ghost" color="white" icon="i-lucide-chevron-right" :disabled="currentPage >= totalPages" class="flex-shrink-0" @click="nextPage" />
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted, watch, onUnmounted, nextTick } from 'vue'
import * as pdfjsLib from 'pdfjs-dist'
import pdfjsWorker from 'pdfjs-dist/build/pdf.worker.min.mjs?url'

pdfjsLib.GlobalWorkerOptions.workerSrc = pdfjsWorker

const props = defineProps({
  src: { type: String, required: true }
})

const canvas = ref(null)
const container = ref(null)
const loading = ref(true)
const error = ref(false)
const currentPage = ref(1)
const totalPages = ref(0)
let pdfDoc = null
let renderTask = null
let resizeObserver = null
let lastWidth = 0
let lastHeight = 0

async function renderPage(pageNum) {
  if (!pdfDoc || !canvas.value) return

  try {
    if (renderTask) {
      renderTask.cancel()
      renderTask = null
    }

    const page = await pdfDoc.getPage(pageNum)

    const containerEl = container.value
    if (!containerEl) return

    const containerWidth = containerEl.clientWidth
    const containerHeight = containerEl.clientHeight

    // 容器尺寸无效时不渲染
    if (containerWidth <= 0 || containerHeight <= 0) return

    const viewport = page.getViewport({ scale: 1 })
    const scaleX = containerWidth / viewport.width
    const scaleY = containerHeight / viewport.height
    const scale = Math.min(scaleX, scaleY, 2)

    const scaledViewport = page.getViewport({ scale })

    const ctx = canvas.value.getContext('2d')
    canvas.value.width = scaledViewport.width
    canvas.value.height = scaledViewport.height

    renderTask = page.render({
      canvasContext: ctx,
      viewport: scaledViewport
    })

    await renderTask.promise
    renderTask = null
  } catch (e) {
    if (e?.name === 'RenderingCancelledException') return
    throw e
  }
}

async function renderPdf() {
  if (!props.src || !canvas.value) return

  loading.value = true
  error.value = false

  try {
    if (renderTask) {
      renderTask.cancel()
      renderTask = null
    }
    if (pdfDoc) {
      pdfDoc.destroy()
      pdfDoc = null
    }

    pdfDoc = await pdfjsLib.getDocument(props.src).promise
    totalPages.value = pdfDoc.numPages
    currentPage.value = 1

    // 等待 DOM 布局完成
    await nextTick()
    await new Promise(resolve => requestAnimationFrame(resolve))

    await renderPage(1)
    loading.value = false
  } catch (e) {
    if (e?.name === 'RenderingCancelledException') return
    console.error('PDF render error:', e)
    error.value = true
    loading.value = false
  }
}

function prevPage() {
  if (currentPage.value <= 1) return
  currentPage.value--
  renderPage(currentPage.value)
}

function nextPage() {
  if (currentPage.value >= totalPages.value) return
  currentPage.value++
  renderPage(currentPage.value)
}

watch(() => props.src, () => {
  if (props.src) {
    nextTick(() => renderPdf())
  }
})

onMounted(() => {
  if (props.src) renderPdf()

  // 监听容器大小变化
  resizeObserver = new ResizeObserver((entries) => {
    const entry = entries[0]
    const { width, height } = entry.contentRect
    // 只在尺寸真正变化时重新渲染
    if (Math.abs(width - lastWidth) > 1 || Math.abs(height - lastHeight) > 1) {
      lastWidth = width
      lastHeight = height
      if (pdfDoc && currentPage.value > 0 && !loading.value) {
        renderPage(currentPage.value)
      }
    }
  })
  if (container.value) {
    resizeObserver.observe(container.value)
  }
})

onUnmounted(() => {
  if (resizeObserver) {
    resizeObserver.disconnect()
    resizeObserver = null
  }
  if (renderTask) {
    renderTask.cancel()
    renderTask = null
  }
  if (pdfDoc) {
    pdfDoc.destroy()
    pdfDoc = null
  }
})
</script>
