#!/usr/bin/env node

/**
 * PWA Icon Generator Script
 *
 * This script generates all required PWA icons from a source SVG file.
 * Run: node scripts/generate-icons.js
 *
 * Requirements: npm install sharp
 */

import fs from 'fs'
import path from 'path'
import { fileURLToPath } from 'url'

const __filename = fileURLToPath(import.meta.url)
const __dirname = path.dirname(__filename)

// Check if sharp is available
let sharp
try {
  sharp = (await import('sharp')).default
} catch {
  console.warn('Sharp not installed. Creating placeholder icons instead.')
  console.warn('To generate real icons, run: npm install -D sharp')
  console.warn('Then run this script again.')

  // Create placeholder icons using simple colored squares
  createPlaceholderIcons()
  process.exit(0)
}

const SOURCE_SVG = path.join(__dirname, '../public/icons/icon.svg')
const OUTPUT_DIR = path.join(__dirname, '../public/icons')

// Icon sizes needed for PWA
const ICON_SIZES = [16, 32, 72, 96, 128, 144, 152, 180, 192, 384, 512]

async function generateIcons() {
  // Ensure output directory exists
  if (!fs.existsSync(OUTPUT_DIR)) {
    fs.mkdirSync(OUTPUT_DIR, { recursive: true })
  }

  // Check if source SVG exists
  if (!fs.existsSync(SOURCE_SVG)) {
    console.error(`Source SVG not found: ${SOURCE_SVG}`)
    process.exit(1)
  }

  console.warn('Generating PWA icons...')

  for (const size of ICON_SIZES) {
    const outputPath = path.join(OUTPUT_DIR, `icon-${size}x${size}.png`)

    try {
      await sharp(SOURCE_SVG).resize(size, size).png().toFile(outputPath)

      console.warn(`  Created: icon-${size}x${size}.png`)
    } catch (error) {
      console.error(`  Failed to create icon-${size}x${size}.png:`, error.message)
    }
  }

  // Generate Apple Touch Icon (180x180)
  const appleTouchIconPath = path.join(OUTPUT_DIR, 'apple-touch-icon.png')
  await sharp(SOURCE_SVG).resize(180, 180).png().toFile(appleTouchIconPath)
  console.warn('  Created: apple-touch-icon.png')

  // Generate favicon.ico (multi-size ico file)
  // Note: sharp doesn't support ICO, so we'll create a 32x32 PNG as fallback
  const faviconPath = path.join(__dirname, '../public/favicon.ico')
  await sharp(SOURCE_SVG).resize(32, 32).png().toFile(faviconPath.replace('.ico', '.png'))
  // Copy as .ico (browsers can handle PNG as favicon)
  fs.copyFileSync(faviconPath.replace('.ico', '.png'), faviconPath)
  console.warn('  Created: favicon.ico (as PNG)')

  console.warn('\nIcon generation complete!')
  console.warn('Note: For production, consider using a proper favicon generator')
  console.warn('like https://realfavicongenerator.net/ for optimal results.')
}

function createPlaceholderIcons() {
  const outputDir = path.join(__dirname, '../public/icons')

  if (!fs.existsSync(outputDir)) {
    fs.mkdirSync(outputDir, { recursive: true })
  }

  // Create a simple SVG placeholder for each size
  const iconSizes = [16, 32, 72, 96, 128, 144, 152, 180, 192, 384, 512]

  console.warn('Creating placeholder icons...')

  // Generate a simple colored square SVG as placeholder
  const createPlaceholderSVG = (size) => `<?xml version="1.0" encoding="UTF-8"?>
<svg xmlns="http://www.w3.org/2000/svg" width="${size}" height="${size}" viewBox="0 0 ${size} ${size}">
  <rect width="${size}" height="${size}" fill="#3b82f6" rx="${Math.round(size * 0.15)}"/>
  <text x="50%" y="55%" font-family="system-ui, sans-serif" font-size="${Math.round(size * 0.4)}"
        fill="white" text-anchor="middle" dominant-baseline="middle" font-weight="bold">СМ</text>
</svg>`

  for (const size of iconSizes) {
    const svgContent = createPlaceholderSVG(size)
    const outputPath = path.join(outputDir, `icon-${size}x${size}.svg`)
    fs.writeFileSync(outputPath, svgContent)
    console.warn(`  Created placeholder: icon-${size}x${size}.svg`)
  }

  // Create apple-touch-icon
  fs.writeFileSync(path.join(outputDir, 'apple-touch-icon.svg'), createPlaceholderSVG(180))
  console.warn('  Created placeholder: apple-touch-icon.svg')

  // Create shortcut icons
  const shortcutIcons = ['shortcut-documents', 'shortcut-calendar', 'shortcut-notifications']
  const shortcutColors = ['#3b82f6', '#10b981', '#f59e0b']
  const shortcutEmojis = ['D', 'K', 'Y']

  shortcutIcons.forEach((name, index) => {
    const svg = `<?xml version="1.0" encoding="UTF-8"?>
<svg xmlns="http://www.w3.org/2000/svg" width="96" height="96" viewBox="0 0 96 96">
  <rect width="96" height="96" fill="${shortcutColors[index]}" rx="14"/>
  <text x="50%" y="55%" font-family="system-ui, sans-serif" font-size="40"
        fill="white" text-anchor="middle" dominant-baseline="middle" font-weight="bold">${shortcutEmojis[index]}</text>
</svg>`
    fs.writeFileSync(path.join(outputDir, `${name}.svg`), svg)
    console.warn(`  Created placeholder: ${name}.svg`)
  })

  console.warn('\nPlaceholder icons created!')
  console.warn(
    'For production, install sharp and regenerate: npm install -D sharp && node scripts/generate-icons.js'
  )
}

generateIcons().catch(console.error)
