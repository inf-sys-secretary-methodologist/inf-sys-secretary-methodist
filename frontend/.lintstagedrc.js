module.exports = {
  // Lint & format TypeScript and JavaScript files
  '*.{js,jsx,ts,tsx}': ['eslint --fix', 'prettier --write'],

  // Format other files
  '*.{json,md,yml,yaml,css}': ['prettier --write'],
}
