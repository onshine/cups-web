<template>
  <UCard>
    <template #header>
      <div class="flex items-center justify-between">
        <div class="flex items-center gap-2 font-semibold">
          <UIcon name="i-lucide-activity" class="w-5 h-5" />
          打印机状态
        </div>
        <UButton variant="ghost" size="xs" icon="i-lucide-refresh-cw" @click="$emit('refresh')" :loading="loading" />
      </div>
    </template>
    <div>
      <div v-if="!printerUri" class="text-center py-6 text-muted text-sm">
        请先选择打印机
      </div>
      <div v-else-if="loading && !printerInfo" class="text-center py-4">
        <UIcon name="i-lucide-loader-circle" class="w-5 h-5 animate-spin mx-auto text-muted" />
      </div>
      <div v-else-if="error" class="text-center py-4 text-sm text-error">
        <UIcon name="i-lucide-wifi-off" class="w-5 h-5 mx-auto mb-1" />
        {{ error }}
      </div>
      <div v-else-if="printerInfo" class="space-y-3">
        <!-- 基本状态 -->
        <div class="flex items-center justify-between p-2 bg-elevated rounded-lg">
          <div class="flex items-center gap-2">
            <UIcon name="i-lucide-info" class="w-4 h-4 text-info" />
            <span class="text-sm font-medium">打印机状态</span>
          </div>
          <UBadge :color="printerStateColor(printerInfo.state)" variant="subtle" size="xs">
            {{ printerStateText(printerInfo.state) }}
          </UBadge>
        </div>

        <!-- 队列 -->
        <div class="flex items-center justify-between p-2 bg-elevated rounded-lg">
          <div class="flex items-center gap-2">
            <UIcon name="i-lucide-list-ordered" class="w-4 h-4 text-primary" />
            <span class="text-sm font-medium">队列任务数</span>
          </div>
          <span class="text-sm font-bold">{{ printerInfo.queuedJobs }}</span>
        </div>

        <!-- 状态持续时间 -->
        <div v-if="printerInfo.attributes && printerInfo.attributes['printer-state-change-date-time']" class="flex items-center justify-between p-2 bg-elevated rounded-lg">
          <div class="flex items-center gap-2">
            <UIcon name="i-lucide-clock" class="w-4 h-4 text-success" />
            <span class="text-sm font-medium">状态持续</span>
          </div>
          <span class="text-sm">{{ formatStateDuration(printerInfo.attributes['printer-state-change-date-time']) }}</span>
        </div>

        <!-- 固件版本 -->
        <div v-if="printerInfo.firmwareVersion" class="flex items-center justify-between p-2 bg-elevated rounded-lg">
          <div class="flex items-center gap-2">
            <UIcon name="i-lucide-cpu" class="w-4 h-4 text-secondary" />
            <span class="text-sm font-medium">固件版本</span>
          </div>
          <span class="text-xs text-muted truncate max-w-32">{{ printerInfo.firmwareVersion }}</span>
        </div>

        <!-- 状态消息 -->
        <div v-if="printerInfo.stateMessage" class="p-2 bg-warning/10 border border-warning/20 rounded-lg">
          <p class="text-xs text-warning">{{ printerInfo.stateMessage }}</p>
        </div>

        <!-- 墨盒信息 -->
        <div v-if="printerInfo.markerNames && printerInfo.markerNames.length > 0">
          <div class="flex items-center gap-2 mb-2">
            <UIcon name="i-lucide-droplets" class="w-4 h-4 text-primary" />
            <span class="text-sm font-semibold">墨盒信息</span>
          </div>
          <div class="space-y-2">
            <div v-for="(name, i) in printerInfo.markerNames" :key="i" class="space-y-1">
              <div class="flex justify-between text-xs">
                <span class="text-muted">{{ name }}</span>
                <span :class="markerLevelColor(printerInfo.markerLevels?.[i])">
                  {{ printerInfo.markerLevels?.[i] ?? '?' }}%
                </span>
              </div>
              <div class="w-full bg-muted/30 rounded-full h-2">
                <div
                  class="h-2 rounded-full transition-all"
                  :class="markerBarColor(printerInfo.markerLevels?.[i])"
                  :style="{ width: Math.max(0, Math.min(100, printerInfo.markerLevels?.[i] ?? 0)) + '%' }"
                ></div>
              </div>
            </div>
          </div>
        </div>

        <!-- 纸盒信息 -->
        <div v-if="printerInfo.mediaReady && printerInfo.mediaReady.length > 0">
          <div class="flex items-center gap-2 mb-2">
            <UIcon name="i-lucide-layers" class="w-4 h-4 text-secondary" />
            <span class="text-sm font-semibold">纸盒信息</span>
          </div>
          <div class="space-y-1">
            <div v-for="(media, i) in printerInfo.mediaReady" :key="i"
              class="flex items-center gap-2 p-1.5 bg-elevated rounded-lg text-xs">
              <UIcon name="i-lucide-square" class="w-3 h-3 text-muted" />
              <span>{{ media }}</span>
            </div>
          </div>
        </div>

        <!-- 状态原因 -->
        <div v-if="printerInfo.stateReasons && printerInfo.stateReasons.filter(r => r !== 'none').length > 0">
          <div class="flex items-center gap-2 mb-1">
            <UIcon name="i-lucide-alert-triangle" class="w-4 h-4 text-warning" />
            <span class="text-sm font-semibold">警报</span>
          </div>
          <div class="space-y-1">
            <div v-for="reason in printerInfo.stateReasons.filter(r => r !== 'none')" :key="reason"
              class="text-xs text-warning bg-warning/10 px-2 py-1 rounded-lg">
              {{ reason }}
            </div>
          </div>
        </div>
      </div>
    </div>
  </UCard>
</template>

<script setup>
import { formatStateDuration, printerStateColor, printerStateText, markerLevelColor, markerBarColor } from '../../utils/format'

defineProps({
  printerInfo: { type: Object, default: null },
  printerUri: { type: String, default: '' },
  loading: { type: Boolean, default: false },
  error: { type: String, default: '' }
})

defineEmits(['refresh'])
</script>
