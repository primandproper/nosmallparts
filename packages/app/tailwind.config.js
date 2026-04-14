/** @type {import('tailwindcss').Config} */
module.exports = {
  content: ['./app/**/*.{js,jsx,ts,tsx}', './components/**/*.{js,jsx,ts,tsx}'],
  presets: [require('nativewind/preset')],
  theme: {
    extend: {
      colors: {
        stage: {
          bg: '#0a0a0a',
          surface: '#141414',
          border: '#2a2a2a',
          gold: '#d4af37',
          goldDim: '#9a7e27',
          text: '#ede6d0',
          muted: '#8a8374',
        },
      },
      fontFamily: {
        serif: ['Cormorant Garamond', 'Georgia', 'serif'],
      },
    },
  },
  plugins: [],
};
