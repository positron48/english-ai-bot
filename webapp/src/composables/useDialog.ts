import { ref } from 'vue'

interface DialogState {
  message: string
  visible: boolean
}

// Global state shared across all components
const alertState = ref<DialogState>({
  message: '',
  visible: false
})

const confirmState = ref<DialogState>({
  message: '',
  visible: false
})

let alertResolve: (() => void) | null = null
let confirmResolve: ((value: boolean) => void) | null = null

// Global functions
export const showAlert = (message: string): Promise<void> => {
  return new Promise((resolve) => {
    alertState.value = {
      message,
      visible: true
    }
    alertResolve = resolve
  })
}

export const closeAlert = () => {
  alertState.value.visible = false
  if (alertResolve) {
    alertResolve()
    alertResolve = null
  }
}

export const showConfirm = (message: string): Promise<boolean> => {
  return new Promise((resolve) => {
    confirmState.value = {
      message,
      visible: true
    }
    confirmResolve = resolve
  })
}

export const closeConfirm = (confirmed: boolean) => {
  confirmState.value.visible = false
  if (confirmResolve) {
    confirmResolve(confirmed)
    confirmResolve = null
  }
}

export function useDialog() {
  return {
    alertState,
    confirmState,
    showAlert,
    closeAlert,
    showConfirm,
    closeConfirm
  }
}
