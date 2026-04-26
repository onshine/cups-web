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
        <input ref="fileInput" type="file" class="hidden" multiple @change="onFileChange" />
        <div v-if="!selectedFile && !displayName">
          <UIcon name="i-lucide-upload-cloud" class="w-8 h-8 sm:w-10 sm:h-10 mx-auto text-muted mb-2" />
          <p class="text-sm text-muted">点击或拖拽文件上传</p>
          <p class="text-xs text-muted mt-1">支持 PDF、Word、Excel、PPT、OFD、图片等格式（可多选图片）</p>
        </div>
        <div v-else class="flex items-center gap-3 w-full">
          <UIcon name="i-lucide-file-check" class="w-8 h-8 text-success shrink-0" />
          <div class="flex-1 min-w-0 text-left">
            <p class="text-sm font-medium break-all line-clamp-2 leading-snug">{{ displayName || (selectedFile && selectedFile.name) }}</p>
            <p v-if="selectedFile" class="text-xs text-muted mt-0.5">{{ formatFileSize(selectedFile.size) }}</p>
            <p v-else-if="isMultiImage && totalSize > 0" class="text-xs text-muted mt-0.5">共 {{ formatFileSize(totalSize) }}</p>
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
      <UAlert v-if="converting" color="info" variant="subtle" icon="i-lucide-loader-circle" :title="props.isMultiImage ? '正在合并图片为 PDF，请稍候…' : '正在转换为 PDF，请稍候…'" />
      <UAlert v-if="converted && !converting" color="success" variant="subtle" icon="i-lucide-check-circle" title="已转换为 PDF，可以打印" />

      <!-- 操作按钮 -->
      <div class="flex flex-wrap gap-2">
        <UButton
          v-if="canConvert"
          variant="outline"
          icon="i-lucide-file-text"
          :loading="converting"
          @click="$emit('convert')"
        >{{ props.isMultiImage ? '合并图片为PDF' : '转换为 PDF' }}</UButton>
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
  displayName: { type: String, default: '' },
  converting: { type: Boolean, default: false },
  converted: { type: Boolean, default: false },
  previewUrl: { type: String, default: '' },
  downloadName: { type: String, default: '' },
  pdfBlob: { type: [Blob, null], default: null },
  canConvert: { type: Boolean, default: false },
  canPrint: { type: Boolean, default: false },
  printing: { type: Boolean, default: false },
  isMultiImage: { type: Boolean, default: false },
  totalSize: { type: Number, default: 0 }
})

const emit = defineEmits(['file-selected', 'files-selected', 'clear', 'convert', 'print', 'download'])

const isDragging = ref(false)
const fileInput = ref(null)

function handleFiles(files) {
  if (!files || files.length === 0) return
  if (files.length === 1) {
    emit('file-selected', files[0])
    return
  }
  // 多文件：过滤出图片
  const images = Array.from(files).filter(f => f.type.startsWith('image/'))
  if (images.length === 0) {
    // 无图片，取第一个文件走单文件流程
    emit('file-selected', files[0])
  } else if (images.length === 1) {
    emit('file-selected', images[0])
  } else {
    emit('files-selected', images)
  }
}

function onDrop(e) {
  isDragging.value = false
  handleFiles(e.dataTransfer.files)
}

function onFileChange(e) {
  handleFiles(e.target.files)
  // 重置 input 以便可以重新选择相同文件
  if (fileInput.value) fileInput.value.value = ''
}
</script>
