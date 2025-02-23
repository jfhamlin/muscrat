/** @type {import('tailwindcss').Config} */
const { tailwindTheme } = require('./src/theme');

export default {
  content: [
    "./index.html",
    "./src/**/*.{js,jsx,ts,tsx}",
  ],
  theme: {
    extend: {
        colors: tailwindTheme,
    },
  },
  plugins: [],
}
