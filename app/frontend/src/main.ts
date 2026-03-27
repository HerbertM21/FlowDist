import {createApp} from 'vue'
import App from './App.vue'
import './style.css';

window.addEventListener('unhandledrejection', (event) => {
  console.error('Unhandled promise rejection:', event.reason)
})

window.addEventListener('error', (event) => {
  console.error('Window error:', event.error || event.message)
})

createApp(App).mount('#app')
