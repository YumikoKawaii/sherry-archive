const KEY = 'sa_device_id'

function generateId(): string {
  if (typeof crypto !== 'undefined' && crypto.randomUUID) {
    return crypto.randomUUID()
  }
  // fallback for non-secure contexts (HTTP)
  return 'xxxxxxxx-xxxx-4xxx-yxxx-xxxxxxxxxxxx'.replace(/[xy]/g, c => {
    const r = Math.random() * 16 | 0
    return (c === 'x' ? r : (r & 0x3 | 0x8)).toString(16)
  })
}

export function getDeviceId(): string {
  let id = localStorage.getItem(KEY)
  if (!id) {
    id = generateId()
    localStorage.setItem(KEY, id)
  }
  return id
}
