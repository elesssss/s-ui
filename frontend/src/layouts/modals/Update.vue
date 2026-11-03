<template>
  <v-dialog
    v-model="visible"
    transition="dialog-bottom-transition"
    width="90%"
    max-width="550"
    :persistent="state == 'restarting'"
  >
    <v-card class="rounded-lg">
      <v-card-title>
        <v-row>
          <v-col>{{ $t('update.title') }}</v-col>
          <v-spacer />
          <v-col
            v-if="state != 'restarting'"
            cols="auto"
          >
            <v-icon
              :aria-label="$t('actions.close')"
              icon="mdi-close"
              @click="visible = false"
            />
          </v-col>
        </v-row>
      </v-card-title>
      <v-divider />
      <v-card-text v-if="update">
        <p class="mb-4">
          <v-chip
            density="compact"
            color="primary"
          >
            v{{ update.current }}
          </v-chip>
          <v-icon
            :icon="$vuetify.locale.isRtl ? 'mdi-arrow-left' : 'mdi-arrow-right'"
            class="mx-2"
          />
          <v-chip
            density="compact"
            color="info"
            variant="flat"
          >
            v{{ update.latest }}
          </v-chip>
          <a
            :href="update.url"
            target="_blank"
            rel="noopener"
            class="ms-4"
          >{{ $t('update.notes') }}</a>
        </p>

        <template v-if="update.method == 'docker'">
          <v-alert
            type="warning"
            variant="tonal"
            :text="$t('update.dockerHint')"
          />
          <pre class="command">docker compose pull &amp;&amp; docker compose up -d</pre>
        </template>

        <template v-else-if="update.method == 'systemd'">
          <v-alert
            v-if="state == 'idle'"
            type="info"
            variant="tonal"
            :text="$t('update.systemdHint')"
          />
          <div
            v-else-if="state == 'restarting'"
            class="text-center"
          >
            <v-progress-circular
              indeterminate
              color="primary"
              class="mb-2"
            />
            <div>{{ $t('update.restarting') }}</div>
          </div>
          <v-alert
            v-else-if="state == 'timeout'"
            type="error"
            variant="tonal"
            :text="$t('update.timeout')"
          />
        </template>

        <template v-else>
          <v-alert
            type="info"
            variant="tonal"
            :text="$t('update.manualHint')"
          />
          <pre class="command">s-ui update</pre>
        </template>
      </v-card-text>
      <v-card-actions v-if="update?.method == 'systemd' && state == 'idle'">
        <v-spacer />
        <v-btn
          color="primary"
          variant="flat"
          prepend-icon="mdi-update"
          :loading="loading"
          @click="upgrade"
        >
          {{ $t('update.now') }}
        </v-btn>
      </v-card-actions>
    </v-card>
  </v-dialog>
</template>

<script setup lang="ts">
import { computed, onBeforeUnmount, ref } from 'vue'
import { i18n } from '@/locales'
import api from '@/plugins/api'
import HttpUtils from '@/plugins/httputil'
import Data from '@/store/modules/data'
import type { UpdateInfo } from '@/store/modules/data'

const visible = defineModel<boolean>({ required: true })

const update = computed(() => Data().update)
const loading = ref(false)
const state = ref<'idle' | 'restarting' | 'timeout'>('idle')

// How long the download, migration and restart may take before the panel
// gives up waiting and tells the operator to look at the server.
const restartTimeout = 5 * 60 * 1000
let pollId: ReturnType<typeof setInterval> | null = null

const upgrade = async () => {
  if (!update.value || !confirm(i18n.global.t('update.confirm'))) return
  loading.value = true
  const msg = await HttpUtils.post('api/update', null)
  loading.value = false
  if (msg.success) waitForRestart(update.value.current)
}

// Polled with the bare client: while the panel is down every request fails,
// and HttpUtils would raise each failure as an error.
const waitForRestart = (oldVersion: string) => {
  state.value = 'restarting'
  const started = Date.now()
  pollId = setInterval(async () => {
    if (Date.now() - started > restartTimeout) {
      stopPolling()
      state.value = 'timeout'
      return
    }
    try {
      const resp = await api.get<{ success: boolean, obj: UpdateInfo }>('api/checkUpdate')
      if (resp.data?.success && resp.data.obj?.current && resp.data.obj.current != oldVersion) {
        stopPolling()
        window.location.reload()
      }
    } catch {
      // Still restarting.
    }
  }, 3000)
}

const stopPolling = () => {
  if (pollId) {
    clearInterval(pollId)
    pollId = null
  }
}

onBeforeUnmount(stopPolling)
</script>

<style scoped>
.command {
  direction: ltr;
  margin-top: 12px;
  padding: 8px 12px;
  border-radius: 4px;
  background: rgba(var(--v-theme-on-surface), 0.08);
  white-space: pre-wrap;
  user-select: all;
}
</style>
