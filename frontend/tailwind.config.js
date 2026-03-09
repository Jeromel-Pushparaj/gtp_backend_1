/** @type {import('tailwindcss').Config} */
export default {
  content: [
    "./index.html",
    "./src/**/*.{js,ts,jsx,tsx}",
  ],
  theme: {
    extend: {
      colors: {
        slack: {
          purple: '#4A154B',
          green: '#2EB67D',
          blue: '#36C5F0',
          yellow: '#ECB22E',
          red: '#E01E5A',
        }
      }
    },
  },
  plugins: [],
}

