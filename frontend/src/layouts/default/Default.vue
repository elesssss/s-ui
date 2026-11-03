<template>
  <v-app style="overflow: auto;">
    <drawer
      :is-mobile="isMobile"
      :display-drawer="displayDrawer"
      @toggle-drawer="toggleDrawer"
    />
    <default-bar
      :is-mobile="isMobile"
      @toggle-drawer="toggleDrawer"
    />
    <default-view />
    <update-modal v-model="updateVisible" />
  </v-app>
</template>

<script lang="ts" setup>
import { computed, onMounted, ref, watch } from 'vue'
import DefaultBar from './AppBar.vue'
import Drawer from './Drawer.vue'
import DefaultView from './View.vue'
import { useDisplay } from 'vuetify'
import Data from '@/store/modules/data'
import UpdateModal from '@/layouts/modals/Update.vue'
import { storeToRefs } from 'pinia'

const { smAndDown } = useDisplay()
const displayDrawer = ref(false)

const toggleDrawer = () => {
  displayDrawer.value = !displayDrawer.value
}

const isMobile = computed((): boolean => smAndDown.value)

// keep the drawer open on desktop and closed on mobile
watch(smAndDown, (v) => { displayDrawer.value = !v }, { immediate: true })

const { updateVisible } = storeToRefs(Data())

onMounted(() => { Data().checkUpdate() })
</script>

<style>
.v-card-subtitle {
  text-align: center;
  border-bottom: 1px solid gray;
  min-height: 20px;
}
.v-switch.v-input {
  padding-inline-start: .6rem;
}
</style>