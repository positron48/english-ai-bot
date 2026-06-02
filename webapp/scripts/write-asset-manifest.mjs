import { readdir, writeFile } from 'node:fs/promises'
import { join, relative } from 'node:path'

const distDir = new URL('../dist', import.meta.url).pathname
const assetDir = join(distDir, 'assets')

async function listFiles(dir) {
  const entries = await readdir(dir, { withFileTypes: true })
  const files = []
  for (const entry of entries) {
    const fullPath = join(dir, entry.name)
    if (entry.isDirectory()) {
      files.push(...await listFiles(fullPath))
    } else {
      files.push(fullPath)
    }
  }
  return files
}

const files = await listFiles(assetDir)
const assets = files
  .map((file) => `/app/${relative(distDir, file).replaceAll('\\', '/')}`)
  .sort()

await writeFile(
  join(distDir, 'asset-manifest.json'),
  JSON.stringify({ generatedAt: new Date().toISOString(), assets }, null, 2),
  'utf8'
)

console.log(`Wrote asset-manifest.json with ${assets.length} assets`)
