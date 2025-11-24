module.exports = {
  // Lint & format TypeScript and JavaScript files
  '*.{js,jsx,ts,tsx}': ['next lint --fix --file', 'prettier --write'],

  // Format other files
  '*.{json,md,yml,yaml,css}': ['prettier --write'],
}
