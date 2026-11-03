<template>
  <v-card subtitle="L2TP/IPsec">
    <v-alert
      type="info"
      variant="tonal"
      class="my-3"
    >
      {{ $t('l2tp.hint') }}
    </v-alert>
    <v-row>
      <v-col
        cols="12"
        sm="6"
        md="4"
      >
        <v-text-field
          v-model="data.listen"
          :label="$t('l2tp.listen')"
          hide-details
        />
      </v-col>
      <v-col
        cols="12"
        sm="6"
        md="4"
      >
        <v-text-field
          v-model.number="data.listen_port"
          :label="$t('l2tp.ikePort')"
          type="number"
          min="1"
          max="65535"
          hide-details
        />
      </v-col>
      <v-col
        cols="12"
        sm="6"
        md="4"
      >
        <v-text-field
          v-model="data.public_ip"
          :label="$t('l2tp.publicIp')"
          hide-details
        />
      </v-col>
      <v-col
        cols="12"
        sm="6"
        md="4"
      >
        <v-text-field
          v-model="data.psk"
          :label="$t('l2tp.psk')"
          :type="showPsk ? 'text' : 'password'"
          :append-inner-icon="showPsk ? 'mdi-eye-off' : 'mdi-eye'"
          hide-details
          @click:append-inner="showPsk = !showPsk"
        />
      </v-col>
      <v-col
        cols="12"
        sm="6"
        md="4"
      >
        <v-text-field
          v-model="data.pool"
          :label="$t('l2tp.pool')"
          hide-details
        />
      </v-col>
    </v-row>
  </v-card>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'
import type { Inbound, L2TP } from '@/types/inbounds'

const model = defineModel<Inbound>('data', { required: true })
const data = computed(() => model.value as L2TP)
const showPsk = ref(false)
</script>
