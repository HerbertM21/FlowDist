import type { InjectionKey, Ref } from 'vue'

export type RefreshBus = {
  refreshTick: Ref<number>
  notifyRefresh: () => void
}

export const refreshBusKey: InjectionKey<RefreshBus> = Symbol('refresh-bus')
