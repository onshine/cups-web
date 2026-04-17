<template>
  <UCard>
    <template #header>
      <div class="flex items-center gap-2 font-semibold">
        <UIcon name="i-lucide-file-up" class="w-5 h-5" />
        文件
      </div>
    </template>
    <div class="space-y-3">
      <div
        class="border-2 border-dashed rounded-lg p-4 sm:p-6 text-center cursor-pointer transition-colors"
        :class="isDragging ? 'border-primary bg-primary/5' : 'border-muted hover:border-primary/50'"
        @dragover.prevent="isDragging = true"
        @dragleave="isDragging = false"
        @drop.prevent="onDrop"
        @click="fileInput.click()"
      >
        <input ref="fileInput" type="file" class="hidden" @change="onFileChange" />
        <div v-if="!selectedFile">
          <UIcon name="i-lucide-upload-cloud" class="w-8 h-8 sm:w-10 sm:h-10 mx-auto text-muted mb-2" />
          <p class="text-sm text-muted">点击或拖拽文件上传</p>
          <p class="text-xs text-muted mt-1">支持 PDF、Word、Excel、PPT、OFD、图片等格式</p>
        </div>
        <div v-else class="flex items-center gap-3 w-full">
          <UIcon name="i-lucide-file-check" class="w-8 h-8 text-success shrink-0" />
          <div class="flex-1 min-w-0 text-left">
            <p class="text-sm font-medium break-all line-clamp-2 leading-snug">{{ selectedFile.name }}</p>
            <p class="text-xs text-muted mt-0.5">{{ formatFileSize(selectedFile.size) }}</p>
          </div>
          <UButton
            variant="ghost"
            size="xs"
            icon="i-lucide-x"
            color="error"
            class="shrink-0"
            @click.stop="$emit('clear')"
          />
        </div>
      </div>

      <!-- 转换状态 -->
      <UAlert v-if="converting" color="info" variant="subtle" icon="i-lucide-loader-circle" title="正在转换为 PDF，请稍候…" />
      <UAlert v-if="converted && !converting" color="success" variant="subtle" icon="i-lucide-check-circle" title="已转换为 PDF，可以打印" />

      <!-- 操作按钮 -->
      <div class="flex flex-wrap gap-2">
        <UButton
          v-if="canConvert"
          variant="outline"
          icon="i-lucide-file-text"
          :loading="converting"
          @click="$emit('convert')"
        >转换为 PDF</UButton>
        <UButton
          v-if="previewUrl"
          variant="ghost"
          icon="i-lucide-download"
          :href="previewUrl"
          :download="downloadName"
          tag="a"
        >下载预览</UButton>
      </div>
    </div>
  </UCard>
</template>

<script setup>
import { ref } from 'vue'
import { formatFileSize } from '../../utils/format'

const props = defineProps({
  selectedFile: { type: [File, null], default: null },
  converting: { type: Boolean, default: false },
  converted: { type: Boolean, default: false },
  previewUrl: { type: String, default: '' },
  downloadName: { type: String, default: '' },
  pdfBlob: { type: [Blob, null], default: null },
  canConvert: { type: Boolean, default: false },
  canPrint: { type: Boolean, default: false },
  printing: { type: Boolean, default: false }
})

const emit = defineEmits(['file-selected', 'clear', 'convert', 'print', 'download'])

const isDragging = ref(false)
const fileInput = ref(null)

function onDrop(e) {
  isDragging.value = false
  const f = e.dataTransfer.files[0]
  if (f) emit('file-selected', f)
}

function onFileChange(e) {
  const f = e.target.files[0]
  if (f) emit('file-selected', f)
}
</script>
