<script lang="ts" setup>
import { ref, shallowRef, onMounted, computed, inject, watch } from 'vue'
import { GetCheques, GetEstados, GetClientes, UpdateEstadoCheque, SoftDeleteCheque, ExportChequesExcel, CrearCheque, UpdateCheque, ResolveOrCreateCliente } from '../../wailsjs/go/main/App'
import { refreshBusKey } from '../sync/refreshBus'

const cheques = shallowRef<any[]>([])
const estados = shallowRef<any[]>([])
const clientes = shallowRef<any[]>([])
const searchQuery = ref('')
const filterEstado = ref('')
const loading = ref(true)
const hasLoadedOnce = ref(false)
const showModal = ref(false)
const showDetailModal = ref(false)
const detailCheque = ref<any>(null)
const editChequeId = ref<number | null>(null)
const refreshBus = inject(refreshBusKey, null)

const filtered = computed(() => {
  let result = cheques.value
  if (searchQuery.value) {
    const q = searchQuery.value.toLowerCase()
    result = result.filter(c =>
      c.nombre_cliente?.toLowerCase().includes(q) ||
      c.rut_cliente?.toLowerCase().includes(q) ||
      c.numero_cheque?.toLowerCase().includes(q) ||
      c.numero_factura?.toLowerCase().includes(q) ||
      c.banco_cheque?.toLowerCase().includes(q)
    )
  }
  if (filterEstado.value) {
    result = result.filter(c => String(c.id_estado) === filterEstado.value)
  }
  return result
})

const totalMonto = computed(() =>
  filtered.value.reduce((sum, c) => sum + (c.monto || 0), 0)
)

const totalCartera = computed(() => {
  const base = cheques.value || []
  return {
    monto: base.reduce((sum, c) => sum + (c.monto || 0), 0),
    cantidad: base.length,
  }
})

const chequesPorDepositar = computed(() => {
  const base = (cheques.value || []).filter(c =>
    c.estado_nombre === 'En cartera' || c.id_estado === 1
  )
  return {
    monto: base.reduce((sum, c) => sum + (c.monto || 0), 0),
    cantidad: base.length,
  }
})

async function loadData(silent = false) {
  const showLoader = !silent && !hasLoadedOnce.value
  if (showLoader) {
    loading.value = true
  }

  try {
    const [ch, es, cl] = await Promise.all([GetCheques(), GetEstados(), GetClientes()])

    const nextCheques = ch || []
    const nextEstados = es || []
    const nextClientes = cl || []

    if (JSON.stringify(nextCheques) !== JSON.stringify(cheques.value)) {
      cheques.value = nextCheques
    }
    if (JSON.stringify(nextEstados) !== JSON.stringify(estados.value)) {
      estados.value = nextEstados
    }
    if (JSON.stringify(nextClientes) !== JSON.stringify(clientes.value)) {
      clientes.value = nextClientes
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

async function changeEstado(chequeId: number, nuevoEstadoId: number) {
  const cheque = (cheques.value || []).find(c => c.id === chequeId)
  if (cheque?.id_estado === 3 || cheque?.estado_nombre === 'Protestado/Devuelto') {
    formError.value = 'El cheque protestado/devuelto solo puede gestionarse en la seccion Cheques devueltos'
    return
  }

  await UpdateEstadoCheque(chequeId, nuevoEstadoId)
  await loadData(true)
}

async function deleteCheque(chequeId: number) {
  await SoftDeleteCheque(chequeId)
  await loadData(true)
}

function formatMoney(amount: number) {
  return new Intl.NumberFormat('es-CL', { style: 'currency', currency: 'CLP', minimumFractionDigits: 0 }).format(amount)
}

function estadoBadge(nombre: string | null) {
  if (!nombre) return 'badge-info'
  if (nombre === 'Depositado') return 'badge-success'
  if (nombre === 'Protestado/Devuelto') return 'badge-danger'
  return 'badge-info'
}

async function exportToExcel() {
  exporting.value = true
  try {
    const result = await ExportChequesExcel()
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

const form = ref({
  rut_cliente: '',
  nombre_cliente: '',
  numero_cheque: '',
  monto: '',
  fecha_recepcion: '',
  fecha_deposito: '',
  fecha_cheque_cobrar: '',
  banco_cheque: '',
  numero_factura: '',
  condiciones_pago: '',
  observaciones: '',
  id_estado: 1,
})
const clienteInfo = ref<any>(null)
const rutSuggestions = shallowRef<any[]>([])
const showSuggestions = ref(false)
const saving = ref(false)
const exporting = ref(false)
const formError = ref('')
const formSuccess = ref(false)

function onRutInput() {
  const q = form.value.rut_cliente.toLowerCase()
  if (q.length < 2) {
    rutSuggestions.value = []
    showSuggestions.value = false
    return
  }
  rutSuggestions.value = clientes.value.filter(c =>
    c.rut.toLowerCase().includes(q) || c.razon_social.toLowerCase().includes(q)
  ).slice(0, 6)
  showSuggestions.value = rutSuggestions.value.length > 0
}

function selectCliente(c: any) {
  form.value.rut_cliente = c.rut
  form.value.nombre_cliente = c.razon_social || ''
  clienteInfo.value = c
  showSuggestions.value = false
}

function normalizeRutInput(value: string) {
  const cleaned = value.replace(/\./g, '').trim().toUpperCase()
  const parts = cleaned.split('-')
  if (parts.length === 2) {
    return `${parts[0]}-${parts[1]}`
  }
  return cleaned
}

function hideSuggestions() {
  setTimeout(() => {
    showSuggestions.value = false
  }, 200)
}

function hasValidRutFormat(value: string) {
  const v = value.trim().toUpperCase().replace(/\./g, '')
  return /^([0-9]{7,8}-[0-9K]|RUT-[0-9]{1,8})$/.test(v)
}

function openModal() {
  resetForm()
  editChequeId.value = null
  showModal.value = true
}

function openEditModal(cheque: any) {
  if (cheque?.id_estado === 3 || cheque?.estado_nombre === 'Protestado/Devuelto') {
    formError.value = 'El cheque protestado/devuelto solo puede gestionarse en la seccion Cheques devueltos'
    return
  }

  editChequeId.value = cheque.id
  form.value = {
    rut_cliente: cheque.rut_cliente || '',
    nombre_cliente: cheque.nombre_cliente || '',
    numero_cheque: cheque.numero_cheque || '',
    monto: String(cheque.monto || ''),
    fecha_recepcion: cheque.fecha_recepcion || '',
    fecha_deposito: cheque.fecha_deposito || '',
    fecha_cheque_cobrar: cheque.fecha_cheque_cobrar || '',
    banco_cheque: cheque.banco_cheque || '',
    numero_factura: cheque.numero_factura || '',
    condiciones_pago: cheque.condiciones_pago || '',
    observaciones: cheque.observaciones || '',
    id_estado: cheque.id_estado || 1,
  }
  clienteInfo.value = { rut: cheque.rut_cliente, razon_social: cheque.nombre_cliente }
  formError.value = ''
  formSuccess.value = false
  showModal.value = true
}

function closeModal() {
  showModal.value = false
  editChequeId.value = null
}

function openDetailModal(cheque: any) {
  detailCheque.value = cheque
  showDetailModal.value = true
}

function closeDetailModal() {
  showDetailModal.value = false
  detailCheque.value = null
}

function resetForm() {
  form.value = {
    rut_cliente: '',
    nombre_cliente: '',
    numero_cheque: '',
    monto: '',
    fecha_recepcion: '',
    fecha_deposito: '',
    fecha_cheque_cobrar: '',
    banco_cheque: '',
    numero_factura: '',
    condiciones_pago: '',
    observaciones: '',
    id_estado: 1,
  }
  clienteInfo.value = null
  formError.value = ''
  formSuccess.value = false
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
      rut_cliente: form.value.rut_cliente,
      nombre_cliente: form.value.nombre_cliente,
      numero_cheque: form.value.numero_cheque,
      monto,
      fecha_recepcion: form.value.fecha_recepcion,
      fecha_deposito: form.value.fecha_deposito,
      fecha_cheque_cobrar: form.value.fecha_cheque_cobrar,
      banco_cheque: form.value.banco_cheque,
      numero_factura: form.value.numero_factura,
      condiciones_pago: form.value.condiciones_pago,
      observaciones: form.value.observaciones,
      id_estado: form.value.id_estado,
    }

    if (editChequeId.value) {
      await UpdateCheque(editChequeId.value, payload as any)
    } else {
      await CrearCheque(payload as any)
    }

    formSuccess.value = true
    await loadData(true)
    setTimeout(() => {
      closeModal()
    }, 800)
  } catch (e: any) {
    console.error('Error guardando cheque:', e)
    const msg = String(e?.message || '')
    if (msg.toLowerCase().includes('foreign key constraint') && msg.toLowerCase().includes('rut_cliente')) {
      formError.value = 'El RUT ingresado no existe en la base de clientes. Selecciona un cliente sugerido o crealo primero.'
    } else if (msg.toLowerCase().includes('no existe')) {
      formError.value = msg
    } else if (msg.toLowerCase().includes('rut invalido') || msg.toLowerCase().includes('formato de rut invalido') || msg.toLowerCase().includes('digito verificador invalido')) {
      formError.value = msg
    } else {
      formError.value = e?.message || 'Error al guardar'
    }
  }
  saving.value = false
}

onMounted(loadData)

if (refreshBus) {
  watch(refreshBus.refreshTick, () => {
    loadData(true)
  })
}
</script>

<template>
  <div class="cheques-view">
    <div class="search-filter-row">
      <div class="search-wrapper">
        <svg class="search-icon" width="15" height="15" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><circle cx="11" cy="11" r="8"/><path d="m21 21-4.35-4.35"/></svg>
        <input v-model="searchQuery" type="text" placeholder="Buscar..." class="search-input" />
      </div>
      <select v-model="filterEstado" class="filter-select">
        <option value="">Todos los estados</option>
        <option v-for="e in estados" :key="e.id" :value="String(e.id)">{{ e.nombre }}</option>
      </select>
    </div>
    <header class="view-header">
      <div class="view-header-left">
        <h2 class="view-title">Cheques</h2>
        <div class="view-meta">
          <span class="meta-count">{{ filtered.length }} registros</span>
          <span class="meta-divider"></span>
          <span class="meta-total">{{ formatMoney(totalMonto) }}</span>
        </div>
      </div>
      <div class="view-header-right">
        <button class="btn btn-outline" @click="exportToExcel" :disabled="exporting">
          <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4"/><polyline points="7 10 12 15 17 10"/><line x1="12" y1="15" x2="12" y2="3"/></svg>
          {{ exporting ? 'Exportando...' : 'Exportar Excel' }}
        </button>
        <button class="btn btn-primary" @click="openModal">
          <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5"><line x1="12" y1="5" x2="12" y2="19"/><line x1="5" y1="12" x2="19" y2="12"/></svg>
          Nuevo Cheque
        </button>
      </div>
    </header>

    <div class="resumen-grid">
      <article class="resumen-card">
        <div class="resumen-content">
          <span class="resumen-label">Total cartera</span>
          <strong class="resumen-monto">{{ formatMoney(totalCartera.monto) }}</strong>
          <span class="resumen-meta">{{ totalCartera.cantidad }} cheques</span>
        </div>
        <div class="resumen-icon icon-info">
          <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><rect x="2" y="7" width="20" height="14" rx="2" ry="2"/><path d="M16 21V5a2 2 0 0 0-2-2h-4a2 2 0 0 0-2 2v16"/></svg>
        </div>
      </article>
      <article class="resumen-card">
        <div class="resumen-content">
          <span class="resumen-label">Cheques por depositar</span>
          <strong class="resumen-monto">{{ formatMoney(chequesPorDepositar.monto) }}</strong>
          <span class="resumen-meta">{{ chequesPorDepositar.cantidad }} cheques</span>
        </div>
        <div class="resumen-icon icon-warning">
          <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><circle cx="12" cy="12" r="10"/><polyline points="12 6 12 12 16 14"/></svg>
        </div>
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
            <th>N Cheque</th>
            <th>Cliente</th>
            <th>Monto</th>
            <th>F. Recepcion</th>
            <th>F. Deposito</th>
            <th>F. Cheque Cobrar</th>
            <th>Banco</th>
            <th>N Factura</th>
            <th>Condiciones Pago</th>
            <th>Estado</th>
            <th>Vendedor</th>
            <th></th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="c in filtered" :key="c.id">
            <td class="cell-mono">{{ c.numero_cheque || '-' }}</td>
            <td>
              <div class="cliente-cell">
                <span class="cell-bold">{{ c.nombre_cliente }}</span>
                <span class="cliente-rut-label">{{ c.rut_cliente }}</span>
              </div>
            </td>
            <td class="cell-bold">{{ formatMoney(c.monto) }}</td>
            <td class="cell-muted">{{ c.fecha_recepcion || '-' }}</td>
            <td class="cell-muted">{{ c.fecha_deposito || '-' }}</td>
            <td class="cell-muted">{{ c.fecha_cheque_cobrar || '-' }}</td>
            <td class="cell-muted">{{ c.banco_cheque || '-' }}</td>
            <td class="cell-mono cell-muted">{{ c.numero_factura || '-' }}</td>
            <td class="cell-muted cell-truncate" :title="c.condiciones_pago || '-'">{{ c.condiciones_pago || '-' }}</td>
            <td>
              <select
                :value="c.id_estado"
                @change="changeEstado(c.id, Number(($event.target as HTMLSelectElement).value))"
                :class="['estado-pill', estadoBadge(c.estado_nombre)]"
                :disabled="c.id_estado === 3 || c.estado_nombre === 'Protestado/Devuelto'"
              >
                <option v-for="e in estados" :key="e.id" :value="e.id">{{ e.nombre }}</option>
              </select>
            </td>
            <td class="cell-muted">{{ c.nombre_vendedor || '-' }}</td>
            <td class="cell-actions">
              <button
                class="btn-ghost-icon"
                title="Editar"
                @click="openEditModal(c)"
                :disabled="c.id_estado === 3 || c.estado_nombre === 'Protestado/Devuelto'"
              >
                <svg width="15" height="15" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M12 20h9"/><path d="M16.5 3.5a2.1 2.1 0 0 1 3 3L7 19l-4 1 1-4Z"/></svg>
              </button>
              <button class="btn-ghost-icon" title="Ver detalles" @click="openDetailModal(c)">
                <svg width="15" height="15" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M1 12s4-8 11-8 11 8 11 8-4 8-11 8S1 12 1 12z"/><circle cx="12" cy="12" r="3"/></svg>
              </button>
              <button class="btn-danger-ghost" title="Eliminar" @click="deleteCheque(c.id)">
                <svg width="15" height="15" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><polyline points="3 6 5 6 21 6"/><path d="M19 6v14a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2V6"/><path d="M8 6V4a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2"/></svg>
              </button>
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
          <h3 class="modal-title">{{ editChequeId ? 'Editar Cheque' : 'Registrar Nuevo Cheque' }}</h3>
          <button class="btn-icon" @click="closeModal">
            <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><line x1="18" y1="6" x2="6" y2="18"/><line x1="6" y1="6" x2="18" y2="18"/></svg>
          </button>
        </div>

        <Transition name="fade">
          <div v-if="formSuccess" class="modal-success">
            <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="var(--success)" stroke-width="2.5"><path d="M20 6 9 17l-5-5"/></svg>
            Registro guardado exitosamente
          </div>
        </Transition>

        <form v-if="!formSuccess" @submit.prevent="handleSubmit" class="modal-form">
          <div class="form-row">
            <div class="form-field autocomplete-wrap">
              <label>RUT Cliente</label>
              <input v-model="form.rut_cliente" @input="onRutInput" @focus="onRutInput" @blur="hideSuggestions" placeholder="Buscar RUT o nombre..." autocomplete="off" />
              <div v-if="showSuggestions" class="suggestions">
                <div v-for="s in rutSuggestions" :key="s.rut" class="sug-item" @mousedown="selectCliente(s)">
                  <span class="sug-rut">{{ s.rut }}</span>
                  <span class="sug-name">{{ s.razon_social }}</span>
                </div>
              </div>
            </div>
            <div class="form-field">
              <label>Nombre cliente</label>
              <input v-model="form.nombre_cliente" placeholder="Nombre del cliente" />
            </div>
          </div>

          <div class="form-row three">
            <div class="form-field">
              <label>N Cheque</label>
              <input v-model="form.numero_cheque" placeholder="Numero" />
            </div>
            <div class="form-field">
              <label>Banco</label>
              <input v-model="form.banco_cheque" placeholder="Banco emisor" />
            </div>
            <div class="form-field">
              <label>N Factura</label>
              <input v-model="form.numero_factura" placeholder="Factura" />
            </div>
          </div>

          <div class="form-row three">
            <div class="form-field">
              <label>Monto ($)</label>
              <input v-model="form.monto" type="number" step="1" min="0" placeholder="0" required />
            </div>
            <div class="form-field">
              <label>Estado</label>
              <select v-model="form.id_estado">
                <option v-for="e in estados" :key="e.id" :value="e.id">{{ e.nombre }}</option>
              </select>
            </div>
            <div class="form-field">
              <label>Condiciones de Pago</label>
              <input v-model="form.condiciones_pago" placeholder="Ej: 30 dias" />
            </div>
          </div>

          <div class="form-row three">
            <div class="form-field">
              <label>Fecha Recepcion</label>
              <input v-model="form.fecha_recepcion" type="date" />
            </div>
            <div class="form-field">
              <label>Fecha Deposito</label>
              <input v-model="form.fecha_deposito" type="date" />
            </div>
            <div class="form-field">
              <label>Fecha Cheque a Cobrar</label>
              <input v-model="form.fecha_cheque_cobrar" type="date" />
            </div>
          </div>

          <div class="form-row one">
            <div class="form-field">
              <label>Observaciones</label>
              <textarea v-model="form.observaciones" rows="3" placeholder="Detalle adicional del cheque"></textarea>
            </div>
          </div>

          <div v-if="formError" class="form-error">{{ formError }}</div>

          <div class="modal-actions">
            <button type="button" class="btn btn-ghost" @click="closeModal">Cancelar</button>
            <button type="submit" class="btn btn-primary" :disabled="saving">
              <svg v-if="!saving" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M19 21H5a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h11l5 5v11a2 2 0 0 1-2 2z"/><polyline points="17 21 17 13 7 13 7 21"/><polyline points="7 3 7 8 15 8"/></svg>
              {{ saving ? 'Guardando...' : (editChequeId ? 'Guardar cambios' : 'Guardar') }}
            </button>
          </div>
        </form>
      </div>
    </Transition>

    <Transition name="overlay">
      <div v-if="showDetailModal" class="modal-overlay" @click.self="closeDetailModal"></div>
    </Transition>
    <Transition name="modal">
      <div v-if="showDetailModal" class="modal-panel detail-panel">
        <div class="modal-header">
          <h3 class="modal-title">Detalle del Cheque</h3>
          <button class="btn-icon" @click="closeDetailModal">
            <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><line x1="18" y1="6" x2="6" y2="18"/><line x1="6" y1="6" x2="18" y2="18"/></svg>
          </button>
        </div>

        <div v-if="detailCheque" class="detail-grid">
          <div class="detail-item"><span>Cliente</span><strong>{{ detailCheque.nombre_cliente }}</strong></div>
          <div class="detail-item"><span>RUT</span><strong>{{ detailCheque.rut_cliente }}</strong></div>
          <div class="detail-item"><span>N Cheque</span><strong>{{ detailCheque.numero_cheque || '-' }}</strong></div>
          <div class="detail-item"><span>Banco</span><strong>{{ detailCheque.banco_cheque || '-' }}</strong></div>
          <div class="detail-item"><span>N Factura</span><strong>{{ detailCheque.numero_factura || '-' }}</strong></div>
          <div class="detail-item"><span>Condiciones Pago</span><strong>{{ detailCheque.condiciones_pago || '-' }}</strong></div>
          <div class="detail-item"><span>Fecha Recepcion</span><strong>{{ detailCheque.fecha_recepcion || '-' }}</strong></div>
          <div class="detail-item"><span>Fecha Deposito</span><strong>{{ detailCheque.fecha_deposito || '-' }}</strong></div>
          <div class="detail-item"><span>Fecha Cheque a Cobrar</span><strong>{{ detailCheque.fecha_cheque_cobrar || '-' }}</strong></div>
          <div class="detail-item"><span>Estado</span><strong>{{ detailCheque.estado_nombre || '-' }}</strong></div>
          <div class="detail-item"><span>Vendedor</span><strong>{{ detailCheque.nombre_vendedor || '-' }}</strong></div>
          <div class="detail-item"><span>Monto</span><strong>{{ formatMoney(detailCheque.monto || 0) }}</strong></div>
        </div>

        <div class="detail-observaciones">
          <span>Observaciones</span>
          <p>{{ detailCheque?.observaciones || '-' }}</p>
        </div>
      </div>
    </Transition>
  </div>
</template>

<style scoped>
.cheques-view { display: flex; flex-direction: column; height: 100%; padding: 24px 28px; gap: 16px; }

.view-header { display: flex; align-items: center; justify-content: space-between; flex-shrink: 0; }
.view-header-left { display: flex; align-items: baseline; gap: 14px; }
.view-title { font-size: 22px; font-weight: 700; letter-spacing: -0.03em; }
.view-meta { display: flex; align-items: center; gap: 10px; }
.meta-count { font-size: 12px; color: var(--text-muted); background: var(--bg-surface); padding: 3px 10px; border-radius: 12px; font-weight: 500; }
.meta-divider { width: 3px; height: 3px; background: var(--border); border-radius: 50%; }
.meta-total { font-size: 13px; color: var(--primary-dark); font-weight: 600; font-family: var(--font-mono); }
.view-header-right { display: flex; align-items: center; gap: 8px; }

.search-filter-row { display: flex; gap: 16px; width: 100%; }
.search-wrapper { position: relative; flex: 1; }
.search-icon { position: absolute; left: 11px; top: 50%; transform: translateY(-50%); color: var(--text-muted); pointer-events: none; }
.search-input { padding-left: 32px; width: 100%; }
.filter-select { flex: 0 0 250px; }

.resumen-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
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
.icon-info { background: var(--info-bg); color: var(--info-text); }
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
.cell-muted { color: var(--text-secondary); }
.cell-truncate {
  max-width: 170px;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.cell-actions { display: flex; align-items: center; gap: 6px; }
.btn-ghost-icon {
  width: 30px;
  height: 30px;
  border-radius: 8px;
  border: 1px solid var(--border-light);
  color: var(--text-muted);
  background: transparent;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: all var(--transition);
}
.btn-ghost-icon:hover {
  color: var(--primary);
  border-color: var(--primary);
  background: var(--primary-light);
}
.btn-ghost-icon:disabled {
  opacity: 0.45;
  cursor: not-allowed;
  color: var(--text-muted);
  border-color: var(--border-light);
  background: transparent;
}

:deep(td.cell-actions) {
  border-bottom: 1px solid var(--border-light) !important;
  background-clip: padding-box;
}

.estado-pill {
  padding: 2px 8px 2px 8px;
  font-size: 11px;
  font-weight: 600;
  border-radius: 20px;
  border: none;
  cursor: pointer;
  appearance: none;
  background-repeat: no-repeat;
  background-position: right 6px center;
  background-image: url("data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' width='10' height='10' viewBox='0 0 24 24' fill='none' stroke='%230e1934' stroke-width='2'%3E%3Cpath d='M6 9l6 6 6-6'/%3E%3C/svg%3E");
  padding-right: 20px;
}
.estado-pill.badge-info { background: var(--info-bg); color: var(--info-text); }
.estado-pill.badge-success { background: var(--success-bg); color: var(--success); }
.estado-pill.badge-danger { background: var(--danger-bg); color: var(--danger); }

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
.detail-panel { width: 700px; }

.modal-header { display: flex; align-items: center; justify-content: space-between; margin-bottom: 22px; }
.modal-title { font-size: 18px; font-weight: 700; letter-spacing: -0.02em; }

.modal-success {
  display: flex; align-items: center; gap: 10px; padding: 18px;
  background: var(--success-bg); border: 1px solid var(--success-border); color: var(--success);
  border-radius: var(--radius-md); font-weight: 600; font-size: 14px;
}

.modal-form { display: flex; flex-direction: column; gap: 18px; }

.form-row { display: grid; grid-template-columns: 1fr 1fr; gap: 14px; }
.form-row.three { grid-template-columns: 1fr 1fr 1fr; }
.form-row.one { grid-template-columns: 1fr; }
.form-field { display: flex; flex-direction: column; gap: 5px; }
.form-field label { font-size: 12px; font-weight: 600; color: var(--text-secondary); }
.form-field input, .form-field select, .form-field textarea { width: 100%; }

.autocomplete-wrap { position: relative; }
.suggestions {
  position: absolute; top: 100%; left: 0; right: 0;
  background: var(--bg-white); border: 1px solid var(--border);
  border-radius: var(--radius-sm); box-shadow: var(--shadow-lg);
  z-index: 50; max-height: 180px; overflow-y: auto;
}
.sug-item { display: flex; align-items: center; gap: 10px; padding: 8px 12px; cursor: pointer; transition: background var(--transition); }
.sug-item:hover { background: var(--bg-surface); }
.sug-rut { font-family: var(--font-mono); font-size: 12px; color: var(--primary); min-width: 70px; }
.sug-name { font-size: 13px; color: var(--text-primary); }

.form-error {
  padding: 10px 14px; background: var(--danger-bg); border: 1px solid var(--danger-border);
  color: var(--danger); border-radius: var(--radius-sm); font-size: 13px;
}
.modal-actions { display: flex; justify-content: flex-end; gap: 10px; padding-top: 8px; border-top: 1px solid var(--border-light); }

.detail-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 10px 14px;
  margin-bottom: 14px;
}
.detail-item {
  display: flex;
  flex-direction: column;
  gap: 2px;
  padding: 10px 12px;
  border: 1px solid var(--border-light);
  border-radius: var(--radius-sm);
  background: var(--bg-surface);
}
.detail-item span { font-size: 11px; color: var(--text-muted); text-transform: uppercase; letter-spacing: 0.06em; }
.detail-item strong { font-size: 13px; color: var(--text-primary); }

.detail-observaciones {
  border: 1px solid var(--border-light);
  border-radius: var(--radius-sm);
  padding: 12px;
  background: var(--bg-surface);
}
.detail-observaciones span {
  display: block;
  font-size: 11px;
  color: var(--text-muted);
  text-transform: uppercase;
  letter-spacing: 0.06em;
  margin-bottom: 8px;
}
.detail-observaciones p {
  margin: 0;
  font-size: 13px;
  color: var(--text-primary);
  line-height: 1.45;
  white-space: pre-wrap;
}
</style>
