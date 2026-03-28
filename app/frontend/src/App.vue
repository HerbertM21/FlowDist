<script lang="ts" setup>
import { ref, computed, provide, onMounted, onBeforeUnmount } from 'vue'
import { GetCheques, GetChequesDevueltos } from '../wailsjs/go/main/App'
import { EventsOn } from '../wailsjs/runtime/runtime'
import ChequesView from './components/ChequesView.vue'
import ChequesDevueltosView from './components/ChequesDevueltosView.vue'
import CabinetsView from './components/CabinetsView.vue'
import { refreshBusKey } from './sync/refreshBus'
import bannerNegro from './assets/images/banner_negro.png'

const activeView = ref('home')
const stats = ref({ cheques: 0, devueltos: 0, montoTotal: 0 })
const refreshTick = ref(0)
const chequesMenuOpen = ref(true)
const isChequesSectionActive = computed(() => activeView.value === 'cheques' || activeView.value === 'devueltos')

type DBUpdatedPayload = {
  source: 'listen-notify' | 'fallback-poll'
  table: string
  operation?: 'INSERT' | 'UPDATE' | 'DELETE'
  id?: number
  at?: string
  raw?: string
}

function notifyRefresh() {
  refreshTick.value += 1
}

provide(refreshBusKey, {
  refreshTick,
  notifyRefresh,
})

async function loadStats() {
  try {
    const [cheques, devueltos] = await Promise.all([GetCheques(), GetChequesDevueltos()])
    const ch = cheques || []
    const dv = devueltos || []
    stats.value = {
      cheques: ch.length,
      devueltos: dv.length,
      montoTotal: ch.reduce((s: number, c: any) => s + (c.monto || 0), 0),
    }
  } catch (e) {
    console.error(e)
  }
}

function formatMoney(amount: number) {
  return new Intl.NumberFormat('es-CL', { style: 'currency', currency: 'CLP', minimumFractionDigits: 0 }).format(amount)
}

function navigateTo(view: string) {
  activeView.value = view
  if (view === 'cheques' || view === 'devueltos') {
    chequesMenuOpen.value = true
  }
  if (view === 'home') loadStats()
}

function toggleChequesMenu() {
  chequesMenuOpen.value = !chequesMenuOpen.value
}

let disposeDBUpdated: (() => void) | null = null

onMounted(() => {
  loadStats()

  disposeDBUpdated = EventsOn('db-updated', (payload?: DBUpdatedPayload) => {
    if (payload?.table && payload.table !== 'cheques' && payload.table !== 'cheques_devueltos' && payload.table !== 'movimientos_cabinets') return

    notifyRefresh()

    if (activeView.value === 'home') {
      loadStats()
    }
  })
})

onBeforeUnmount(() => {
  if (disposeDBUpdated) {
    disposeDBUpdated()
    disposeDBUpdated = null
  }
})
</script>

<template>
  <div class="app-shell">
    <aside class="sidebar">
      <div class="sidebar-brand" style="--wails-draggable: drag;">
        <img :src="bannerNegro" alt="FlowDist" class="brand-logo" draggable="false" />
      </div>

      <nav class="sidebar-nav">
        <span class="nav-section-label">Menú</span>
        <button
          :class="['nav-item', { active: activeView === 'home' }]"
          @click="navigateTo('home')"
        >
          <svg class="nav-icon" width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="m3 9 9-7 9 7v11a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2z"/><polyline points="9 22 9 12 15 12 15 22"/></svg>
          <span class="nav-item-text">Inicio</span>
        </button>

        <div :class="['nav-group', { active: isChequesSectionActive }]">
          <button
            :class="['nav-item', 'nav-parent', { active: isChequesSectionActive }]"
            @click="toggleChequesMenu"
          >
            <svg class="nav-icon" width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z"/><polyline points="14 2 14 8 20 8"/><line x1="16" y1="13" x2="8" y2="13"/><line x1="16" y1="17" x2="8" y2="17"/><polyline points="10 9 9 9 8 9"/></svg>
            <span class="nav-item-text">Cheques</span>
            <svg :class="['nav-caret', { open: chequesMenuOpen }]" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><polyline points="6 9 12 15 18 9"/></svg>
          </button>

          <Transition name="nav-collapse">
            <div v-if="chequesMenuOpen" class="nav-submenu">
              <button
                :class="['nav-subitem', { active: activeView === 'cheques' }]"
                @click="navigateTo('cheques')"
              >
                <span class="nav-subdot"></span>
                <span>Cheques</span>
              </button>
              <button
                :class="['nav-subitem', { active: activeView === 'devueltos' }]"
                @click="navigateTo('devueltos')"
              >
                <span class="nav-subdot"></span>
                <span>Protestados</span>
              </button>
            </div>
          </Transition>
        </div>

        <button
          :class="['nav-item', { active: activeView === 'cabinets' }]"
          @click="navigateTo('cabinets')"
        >
          <svg class="nav-icon" width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><rect x="3" y="4" width="18" height="4" rx="1"/><path d="M5 8h14v11a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2V8Z"/><path d="M10 12h4"/></svg>
          <span class="nav-item-text">Control Cabinets</span>
        </button>
      </nav>
    </aside>

    <main class="main-area">
      <Transition name="fade" mode="out-in">
        <div v-if="activeView === 'home'" class="home-view" key="home">
          <header class="home-header">
            <div>
              <h1 class="home-title">Bienvenido de vuelta</h1>
              <p class="home-subtitle">Resumen general del estado de cobranzas</p>
            </div>
          </header>

          <div class="stats-grid">
            <div class="stat-card" @click="navigateTo('cheques')">
              <div class="stat-card-icon teal">
                <svg width="22" height="22" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z"/><polyline points="14 2 14 8 20 8"/></svg>
              </div>
              <div class="stat-card-body">
                <span class="stat-card-label">Cheques Registrados</span>
                <span class="stat-card-value">{{ stats.cheques }}</span>
              </div>
              <svg class="stat-card-arrow" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><polyline points="9 18 15 12 9 6"/></svg>
            </div>

            <div class="stat-card" @click="navigateTo('devueltos')">
              <div class="stat-card-icon amber">
                <svg width="22" height="22" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="m21.73 18-8-14a2 2 0 0 0-3.48 0l-8 14A2 2 0 0 0 4 21h16a2 2 0 0 0 1.73-3Z"/><line x1="12" y1="9" x2="12" y2="13"/><line x1="12" y1="17" x2="12.01" y2="17"/></svg>
              </div>
              <div class="stat-card-body">
                <span class="stat-card-label">Cheques Devueltos</span>
                <span class="stat-card-value">{{ stats.devueltos }}</span>
              </div>
              <svg class="stat-card-arrow" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><polyline points="9 18 15 12 9 6"/></svg>
            </div>

            <div class="stat-card wide">
              <div class="stat-card-icon green">
                <svg width="22" height="22" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M12 2v20M17 5H9.5a3.5 3.5 0 0 0 0 7h5a3.5 3.5 0 0 1 0 7H6"/></svg>
              </div>
              <div class="stat-card-body">
                <span class="stat-card-label">Monto Total en Cartera</span>
                <span class="stat-card-value large">{{ formatMoney(stats.montoTotal) }}</span>
              </div>
            </div>
          </div>

          <div class="quick-actions">
            <h3 class="section-label">Accesos Rapidos</h3>
            <div class="actions-row">
              <button class="action-card" @click="navigateTo('cheques')">
                <div class="action-icon">
                  <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="var(--primary)" stroke-width="2"><line x1="12" y1="5" x2="12" y2="19"/><line x1="5" y1="12" x2="19" y2="12"/></svg>
                </div>
                <span class="action-label">Registrar Cheque</span>
                <span class="action-desc">Ingresar un nuevo cheque</span>
              </button>
              <button class="action-card" @click="navigateTo('cheques')">
                <div class="action-icon">
                  <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="var(--primary)" stroke-width="2"><rect x="3" y="3" width="18" height="18" rx="2" ry="2"/><line x1="3" y1="9" x2="21" y2="9"/><line x1="9" y1="21" x2="9" y2="9"/></svg>
                </div>
                <span class="action-label">Ver Cheques</span>
                <span class="action-desc">Gestionar estados y cartera</span>
              </button>
              <button class="action-card" @click="navigateTo('devueltos')">
                <div class="action-icon">
                  <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="var(--primary)" stroke-width="2"><path d="m21.73 18-8-14a2 2 0 0 0-3.48 0l-8 14A2 2 0 0 0 4 21h16a2 2 0 0 0 1.73-3Z"/><line x1="12" y1="9" x2="12" y2="13"/><line x1="12" y1="17" x2="12.01" y2="17"/></svg>
                </div>
                <span class="action-label">Cheques Devueltos</span>
                <span class="action-desc">Registrar y saldar devoluciones</span>
              </button>
            </div>
          </div>
        </div>

        <ChequesView v-else-if="activeView === 'cheques'" key="cheques" />
        <ChequesDevueltosView v-else-if="activeView === 'devueltos'" key="devueltos" />
        <CabinetsView v-else-if="activeView === 'cabinets'" key="cabinets" />
      </Transition>
    </main>
  </div>
</template>

<style>
.app-shell {
  display: flex;
  height: 100vh;
  width: 100%;
  overflow: hidden;
}

.sidebar {
  width: 230px;
  background: var(--bg-sidebar);
  display: flex;
  flex-direction: column;
  flex-shrink: 0;
  border-right: 1px solid rgba(255,255,255,0.04);
}
.sidebar-brand {
  padding: 16px 12px 14px;
}
.brand-logo {
  width: 100%;
  height: auto;
  display: block;
  user-select: none;
}

.sidebar-nav {
  flex: 1;
  padding: 8px 12px;
  display: flex;
  flex-direction: column;
  gap: 2px;
}
.nav-section-label {
  font-size: 10px;
  font-weight: 700;
  text-transform: uppercase;
  letter-spacing: 0.1em;
  color: var(--text-sidebar);
  opacity: 0.5;
  padding: 12px 12px 6px;
}
.nav-item {
  display: flex;
  align-items: center;
  gap: 10px;
  width: 100%;
  padding: 9px 12px;
  border-radius: var(--radius-sm);
  font-size: 13.5px;
  font-weight: 500;
  color: var(--text-sidebar);
  background: transparent;
  transition: all var(--transition);
  text-align: left;
}
.nav-item:hover {
  color: var(--text-white);
  background: var(--bg-sidebar-hover);
}
.nav-item.active {
  color: var(--text-sidebar-active);
  background: var(--bg-sidebar-active);
}
.nav-item.active .nav-icon,
.nav-item.active .nav-caret {
  color: var(--text-sidebar-active);
}
.nav-item-text {
  flex: 1;
}
.nav-icon {
  flex-shrink: 0;
}

.nav-group {
  border-radius: 12px;
}
.nav-caret {
  margin-left: auto;
  opacity: 0.9;
  transform: rotate(-90deg);
  transition: transform var(--transition-slow);
}
.nav-caret.open {
  transform: rotate(0deg);
}

.nav-submenu {
  display: flex;
  flex-direction: column;
  gap: 2px;
  padding: 0 0 4px 32px;
}
.nav-subitem {
  width: 100%;
  display: flex;
  align-items: center;
  gap: 8px;
  text-align: left;
  font-size: 12.6px;
  color: var(--text-sidebar);
  padding: 7px 10px;
  border-radius: var(--radius-sm);
  background: transparent;
  transition: all var(--transition);
}
.nav-subitem:hover {
  color: var(--text-white);
  background: var(--bg-sidebar-hover);
}
.nav-subitem.active {
  color: var(--text-sidebar-active);
  background: var(--bg-sidebar-active);
}
.nav-subdot {
  width: 6px;
  height: 6px;
  border-radius: 50%;
  background: var(--text-sidebar);
}
.nav-subitem.active .nav-subdot {
  background: var(--text-sidebar-active);
}

.nav-collapse-enter-active,
.nav-collapse-leave-active {
  transition: max-height 180ms ease, opacity 140ms ease, transform 180ms ease;
  transform-origin: top;
  overflow: hidden;
}
.nav-collapse-enter-from,
.nav-collapse-leave-to {
  max-height: 0;
  opacity: 0;
  transform: translateY(-4px);
}
.nav-collapse-enter-to,
.nav-collapse-leave-from {
  max-height: 120px;
  opacity: 1;
  transform: translateY(0);
}

.main-area {
  flex: 1;
  overflow: hidden;
  display: flex;
  flex-direction: column;
}

.home-view {
  padding: 32px 36px;
  overflow-y: auto;
  height: 100%;
}
.home-header { margin-bottom: 28px; }
.home-title {
  font-size: 26px;
  font-weight: 700;
  color: var(--text-primary);
  letter-spacing: -0.03em;
}
.home-subtitle {
  font-size: 14px;
  color: var(--text-secondary);
  margin-top: 4px;
}

.stats-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(360px, 1fr));
  gap: 14px;
  margin-bottom: 32px;
}
.stat-card {
  display: flex;
  align-items: center;
  gap: 16px;
  padding: 22px 24px;
  background: var(--bg-white);
  border: 1px solid var(--border-light);
  border-radius: var(--radius-lg);
  box-shadow: var(--shadow-xs);
  cursor: pointer;
  transition: all var(--transition);
}
.stat-card:hover {
  box-shadow: var(--shadow-md);
  border-color: var(--border-hover);
  transform: translateY(-2px);
}
.stat-card.wide {
  grid-column: 1 / -1;
  cursor: default;
}
.stat-card.wide:hover { transform: none; }
.stat-card-icon {
  width: 46px;
  height: 46px;
  border-radius: var(--radius-md);
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}
.stat-card-icon.teal { background: var(--primary-light); color: var(--primary); }
.stat-card-icon.amber { background: var(--warning-bg); color: var(--warning); }
.stat-card-icon.green { background: var(--success-bg); color: var(--success); }
.stat-card-body { flex: 1; }
.stat-card-label {
  display: block;
  font-size: 12px;
  color: var(--text-muted);
  font-weight: 500;
  letter-spacing: 0.02em;
}
.stat-card-value {
  display: block;
  font-size: 28px;
  font-weight: 700;
  color: var(--text-primary);
  letter-spacing: -0.03em;
  line-height: 1.2;
  margin-top: 2px;
}
.stat-card-value.large { font-size: 32px; }
.stat-card-arrow { color: var(--text-muted); flex-shrink: 0; }

.section-label {
  font-size: 11px;
  font-weight: 700;
  text-transform: uppercase;
  letter-spacing: 0.08em;
  color: var(--text-muted);
  margin-bottom: 12px;
}

.actions-row {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(280px, 1fr));
  gap: 14px;
}
.action-card {
  display: flex;
  flex-direction: column;
  align-items: flex-start;
  gap: 10px;
  padding: 22px;
  background: var(--bg-white);
  border: 1px solid var(--border-light);
  border-radius: var(--radius-lg);
  box-shadow: var(--shadow-xs);
  text-align: left;
  transition: all var(--transition);
}
.action-card:hover {
  box-shadow: var(--shadow-md);
  border-color: var(--primary);
  transform: translateY(-2px);
}
.action-icon {
  width: 40px;
  height: 40px;
  border-radius: var(--radius-sm);
  background: var(--primary-light);
  display: flex;
  align-items: center;
  justify-content: center;
}
.action-label {
  font-size: 14px;
  font-weight: 600;
  color: var(--text-primary);
}
.action-desc {
  font-size: 12px;
  color: var(--text-muted);
  line-height: 1.4;
}
</style>
