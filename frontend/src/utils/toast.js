// 简单的Toast通知实现
class Toast {
  constructor() {
    this.container = null
    this.createContainer()
  }

  createContainer() {
    if (this.container) return

    this.container = document.createElement('div')
    this.container.className = 'toast-container'
    this.container.style.cssText = `
      position: fixed;
      top: 20px;
      right: 20px;
      z-index: 10000;
      max-width: 400px;
    `
    document.body.appendChild(this.container)
  }

  show(message, type = 'info', duration = 3000) {
    const toast = document.createElement('div')
    toast.className = `toast toast-${type}`

    const colors = {
      success: '#10b981',
      error: '#ef4444',
      warning: '#f59e0b',
      info: '#3b82f6'
    }

    const icons = {
      success: '✓',
      error: '✕',
      warning: '⚠',
      info: 'ℹ'
    }

    toast.style.cssText = `
      background: white;
      border: 1px solid #e5e7eb;
      border-left: 4px solid ${colors[type]};
      border-radius: 8px;
      padding: 16px;
      margin-bottom: 8px;
      box-shadow: 0 10px 15px -3px rgba(0, 0, 0, 0.1);
      transform: translateX(100%);
      transition: transform 0.3s ease;
      display: flex;
      align-items: center;
      gap: 12px;
    `

    toast.innerHTML = `
      <span style="color: ${colors[type]}; font-weight: bold; font-size: 16px;">
        ${icons[type]}
      </span>
      <span style="color: #374151; flex: 1;">${message}</span>
    `

    this.container.appendChild(toast)

    // 显示动画
    setTimeout(() => {
      toast.style.transform = 'translateX(0)'
    }, 10)

    // 自动移除
    setTimeout(() => {
      toast.style.transform = 'translateX(100%)'
      setTimeout(() => {
        if (toast.parentNode) {
          toast.parentNode.removeChild(toast)
        }
      }, 300)
    }, duration)

    // 点击移除
    toast.addEventListener('click', () => {
      toast.style.transform = 'translateX(100%)'
      setTimeout(() => {
        if (toast.parentNode) {
          toast.parentNode.removeChild(toast)
        }
      }, 300)
    })

    return toast
  }

  success(message, duration) {
    return this.show(message, 'success', duration)
  }

  error(message, duration) {
    return this.show(message, 'error', duration)
  }

  warning(message, duration) {
    return this.show(message, 'warning', duration)
  }

  info(message, duration) {
    return this.show(message, 'info', duration)
  }
}

const toast = new Toast()

export default toast

// Vue插件
export const ToastPlugin = {
  install(app) {
    app.config.globalProperties.$toast = toast
    app.provide('toast', toast)
  }
}