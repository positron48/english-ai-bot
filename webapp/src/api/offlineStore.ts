import { currentUserScope } from './sessionScope'

// Old unscoped databases deliberately remain unread: their owner cannot be established.
export function createOfflineStore<S extends string>(name: string, stores: Record<S, IDBObjectStoreParameters | undefined>) {
  const connections = new Map<string, Promise<IDBDatabase>>()
  function openDB(scope: string): Promise<IDBDatabase> {
    let pending = connections.get(scope)
    if (!pending) {
      pending = new Promise((resolve, reject) => {
        const request = indexedDB.open(`${name}:v3:${scope}`, 1)
        request.onupgradeneeded = () => {
          for (const [store, options] of Object.entries(stores)) {
            request.result.createObjectStore(store, options as IDBObjectStoreParameters | undefined)
          }
        }
        request.onsuccess = () => resolve(request.result)
        request.onerror = () => { connections.delete(scope); reject(request.error) }
      })
      connections.set(scope, pending)
    }
    return pending
  }
  return async function tx<T>(storeName: S, mode: IDBTransactionMode, operation: (store: IDBObjectStore) => IDBRequest<T> | void): Promise<T | undefined> {
    const scope = currentUserScope()
    if (scope === 'anon') {
      if (mode === 'readwrite') throw new Error('Authentication required for offline data')
      return undefined
    }
    const db = await openDB(scope)
    if (currentUserScope() !== scope) throw new Error('Session changed')
    return new Promise((resolve, reject) => {
      const transaction = db.transaction(storeName, mode)
      const request = operation(transaction.objectStore(storeName))
      let result: T | undefined
      if (request) {
        request.onsuccess = () => { result = request.result }
        request.onerror = () => reject(request.error)
      }
      transaction.oncomplete = () => {
        if (currentUserScope() !== scope) reject(new Error('Session changed'))
        else resolve(result)
      }
      transaction.onerror = () => reject(transaction.error)
      transaction.onabort = () => reject(transaction.error || new Error('Offline transaction aborted'))
    })
  }
}
