<script lang="ts" setup>
import { ref, shallowRef, computed, onMounted, inject, watch } from 'vue'
import { CrearMovimientoCabinet, ExportCabinetsExcel, GetMovimientosCabinets, SoftDeleteMovimientoCabinet, UpdateMovimientoCabinet } from '../../wailsjs/go/main/App'
import { refreshBusKey } from '../sync/refreshBus'

const movimientos = shallowRef<any[]>([])
const loading = ref(true)
const hasLoadedOnce = ref(false)
const searchQuery = ref('')
const showModal = ref(false)
const editCodigoMovimiento = ref('')
const saving = ref(false)
const exporting = ref(false)
const formError = ref('')
const refreshBus = inject(refreshBusKey, null)

const filtered = computed(() => {
  const q = searchQuery.value.trim().toLowerCase()
  if (!q) {
    return movimientos.value
  }

  return movimientos.value.filter(m =>
    String(m.nombre_cliente || '').toLowerCase().includes(q)
  )
})

const totalMaquinasMes = computed(() => {
  const now = new Date()
  const yyyy = now.getFullYear()
  const mm = now.getMonth() + 1

  return (movimientos.value || []).reduce((sum, item) => {
    const fecha = item.fecha_entrada || item.created_at
    if (!fecha) return sum

    const dt = new Date(fecha)
    if (Number.isNaN(dt.getTime())) return sum

    if (dt.getFullYear() === yyyy && dt.getMonth() + 1 === mm) {
      return sum + (Number(item.cantidad_cabinets) || 0)
    }

    return sum
  }, 0)
})

const form = ref({
  nombre_cliente: '',
  direccion: '',
  localidad: '',
  cantidad_cabinets: '1',
  descripcion: '',
  codigo_movimiento: '',
  fecha_entrada: '',
  fecha_salida: '',
})

function formatMoney(amount: number) {
  return new Intl.NumberFormat('es-CL', { style: 'currency', currency: 'CLP', minimumFractionDigits: 0 }).format(amount)
}

async function loadData(silent = false) {
  const showLoader = !silent && !hasLoadedOnce.value
  if (showLoader) {
    loading.value = true
  }

  try {
    const nextItems = await GetMovimientosCabinets() || []
    if (JSON.stringify(nextItems) !== JSON.stringify(movimientos.value)) {
      movimientos.value = nextItems
    }
  } catch (e) {
    console.error(e)
  } finally {
    if (!hasLoadedOnce.value) {
      loading.value = false
      hasLoadedOnce.value = true
    }
  }
}

function openCreateModal() {
  editCodigoMovimiento.value = ''
  form.value = {
    nombre_cliente: '',
    direccion: '',
    localidad: '',
    cantidad_cabinets: '1',
    descripcion: '',
    codigo_movimiento: '',
    fecha_entrada: '',
    fecha_salida: '',
  }
  formError.value = ''
  showModal.value = true
}

function openEditModal(item: any) {
  editCodigoMovimiento.value = item.codigo_movimiento || ''
  form.value = {
    nombre_cliente: item.nombre_cliente || '',
    direccion: item.direccion || '',
    localidad: item.localidad || '',
    cantidad_cabinets: String(item.cantidad_cabinets || 1),
    descripcion: item.descripcion || '',
    codigo_movimiento: item.codigo_movimiento || '',
    fecha_entrada: item.fecha_entrada || '',
    fecha_salida: item.fecha_salida || '',
  }
  formError.value = ''
  showModal.value = true
}

function closeModal() {
  showModal.value = false
  editCodigoMovimiento.value = ''
}

async function handleSubmit() {
  formError.value = ''

  const nombre = form.value.nombre_cliente.trim()
  const codigoMovimiento = form.value.codigo_movimiento.trim()
  const cantidad = Number(form.value.cantidad_cabinets)

  if (!nombre) {
    formError.value = 'El nombre del cliente es obligatorio'
    return
  }

  if (!codigoMovimiento) {
    formError.value = 'El codigo de movimiento es obligatorio'
    return
  }

  if (!Number.isFinite(cantidad) || cantidad <= 0) {
    formError.value = 'La cantidad de cabinets debe ser mayor a 0'
    return
  }

  saving.value = true
  try {
    const payload = {
      nombre_cliente: nombre,
      direccion: form.value.direccion,
      localidad: form.value.localidad,
      cantidad_cabinets: cantidad,
      descripcion: form.value.descripcion,
      codigo_movimiento: codigoMovimiento,
      fecha_entrada: form.value.fecha_entrada,
      fecha_salida: form.value.fecha_salida,
    }

    if (editCodigoMovimiento.value) {
      await UpdateMovimientoCabinet(editCodigoMovimiento.value, payload as any)
    } else {
      await CrearMovimientoCabinet(payload as any)
    }

    closeModal()
    await loadData(true)
  } catch (e: any) {
    formError.value = e?.message || 'Error al guardar movimiento de cabinet'
  }
  saving.value = false
}

async function deleteItem(codigoMovimiento: string) {
  const confirmed = window.confirm('Eliminar movimiento de cabinet? Esta accion lo ocultara del listado.')
  if (!confirmed) return

  await SoftDeleteMovimientoCabinet(codigoMovimiento)
  await loadData(true)
}

async function exportToExcel() {
  exporting.value = true
  try {
    const result = await ExportCabinetsExcel()
    if (!result) {
      return
    }
    if (result?.file_path) {
      window.alert(`Archivo exportado: ${result.file_path}`)
    }
  } catch (e: any) {
    const msg = typeof e === 'string' ? e : String(e?.message || e || '')
    if (msg.toLowerCase().includes('cancelada')) {
      return
    }
    window.alert(msg || 'Error al exportar archivo Excel')
    console.error(e)
  } finally {
    exporting.value = false
  }
}

onMounted(loadData)

if (refreshBus) {
  watch(refreshBus.refreshTick, () => {
    loadData(true)
  })
}
</script>

<template>
  <div class="cabinets-view">
    <div class="search-filter-row">
      <div class="search-wrapper">
        <svg class="search-icon" width="15" height="15" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><circle cx="11" cy="11" r="8"/><path d="m21 21-4.35-4.35"/></svg>
        <input v-model="searchQuery" type="text" placeholder="Buscar por nombre cliente..." class="search-input" />
      </div>
    </div>
    <header class="view-header">
      <div class="view-header-left">
        <h2 class="view-title">Control Cabinets</h2>
        <div class="view-meta">
          <span class="meta-count">{{ filtered.length }} registros</span>
        </div>
      </div>
      <div class="view-header-right">
        <button class="btn btn-outline" @click="exportToExcel" :disabled="exporting">
          <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4"/><polyline points="7 10 12 15 17 10"/><line x1="12" y1="15" x2="12" y2="3"/></svg>
          {{ exporting ? 'Exportando...' : 'Exportar Excel' }}
        </button>
        <button class="btn btn-primary" @click="openCreateModal">Nuevo Movimiento</button>
      </div>
    </header>

    <div class="resumen-grid">
      <article class="resumen-card">
        <div class="resumen-content">
          <span class="resumen-label">Total de maquinas movidas (mes actual)</span>
          <strong class="resumen-valor">{{ totalMaquinasMes }}</strong>
        </div>
        <div class="resumen-icon icon-primary">
          <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M21 16V8a2 2 0 0 0-1-1.73l-7-4a2 2 0 0 0-2 0l-7 4A2 2 0 0 0 3 8v8a2 2 0 0 0 1 1.73l7 4a2 2 0 0 0 2 0l7-4A2 2 0 0 0 21 16z"/><polyline points="3.27 6.96 12 12.01 20.73 6.96"/><line x1="12" y1="22.08" x2="12" y2="12"/></svg>
        </div>
      </article>
    </div>

    <div class="table-container card">
      <div v-if="loading" class="loading-state">
        <div class="spinner"></div>
        <span>Cargando movimientos...</span>
      </div>
      <table v-else>
        <thead>
          <tr>
            <th>Codigo movimiento</th>
            <th>Cliente</th>
            <th>Direccion</th>
            <th>Localidad</th>
            <th>Cantidad</th>
            <th>Descripcion</th>
            <th>Fecha entrada</th>
            <th>Fecha salida</th>
            <th>Valor</th>
            <th></th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="item in filtered" :key="item.id">
            <td class="cell-mono">{{ item.codigo_movimiento || '-' }}</td>
            <td class="cell-bold">{{ item.nombre_cliente }}</td>
            <td>{{ item.direccion || '-' }}</td>
            <td>{{ item.localidad || '-' }}</td>
            <td>{{ item.cantidad_cabinets }}</td>
            <td class="cell-truncate" :title="item.descripcion || '-'">{{ item.descripcion || '-' }}</td>
            <td>{{ item.fecha_entrada || '-' }}</td>
            <td>{{ item.fecha_salida || '-' }}</td>
            <td class="cell-bold">{{ formatMoney(1) }}</td>
            <td class="cell-actions">
              <button class="btn-ghost-icon" @click="openEditModal(item)">Editar</button>
              <button class="btn-danger-ghost" @click="deleteItem(item.codigo_movimiento)">X</button>
            </td>
          </tr>
          <tr v-if="filtered.length === 0 && !loading">
            <td colspan="10" class="empty-row">No se encontraron movimientos</td>
          </tr>
        </tbody>
      </table>
    </div>

    <Transition name="overlay">
      <div v-if="showModal" class="modal-overlay" @click.self="closeModal"></div>
    </Transition>
    <Transition name="modal">
      <div v-if="showModal" class="modal-panel">
        <div class="modal-header">
          <h3 class="modal-title">{{ editCodigoMovimiento ? 'Editar' : 'Nuevo' }} Movimiento Cabinet</h3>
          <button class="btn-icon" @click="closeModal">X</button>
        </div>

        <form @submit.prevent="handleSubmit" class="modal-form">
          <div class="form-row two">
            <div class="form-field">
              <label>Nombre cliente</label>
              <input v-model="form.nombre_cliente" />
            </div>
            <div class="form-field">
              <label>Direccion</label>
              <input v-model="form.direccion" />
            </div>
          </div>

          <div class="form-row three">
            <div class="form-field">
              <label>Localidad</label>
              <input v-model="form.localidad" />
            </div>
            <div class="form-field">
              <label>Cantidad cabinets</label>
              <input v-model="form.cantidad_cabinets" type="number" min="1" step="1" />
            </div>
            <div class="form-field">
              <label>Codigo movimiento</label>
              <input v-model="form.codigo_movimiento" />
            </div>
          </div>

          <div class="form-row two">
            <div class="form-field">
              <label>Fecha entrada</label>
              <input v-model="form.fecha_entrada" type="date" />
            </div>
            <div class="form-field">
              <label>Fecha salida</label>
              <input v-model="form.fecha_salida" type="date" />
            </div>
          </div>

          <div class="form-row one">
            <div class="form-field">
              <label>Descripcion</label>
              <textarea v-model="form.descripcion" rows="3"></textarea>
            </div>
          </div>

          <div v-if="formError" class="form-error">{{ formError }}</div>

          <div class="modal-actions">
            <button type="button" class="btn btn-ghost" @click="closeModal">Cancelar</button>
            <button type="submit" class="btn btn-primary" :disabled="saving">{{ saving ? 'Guardando...' : 'Guardar' }}</button>
          </div>
        </form>
      </div>
    </Transition>
  </div>
</template>

<style scoped>
.cabinets-view { display: flex; flex-direction: column; height: 100%; padding: 24px 28px; gap: 16px; }
.view-header { display: flex; align-items: center; justify-content: space-between; }
.view-header-left { display: flex; align-items: baseline; gap: 14px; }
.view-title { font-size: 22px; font-weight: 700; letter-spacing: -0.03em; }
.view-meta { display: flex; align-items: center; gap: 10px; }
.meta-count { font-size: 12px; color: var(--text-muted); background: var(--bg-surface); padding: 3px 10px; border-radius: 12px; font-weight: 500; }
.view-header-right { display: flex; align-items: center; gap: 8px; }

.search-filter-row { display: flex; width: 100%; }
.search-wrapper { position: relative; flex: 1; }
.search-icon { position: absolute; left: 11px; top: 50%; transform: translateY(-50%); color: var(--text-muted); pointer-events: none; }
.search-input { padding-left: 32px; width: 100%; }

.resumen-grid {
  display: grid;
  grid-template-columns: minmax(0, 1fr);
  gap: 12px;
}

.resumen-card {
  background: var(--bg-white);
  border: 1px solid var(--border-light);
  border-radius: var(--radius-md);
  padding: 16px 20px;
  display: flex;
  align-items: center;
  justify-content: space-between;
  box-shadow: var(--shadow-xs);
}
.resumen-content {
  display: flex;
  flex-direction: column;
  gap: 4px;
}
.resumen-icon {
  width: 44px;
  height: 44px;
  border-radius: 12px;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}
.icon-primary { background: var(--primary-light); color: var(--primary-dark); }
.icon-success { background: #dcfce7; color: #16a34a; }
.icon-warning { background: #fef08a; color: #ca8a04; }
.icon-danger { background: #fee2e2; color: #dc2626; }

.resumen-label {
  font-size: 12px;
  color: var(--text-muted);
  text-transform: uppercase;
  letter-spacing: 0.06em;
  font-weight: 600;
}

.resumen-valor {
  font-size: 26px;
  letter-spacing: -0.03em;
  color: var(--primary-dark);
}

.table-container { flex: 1; overflow: auto; }
.cell-mono { font-family: var(--font-mono); font-size: 12px; }
.cell-bold { font-weight: 600; }
.cell-truncate { max-width: 180px; white-space: nowrap; overflow: hidden; text-overflow: ellipsis; }
.cell-actions { display: flex; align-items: center; gap: 6px; }

.btn-ghost-icon {
  height: 30px;
  padding: 0 8px;
  border-radius: 8px;
  border: 1px solid var(--border-light);
  background: transparent;
}

.loading-state { display: flex; flex-direction: column; align-items: center; gap: 12px; padding: 60px; color: var(--text-muted); }
.spinner { width: 24px; height: 24px; border: 3px solid var(--border); border-top-color: var(--primary); border-radius: 50%; animation: spin 600ms linear infinite; }
@keyframes spin { to { transform: rotate(360deg); } }
.empty-row { text-align: center; padding: 40px !important; color: var(--text-muted); }

.modal-overlay { position: fixed; inset: 0; background: rgba(0,0,0,0.5); z-index: 100; }
.modal-panel {
  position: fixed;
  top: 50%; left: 50%;
  transform: translate(-50%, -50%);
  width: 820px;
  max-height: 88vh;
  overflow-y: auto;
  background: var(--bg-white);
  border-radius: var(--radius-xl);
  box-shadow: var(--shadow-xl);
  z-index: 101;
  padding: 28px;
}
.modal-header { display: flex; align-items: center; justify-content: space-between; margin-bottom: 22px; }
.modal-title { font-size: 18px; font-weight: 700; letter-spacing: -0.02em; }
.modal-form { display: flex; flex-direction: column; gap: 18px; }

.form-row { display: grid; gap: 14px; }
.form-row.two { grid-template-columns: 1fr 1fr; }
.form-row.three { grid-template-columns: 1fr 1fr 1fr; }
.form-row.one { grid-template-columns: 1fr; }
.form-field { display: flex; flex-direction: column; gap: 5px; }
.form-field label { font-size: 12px; font-weight: 600; color: var(--text-secondary); }
.form-field input, .form-field select, .form-field textarea { width: 100%; }

.form-error {
  padding: 10px 14px; background: var(--danger-bg); border: 1px solid var(--danger-border);
  color: var(--danger); border-radius: var(--radius-sm); font-size: 13px;
}
.modal-actions { display: flex; justify-content: flex-end; gap: 10px; padding-top: 8px; border-top: 1px solid var(--border-light); }
</style>
