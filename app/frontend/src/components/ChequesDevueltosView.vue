<script lang="ts" setup>
import { ref, shallowRef, onMounted, computed, inject, watch } from 'vue'
import { GetChequesDevueltos, CrearChequeDevuelto, UpdateChequeDevuelto, SoftDeleteChequeDevuelto, ResolveOrCreateCliente } from '../../wailsjs/go/main/App'
import { refreshBusKey } from '../sync/refreshBus'

const items = shallowRef<any[]>([])
const loading = ref(true)
const hasLoadedOnce = ref(false)
const searchQuery = ref('')
const filterEstado = ref('')
const showModal = ref(false)
const editId = ref<number | null>(null)
const saving = ref(false)
const formError = ref('')
const refreshBus = inject(refreshBusKey, null)

const filtered = computed(() => {
  let result = items.value
  if (searchQuery.value) {
    const q = searchQuery.value.toLowerCase()
    result = result.filter(i =>
      i.nombre_cliente?.toLowerCase().includes(q) ||
      i.rut_cliente?.toLowerCase().includes(q) ||
      i.numero_cheque?.toLowerCase().includes(q) ||
      i.motivo?.toLowerCase().includes(q)
    )
  }
  if (filterEstado.value) {
    result = result.filter(i => i.estado_pago === filterEstado.value)
  }
  return result
})

const totalMonto = computed(() =>
  filtered.value.reduce((sum, i) => sum + (i.monto || 0), 0)
)

const chequesProtestados = computed(() => {
  const base = items.value || []
  return {
    monto: base.reduce((sum, i) => sum + (i.monto || 0), 0),
    cantidad: base.length,
  }
})

const chequesPendientes = computed(() => {
  const base = (items.value || []).filter(i => i.estado_pago === 'Pendiente')
  return {
    monto: base.reduce((sum, i) => sum + (i.monto || 0), 0),
    cantidad: base.length,
  }
})

const chequesPagados = computed(() => {
  const base = (items.value || []).filter(i => i.estado_pago === 'Pagado')
  return {
    monto: base.reduce((sum, i) => sum + (i.monto || 0), 0),
    cantidad: base.length,
  }
})

const form = ref({
  id_cheque: '',
  rut_cliente: '',
  nombre_cliente: '',
  numero_cheque: '',
  monto: '',
  plaza: '',
  fecha_registro: '',
  fecha_cheque: '',
  tipo_pago: 'Pendiente',
  motivo: '',
  fecha_saldada: '',
  comentario: '',
})

async function loadData(silent = false) {
  const showLoader = !silent && !hasLoadedOnce.value
  if (showLoader) {
    loading.value = true
  }

  try {
    const nextItems = await GetChequesDevueltos() || []
    if (JSON.stringify(nextItems) !== JSON.stringify(items.value)) {
      items.value = nextItems
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

function formatMoney(amount: number) {
  return new Intl.NumberFormat('es-CL', { style: 'currency', currency: 'CLP', minimumFractionDigits: 0 }).format(amount)
}

function normalizeRutInput(value: string) {
  return value.replace(/\./g, '').trim().toUpperCase()
}

function hasValidRutFormat(value: string) {
  const v = value.trim().toUpperCase().replace(/\./g, '')
  return /^([0-9]{7,8}-[0-9K]|RUT-[0-9]{1,8})$/.test(v)
}

function resetForm() {
  editId.value = null
  form.value = {
    id_cheque: '',
    rut_cliente: '',
    nombre_cliente: '',
    numero_cheque: '',
    monto: '',
    plaza: '',
    fecha_registro: '',
    fecha_cheque: '',
    tipo_pago: 'Pendiente',
    motivo: '',
    fecha_saldada: '',
    comentario: '',
  }
  formError.value = ''
}

function openCreateModal() {
  resetForm()
  showModal.value = true
}

function openEditModal(item: any) {
  editId.value = item.id
  form.value = {
    id_cheque: item.id_cheque ? String(item.id_cheque) : '',
    rut_cliente: item.rut_cliente || '',
    nombre_cliente: item.nombre_cliente || '',
    numero_cheque: item.numero_cheque || '',
    monto: String(item.monto || ''),
    plaza: item.plaza || '',
    fecha_registro: item.fecha_registro || '',
    fecha_cheque: item.fecha_cheque || '',
    tipo_pago: item.tipo_pago || 'Pendiente',
    motivo: item.motivo || '',
    fecha_saldada: item.fecha_saldada || '',
    comentario: item.comentario || '',
  }
  formError.value = ''
  showModal.value = true
}

function closeModal() {
  showModal.value = false
}

function validatePagoRules() {
  const tipoPago = (form.value.tipo_pago || '').trim().toLowerCase()
  const fecha = (form.value.fecha_saldada || '').trim()
  if (tipoPago !== 'pendiente' && tipoPago !== '' && !fecha) {
    return 'Si seleccionas tipo de pago debes informar fecha saldada'
  }
  if ((tipoPago === 'pendiente' || tipoPago === '') && fecha) {
    return 'Si informas fecha saldada debes seleccionar tipo de pago'
  }
  return ''
}

async function handleSubmit() {
  formError.value = ''

  form.value.rut_cliente = normalizeRutInput(form.value.rut_cliente)

  if (!form.value.rut_cliente || !form.value.nombre_cliente || !form.value.monto) {
    formError.value = 'RUT, Nombre cliente y Monto son obligatorios'
    return
  }
  if (!hasValidRutFormat(form.value.rut_cliente)) {
    formError.value = 'Formato de RUT invalido. Usa 21189898-4 o RUT-0001'
    return
  }

  const pagoError = validatePagoRules()
  if (pagoError) {
    formError.value = pagoError
    return
  }

  const monto = parseFloat(form.value.monto)
  if (Number.isNaN(monto) || monto <= 0) {
    formError.value = 'El monto debe ser mayor a 0'
    return
  }

  try {
    await ResolveOrCreateCliente(form.value.rut_cliente, form.value.nombre_cliente)
  } catch (e: any) {
    formError.value = e?.message || 'No se pudo validar cliente'
    return
  }

  saving.value = true
  try {
    const payload = {
      id_cheque: form.value.id_cheque ? Number(form.value.id_cheque) : undefined,
      rut_cliente: form.value.rut_cliente,
      nombre_cliente: form.value.nombre_cliente,
      numero_cheque: form.value.numero_cheque,
      numero_factura: '',
      monto,
      plaza: form.value.plaza,
      fecha_registro: form.value.fecha_registro,
      fecha_cheque: form.value.fecha_cheque,
      tipo_pago: form.value.tipo_pago,
      motivo: form.value.motivo,
      fecha_saldada: form.value.fecha_saldada,
      comentario: form.value.comentario,
    }

    if (editId.value) {
      await UpdateChequeDevuelto(editId.value, payload as any)
    } else {
      await CrearChequeDevuelto(payload as any)
    }

    closeModal()
    await loadData(true)
  } catch (e: any) {
    console.error('Error guardando cheque devuelto:', e)
    formError.value = e?.message || 'Error al guardar'
  }
  saving.value = false
}

async function deleteItem(id: number) {
  const confirmed = window.confirm('Eliminar cheque devuelto? Esta accion lo ocultara del listado.')
  if (!confirmed) return

  await SoftDeleteChequeDevuelto(id)
  await loadData(true)
}

onMounted(loadData)

if (refreshBus) {
  watch(refreshBus.refreshTick, () => {
    loadData(true)
  })
}
</script>

<template>
  <div class="devueltos-view">
    <header class="view-header">
      <div class="view-header-left">
        <h2 class="view-title">Cheques Devueltos</h2>
        <div class="view-meta">
          <span class="meta-count">{{ filtered.length }} registros</span>
          <span class="meta-divider"></span>
          <span class="meta-total">{{ formatMoney(totalMonto) }}</span>
        </div>
      </div>
      <div class="view-header-right">
        <div class="search-wrapper">
          <svg class="search-icon" width="15" height="15" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><circle cx="11" cy="11" r="8"/><path d="m21 21-4.35-4.35"/></svg>
          <input v-model="searchQuery" type="text" placeholder="Buscar..." class="search-input" />
        </div>
        <select v-model="filterEstado" class="filter-select">
          <option value="">Todos</option>
          <option value="Pendiente">Pendiente</option>
          <option value="Pagado">Pagado</option>
        </select>
        <button class="btn btn-primary" @click="openCreateModal">
          Nuevo Devuelto
        </button>
      </div>
    </header>

    <div class="resumen-grid">
      <article class="resumen-card protestados">
        <span class="resumen-label">Cheques protestados</span>
        <strong class="resumen-monto">{{ formatMoney(chequesProtestados.monto) }}</strong>
        <span class="resumen-meta">{{ chequesProtestados.cantidad }} cheques protestados</span>
      </article>
      <article class="resumen-card pendientes">
        <span class="resumen-label">Cheques pendientes</span>
        <strong class="resumen-monto">{{ formatMoney(chequesPendientes.monto) }}</strong>
        <span class="resumen-meta">{{ chequesPendientes.cantidad }} cheques pendientes</span>
      </article>
      <article class="resumen-card pagados">
        <span class="resumen-label">Cheques pagados</span>
        <strong class="resumen-monto">{{ formatMoney(chequesPagados.monto) }}</strong>
        <span class="resumen-meta">{{ chequesPagados.cantidad }} cheques pagados</span>
      </article>
    </div>

    <div class="table-container card">
      <div v-if="loading" class="loading-state">
        <div class="spinner"></div>
        <span>Cargando datos...</span>
      </div>
      <table v-else>
        <thead>
          <tr>
            <th>Estado</th>
            <th>Cliente</th>
            <th>N Cheque</th>
            <th>Monto</th>
            <th>PLAZA</th>
            <th>Fecha registro</th>
            <th>Fecha cheque</th>
            <th>Tipo pago</th>
            <th>Motivo</th>
            <th>Fecha saldada</th>
            <th>Comentario</th>
            <th></th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="i in filtered" :key="i.id">
            <td>
              <span :class="['badge', i.estado_pago === 'Pagado' ? 'badge-success' : 'badge-warning']">{{ i.estado_pago }}</span>
            </td>
            <td>
              <div class="cliente-cell">
                <span class="cell-bold">{{ i.nombre_cliente }}</span>
                <span class="cliente-rut-label">{{ i.rut_cliente }}</span>
              </div>
            </td>
            <td class="cell-mono">{{ i.numero_cheque || '-' }}</td>
            <td class="cell-bold">{{ formatMoney(i.monto || 0) }}</td>
            <td>{{ i.plaza || '-' }}</td>
            <td>{{ i.fecha_registro || '-' }}</td>
            <td>{{ i.fecha_cheque || '-' }}</td>
            <td>{{ i.tipo_pago || 'Pendiente' }}</td>
            <td class="cell-truncate" :title="i.motivo || '-'">{{ i.motivo || '-' }}</td>
            <td>{{ i.fecha_saldada || '-' }}</td>
            <td class="cell-truncate" :title="i.comentario || '-'">{{ i.comentario || '-' }}</td>
            <td class="cell-actions">
              <button class="btn-ghost-icon" @click="openEditModal(i)">Editar</button>
              <button class="btn-danger-ghost" @click="deleteItem(i.id)">X</button>
            </td>
          </tr>
          <tr v-if="filtered.length === 0 && !loading">
            <td colspan="12" class="empty-row">No se encontraron registros</td>
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
          <h3 class="modal-title">{{ editId ? 'Editar' : 'Nuevo' }} Cheque Devuelto</h3>
          <button class="btn-icon" @click="closeModal">X</button>
        </div>

        <form @submit.prevent="handleSubmit" class="modal-form">
          <div class="form-row">
            <div class="form-field">
              <label>RUT Cliente</label>
              <input v-model="form.rut_cliente" placeholder="21189898-4" autocomplete="off" />
            </div>
            <div class="form-field">
              <label>Nombre cliente</label>
              <input v-model="form.nombre_cliente" placeholder="Nombre del cliente" autocomplete="off" />
            </div>
          </div>

          <div class="form-row three">
            <div class="form-field">
              <label>N Cheque</label>
              <input v-model="form.numero_cheque" />
            </div>
            <div class="form-field">
              <label>Monto</label>
              <input v-model="form.monto" type="number" step="1" min="0" />
            </div>
            <div class="form-field">
              <label>PLAZA</label>
              <input v-model="form.plaza" placeholder="Comuna" />
            </div>
            <div class="form-field">
              <label>Motivo</label>
              <input v-model="form.motivo" />
            </div>
          </div>

          <div class="form-row three">
            <div class="form-field">
              <label>Fecha registro</label>
              <input v-model="form.fecha_registro" type="date" />
            </div>
            <div class="form-field">
              <label>Fecha cheque</label>
              <input v-model="form.fecha_cheque" type="date" />
            </div>
            <div class="form-field">
              <label>Fecha saldada</label>
              <input v-model="form.fecha_saldada" type="date" />
            </div>
          </div>

          <div class="form-row two">
            <div class="form-field">
              <label>Tipo de pago</label>
              <select v-model="form.tipo_pago">
                <option value="Pendiente">Pendiente</option>
                <option value="Transferencia">Transferencia</option>
                <option value="Efectivo">Efectivo</option>
                <option value="Cheque">Cheque</option>
              </select>
            </div>
            <div class="form-field">
              <label>Comentario</label>
              <textarea v-model="form.comentario" rows="2"></textarea>
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
.devueltos-view { display: flex; flex-direction: column; height: 100%; padding: 24px 28px; gap: 16px; }
.view-header { display: flex; align-items: center; justify-content: space-between; }
.view-header-left { display: flex; align-items: baseline; gap: 14px; }
.view-title { font-size: 22px; font-weight: 700; letter-spacing: -0.03em; }
.view-meta { display: flex; align-items: center; gap: 10px; }
.meta-count { font-size: 12px; color: var(--text-muted); background: var(--bg-surface); padding: 3px 10px; border-radius: 12px; font-weight: 500; }
.meta-divider { width: 3px; height: 3px; background: var(--border); border-radius: 50%; }
.meta-total { font-size: 13px; color: var(--primary-dark); font-weight: 600; font-family: var(--font-mono); }
.view-header-right { display: flex; align-items: center; gap: 8px; }

.search-wrapper { position: relative; }
.search-icon { position: absolute; left: 11px; top: 50%; transform: translateY(-50%); color: var(--text-muted); pointer-events: none; }
.search-input { padding-left: 32px; width: 210px; }
.filter-select { min-width: 140px; }

.resumen-grid {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 12px;
}

.resumen-card {
  background: var(--bg-white);
  border: 1px solid var(--border-light);
  border-radius: var(--radius-md);
  padding: 14px 16px;
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.resumen-card.protestados {
  border-color: var(--danger-border);
  background: var(--danger-bg);
}

.resumen-card.pendientes {
  border-color: var(--warning-border);
  background: var(--warning-bg);
}

.resumen-card.pagados {
  border-color: var(--success-border);
  background: var(--success-bg);
}

.resumen-label {
  font-size: 12px;
  color: var(--text-muted);
  text-transform: uppercase;
  letter-spacing: 0.06em;
  font-weight: 600;
}

.resumen-monto {
  font-size: 22px;
  letter-spacing: -0.03em;
  color: var(--text-primary);
}

.resumen-meta {
  font-size: 12px;
  color: var(--text-secondary);
}

.table-container { flex: 1; overflow: auto; }
.cliente-cell { display: flex; flex-direction: column; gap: 3px; }
.cliente-rut-label {
  font-family: var(--font-mono);
  font-size: 11px;
  color: var(--primary-dark);
  background: var(--primary-light);
  border-radius: 10px;
  padding: 2px 7px;
  width: fit-content;
}
.cell-mono { font-family: var(--font-mono); font-size: 12px; }
.cell-bold { font-weight: 600; }
.cell-truncate { max-width: 180px; white-space: nowrap; overflow: hidden; text-overflow: ellipsis; }
.cell-actions { display: flex; align-items: center; gap: 6px; }

:deep(td.cell-actions) {
  border-bottom: 1px solid var(--border-light) !important;
  background-clip: padding-box;
}

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
  width: 760px;
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

.form-row { display: grid; grid-template-columns: 1fr 1fr; gap: 14px; }
.form-row.three { grid-template-columns: 1fr 1fr 1fr; }
.form-row.two { grid-template-columns: 1fr 1fr; }
.form-field { display: flex; flex-direction: column; gap: 5px; }
.form-field label { font-size: 12px; font-weight: 600; color: var(--text-secondary); }
.form-field input, .form-field select, .form-field textarea { width: 100%; }

.form-error {
  padding: 10px 14px; background: var(--danger-bg); border: 1px solid var(--danger-border);
  color: var(--danger); border-radius: var(--radius-sm); font-size: 13px;
}
.modal-actions { display: flex; justify-content: flex-end; gap: 10px; padding-top: 8px; border-top: 1px solid var(--border-light); }
</style>
