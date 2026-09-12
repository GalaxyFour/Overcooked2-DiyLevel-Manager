/** @type {import('tailwindcss').Config} */
export default {
  content: ['./index.html', './src/**/*.{js,ts,jsx,tsx}'],
  darkMode: 'class',
  theme: {
    extend: {
      colors: {
        tomato: '#E85D4C',
        cream: '#FFF8E7',
        mint: '#7EC8A3',
        kitchen: {
          dark: '#1a1f2e',
          card: '#252b3d',
        },
      },
      fontFamily: {
        cute: ['"Nunito"', 'system-ui', 'sans-serif'],
      },
      boxShadow: {
        soft: '0 8px 32px rgba(232, 93, 76, 0.12)',
      },
    },
  },
  plugins: [],
}
