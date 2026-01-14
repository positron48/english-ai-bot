<template>
  <div class="admin-layout">
    <AdminMenu />
    
    <div class="admin-content">
      <div class="db-schema-view">
      <div class="header">
        <h2>Database Schema</h2>
        <button @click="loadSchema" :disabled="loading" class="refresh-btn">
          {{ loading ? 'Loading...' : 'Refresh' }}
        </button>
      </div>

    <div v-if="error" class="error-message">
      Error loading schema: {{ error }}
    </div>

    <div v-if="schema && !loading" class="schema-container">
      <div class="controls">
        <label>
          <input type="checkbox" v-model="showColumnTypes" />
          Show column types
        </label>
        <label>
          <input type="checkbox" v-model="showPrimaryKeys" />
          Highlight primary keys
        </label>
        <label>
          <input type="checkbox" v-model="showForeignKeys" checked />
          Show relationships
        </label>
      </div>

      <div ref="networkContainer" class="network-container"></div>

      <div class="table-list">
        <h2>Tables ({{ schema?.tables?.length || 0 }})</h2>
        <div class="tables-grid">
          <div
            v-for="table in schema.tables"
            :key="table.name"
            class="table-card"
            @click="focusTable(table.name)"
          >
            <h3>{{ table.name }}</h3>
            <div class="table-info">
              <span class="columns-count">{{ table.columns?.length || 0 }} columns</span>
              <span class="fks-count" v-if="table.foreign_keys && table.foreign_keys.length > 0">
                {{ table.foreign_keys.length }} {{ table.foreign_keys.length === 1 ? 'relation' : 'relations' }}
              </span>
            </div>
            <div class="columns-preview">
              <div
                v-for="col in (table.columns || [])"
                :key="col.name"
                class="column-item"
                :class="{ 'primary-key': col.primary_key }"
              >
                <span class="col-name">{{ col.name }}</span>
                <span class="col-type">{{ col.type }}</span>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
    </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, onUnmounted, watch, nextTick } from 'vue'
import { Network } from 'vis-network'
import { DataSet } from 'vis-data'
import type { Node, Edge } from 'vis-network'
import { apiClient } from '../api/client'
import AdminMenu from '../components/AdminMenu.vue'

interface TableColumn {
  name: string
  type: string
  not_null: boolean
  default_value?: string
  primary_key: boolean
}

interface ForeignKey {
  from_table: string
  from_column: string
  to_table: string
  to_column: string
  on_delete?: string
}

interface TableInfo {
  name: string
  columns: TableColumn[]
  foreign_keys: ForeignKey[]
}

interface SchemaResponse {
  tables: TableInfo[]
}

const schema = ref<SchemaResponse | null>(null)
const loading = ref(false)
const error = ref<string | null>(null)
const networkContainer = ref<HTMLElement | null>(null)
const network = ref<Network | null>(null)
const showColumnTypes = ref(false)
const showPrimaryKeys = ref(true)
const showForeignKeys = ref(true)
const isDarkTheme = ref(false)

// Check for dark theme
const checkDarkTheme = () => {
  if (window.matchMedia) {
    isDarkTheme.value = window.matchMedia('(prefers-color-scheme: dark)').matches
  }
  // Also check if body has dark class or data attribute
  const body = document.body
  if (body.classList.contains('dark') || body.getAttribute('data-theme') === 'dark') {
    isDarkTheme.value = true
  }
}

const loadSchema = async () => {
  loading.value = true
  error.value = null

  try {
    const response = await apiClient.request<SchemaResponse>('/app/admin/db-schema')
    schema.value = response
    // Wait for DOM to be ready and ensure networkContainer is available
    await nextTick()
    // Add small delay to ensure container is fully rendered
    setTimeout(() => {
      renderNetwork()
    }, 100)
  } catch (err) {
    error.value = err instanceof Error ? err.message : 'Unknown error'
  } finally {
    loading.value = false
  }
}

const renderNetwork = () => {
  if (!networkContainer.value || !schema.value) {
    // Retry after a short delay if container is not ready
    setTimeout(() => {
      if (networkContainer.value && schema.value) {
        renderNetwork()
      }
    }, 100)
    return
  }

  // Create nodes for each table
  const nodes = new DataSet<Node>(
    schema.value.tables.map((table, index) => {
      const pkColumns = table.columns.filter(c => c.primary_key)
      const fkCount = table.foreign_keys.length
      
      // Create label with table name and all columns
      let label = table.name
      if (showPrimaryKeys.value && pkColumns.length > 0) {
        label += `\n[PK: ${pkColumns.map(c => c.name).join(', ')}]`
      }
      
      // Add all columns
      if (table.columns.length > 0) {
        const columnsList = table.columns.map(col => {
          const pkMark = col.primary_key ? '🔑' : ''
          if (showColumnTypes.value) {
            return `${pkMark}${col.name}: ${col.type}`
          } else {
            return `${pkMark}${col.name}`
          }
        }).join('\n')
        label += `\n${columnsList}`
      } else {
        label += `\n(no columns)`
      }

      // Use CSS variables for dark theme support
      const isDark = isDarkTheme.value
      const bgColor = fkCount > 0 
        ? (isDark ? '#1e3a5f' : '#e3f2fd')
        : (isDark ? '#2d2d2d' : '#f5f5f5')
      const borderColor = fkCount > 0
        ? (isDark ? '#4a90e2' : '#2196f3')
        : (isDark ? '#555' : '#9e9e9e')
      
      return {
        id: table.name,
        label: label,
        shape: 'box',
        color: {
          background: bgColor,
          border: borderColor,
          highlight: {
            background: isDark ? '#2d5a8a' : '#bbdefb',
            border: isDark ? '#5ba3f5' : '#1976d2'
          }
        },
        font: {
          size: 14,
          face: 'monospace',
          color: isDark ? '#e0e0e0' : '#333',
          align: 'left'
        },
        margin: 10,
        widthConstraint: {
          minimum: 200,
          maximum: showColumnTypes.value ? 500 : 400
        }
      }
    })
  )

  // Create edges for foreign keys
  const edges = new DataSet<Edge>()
  
  if (showForeignKeys.value) {
    schema.value.tables.forEach(table => {
      table.foreign_keys.forEach(fk => {
        edges.add({
          id: `${fk.from_table}_${fk.from_column}_${fk.to_table}`,
          from: fk.from_table,
          to: fk.to_table,
          label: fk.from_column,
          arrows: 'to',
          color: {
            color: isDarkTheme.value ? '#888' : '#666',
            highlight: '#1976d2'
          },
          font: {
            size: 11,
            align: 'middle',
            color: isDarkTheme.value ? '#ccc' : '#333'
          },
          smooth: {
            type: 'curvedCW',
            roundness: 0.2
          }
        })
      })
    })
  }

  // Detect dark theme
  const isDark = window.matchMedia && window.matchMedia('(prefers-color-scheme: dark)').matches
  
  // Network options
  const options = {
    nodes: {
      shape: 'box',
      font: {
        size: 14,
        color: isDarkTheme.value ? '#e0e0e0' : '#333'
      },
      margin: 20,
      borderWidth: 2,
      borderWidthSelected: 3
    },
    edges: {
      arrows: {
        to: {
          enabled: true,
          scaleFactor: 1.2
        }
      },
      smooth: {
        type: 'curvedCW',
        roundness: 0.2
      }
    },
    physics: {
      enabled: true,
      stabilization: {
        enabled: true,
        iterations: 300
      },
      barnesHut: {
        gravitationalConstant: -5000,
        centralGravity: 0.05,
        springLength: 400,
        springConstant: 0.02,
        damping: 0.15
      }
    },
    interaction: {
      dragNodes: true,
      dragView: true,
      zoomView: true
    },
    layout: {
      improvedLayout: true
    }
  }

  const data = {
    nodes: nodes,
    edges: edges
  }

  // Destroy existing network if it exists
  if (network.value) {
    network.value.destroy()
  }

  network.value = new Network(networkContainer.value, data, options)

  // Add click event to highlight connected nodes
  network.value.on('click', (params) => {
    if (params.nodes.length > 0) {
      const nodeId = params.nodes[0] as string
      highlightConnections(nodeId)
    } else {
      // Reset all nodes
      const isDark = isDarkTheme.value
      nodes.forEach(node => {
        const table = schema.value!.tables.find(t => t.name === node.id)
        const fkCount = table?.foreign_keys.length || 0
        nodes.update({
          ...node,
          color: {
            background: fkCount > 0 
              ? (isDark ? '#1e3a5f' : '#e3f2fd')
              : (isDark ? '#2d2d2d' : '#f5f5f5'),
            border: fkCount > 0
              ? (isDark ? '#4a90e2' : '#2196f3')
              : (isDark ? '#555' : '#9e9e9e')
          }
        })
      })
    }
  })
}

const highlightConnections = (nodeId: string) => {
  if (!schema.value || !network.value) return

  const table = schema.value.tables.find(t => t.name === nodeId)
  if (!table) return

  const connectedTables = new Set<string>()
  connectedTables.add(nodeId)

  // Find all tables connected via foreign keys
  table.foreign_keys.forEach(fk => {
    connectedTables.add(fk.to_table)
  })

  schema.value.tables.forEach(t => {
    t.foreign_keys.forEach(fk => {
      if (fk.to_table === nodeId) {
        connectedTables.add(t.name)
      }
    })
  })

  // Update node colors
  const isDark = isDarkTheme.value
  const nodes = network.value.body.data.nodes as DataSet<Node>
  nodes.forEach(node => {
    const isConnected = connectedTables.has(node.id as string)
    const isSelected = node.id === nodeId
    
    nodes.update({
      ...node,
      color: {
        background: isSelected 
          ? (isDark ? '#2d5a8a' : '#bbdefb')
          : isConnected 
            ? (isDark ? '#1a4a6a' : '#e1f5fe')
            : (isDark ? '#2d2d2d' : '#f5f5f5'),
        border: isSelected
          ? (isDark ? '#5ba3f5' : '#1976d2')
          : isConnected
            ? (isDark ? '#4a90e2' : '#03a9f4')
            : (isDark ? '#555' : '#9e9e9e'),
        highlight: {
          background: isDark ? '#2d5a8a' : '#bbdefb',
          border: isDark ? '#5ba3f5' : '#1976d2'
        }
      }
    })
  })
}

const focusTable = (tableName: string) => {
  if (!network.value) return
  
  network.value.focus(tableName, {
    scale: 1.5,
    animation: {
      duration: 1000,
      easingFunction: 'easeInOutQuad'
    }
  })
  
  highlightConnections(tableName)
}

// Watch for changes in display options
watch([showColumnTypes, showPrimaryKeys, showForeignKeys], () => {
  if (schema.value) {
    renderNetwork()
  }
})

onMounted(() => {
  checkDarkTheme()
  
  // Listen for theme changes
  if (window.matchMedia) {
    const mediaQuery = window.matchMedia('(prefers-color-scheme: dark)')
    mediaQuery.addEventListener('change', checkDarkTheme)
  }
  
  // Also observe body class changes
  const observer = new MutationObserver(checkDarkTheme)
  observer.observe(document.body, {
    attributes: true,
    attributeFilter: ['class', 'data-theme']
  })
  
  loadSchema()
})

onUnmounted(() => {
  if (network.value) {
    network.value.destroy()
  }
})
</script>

<style scoped>
.admin-layout {
  display: flex;
  gap: 20px;
  min-height: calc(100vh - 60px);
  font-size: 16px;
}

.admin-content {
  flex: 1;
  max-width: 100%;
  margin: 0 auto;
  padding: 10px;
  width: 100%;
}

@media (max-width: 767px) {
  .admin-layout {
    flex-direction: column;
    gap: 0;
  }
  
  .admin-content {
    padding: 10px;
    margin-top: 60px;
  }
}

@media (min-width: 768px) {
  .admin-layout {
    padding: 20px;
  }
  
  .admin-content {
    padding: 0;
  }
}

.db-schema-view {
  padding: 20px;
  max-width: 100%;
  overflow-x: auto;
}

.header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
}

.header h2 {
  margin: 0;
  font-size: 24px;
}

.refresh-btn {
  padding: 8px 16px;
  background: #2196f3;
  color: white;
  border: none;
  border-radius: 4px;
  cursor: pointer;
  font-size: 14px;
}

.refresh-btn:hover:not(:disabled) {
  background: #1976d2;
}

.refresh-btn:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

.error-message {
  padding: 12px;
  background: #ffebee;
  color: #c62828;
  border-radius: 4px;
  margin-bottom: 20px;
}

.schema-container {
  display: flex;
  flex-direction: column;
  gap: 20px;
}

.controls {
  display: flex;
  gap: 20px;
  padding: 12px;
  background: var(--bg-secondary, #f5f5f5);
  border-radius: 4px;
  border: 1px solid var(--border-primary, #ddd);
}

.controls label {
  display: flex;
  align-items: center;
  gap: 8px;
  cursor: pointer;
  font-size: 14px;
  color: var(--text-primary, #333);
}

.network-container {
  width: 100%;
  height: 600px;
  border: 1px solid var(--border-primary, #ddd);
  border-radius: 4px;
  background: var(--card-bg, var(--bg-primary, #fff));
}

.table-list {
  margin-top: 20px;
}

.table-list h2 {
  margin-bottom: 16px;
  font-size: 20px;
}

.tables-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(300px, 1fr));
  gap: 16px;
}

.table-card {
  padding: 16px;
  background: var(--card-bg, var(--bg-primary, #fff));
  border: 1px solid var(--border-primary, #ddd);
  border-radius: 4px;
  cursor: pointer;
  transition: all 0.2s;
}

.table-card:hover {
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
  border-color: var(--color-primary, #2196f3);
  background: var(--bg-hover, var(--bg-secondary, #f5f5f5));
}

.table-card h3 {
  margin: 0 0 8px 0;
  font-size: 16px;
  color: var(--color-primary, #1976d2);
}

.table-info {
  display: flex;
  gap: 12px;
  margin-bottom: 12px;
  font-size: 12px;
  color: var(--text-secondary, #666);
}

.columns-count,
.fks-count {
  padding: 2px 8px;
  background: var(--bg-secondary, #e3f2fd);
  border-radius: 4px;
  color: var(--text-primary, #333);
}

.columns-preview {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.column-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 4px 8px;
  font-size: 12px;
  background: var(--bg-secondary, #f5f5f5);
  border-radius: 2px;
  color: var(--text-primary, #333);
}

.column-item.primary-key {
  background: var(--bg-hover, #fff3e0);
  font-weight: bold;
}

.col-name {
  font-family: monospace;
}

.col-type {
  color: var(--text-secondary, #666);
  font-size: 11px;
}

.more-columns {
  padding: 4px 8px;
  font-size: 11px;
  color: var(--text-secondary, #666);
  font-style: italic;
}
</style>
