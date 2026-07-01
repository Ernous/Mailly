import { createApp } from 'vue'
import { createVuetify } from 'vuetify'
import 'vuetify/styles'

// Only import components actually used in the app
import {
  VApp,
  VBtn,
  VIcon,
  VAvatar,
  VImg,
  VCard,
  VCardText,
  VCardActions,
  VDialog,
  VDivider,
  VList,
  VListItem,
  VListItemTitle,
  VMenu,
  VProgressCircular,
  VProgressLinear,
  VSelect,
  VSpacer,
  VTextarea,
  VTextField,
  VExpandTransition,
  VAlert,
} from 'vuetify/components'

import {
  Ripple,
} from 'vuetify/directives'


import App from './App.vue'
import router from './router'

const theme = {
  dark: true,
  colors: {
    primary: '#4d8080',
    secondary: '#607D8B',
    accent: '#56b04c',
    error: '#f44336',
    info: '#2196f3',
    success: '#56b04c',
    warning: '#ff9800',
    surface: '#1e1e1e',
    background: '#121212',
    'surface-bright': '#2a2a2a',
    'surface-variant': '#252525',
    'on-surface': '#e0e0e0',
    'on-surface-variant': '#9e9e9e',
    outline: '#333333',
  },
}

const vuetify = createVuetify({
  components: {
    VApp,
    VBtn,
    VIcon,
    VAvatar,
    VImg,
    VCard,
    VCardText,
    VCardActions,
    VDialog,
    VDivider,
    VList,
    VListItem,
    VListItemTitle,
    VMenu,
    VProgressCircular,
    VProgressLinear,
    VSelect,
    VSpacer,
    VTextarea,
    VTextField,
    VExpandTransition,
    VAlert,
  },
  directives: { Ripple },
  theme: {
    defaultTheme: 'theme',
    themes: { theme },
  },
  defaults: {
    VBtn: { variant: 'flat', rounded: 'lg' },
    VCard: { rounded: 'lg' },
    VTextField: { variant: 'outlined', density: 'compact', hideDetails: true },
    VSelect: { variant: 'outlined', density: 'compact', hideDetails: true },
  },
})

createApp(App).use(vuetify).use(router).mount('#app')
