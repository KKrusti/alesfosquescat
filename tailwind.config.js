/** @type {import('tailwindcss').Config} */
export default {
  content: [
    './index.html',
    './src/**/*.{js,ts,jsx,tsx}',
  ],
  theme: {
    extend: {
      fontFamily: {
        anton: ['Anton', 'Impact', 'sans-serif'],
        mono: ['JetBrains Mono', 'Fira Code', 'monospace'],
      },
      colors: {
        night: {
          950: '#020208',
          900: '#050510',
          800: '#0a0a1e',
          700: '#0f0f2e',
          600: '#16163e',
        },
        signal: {
          400: '#fde68a',
          500: '#fcd34d',
          600: '#fbbf24',
          700: '#f59e0b',
        },
      },
      animation: {
        oscillate: 'oscillate 5s ease-in-out infinite',
        shake: 'shake 0.6s ease-in-out',
        'pulse-beam': 'pulse-beam 2s ease-in-out forwards',
        blink: 'blink 0.4s ease-in-out 6 forwards',
        twinkle: 'twinkle 4s ease-in-out infinite',
        'fade-in': 'fade-in 0.8s ease-out forwards',
        'slide-up': 'slide-up 0.6s ease-out forwards',
        'count-in': 'count-in 0.4s ease-out forwards',
      },
      keyframes: {
        oscillate: {
          '0%, 100%': { transform: 'rotate(-15deg)' },
          '50%': { transform: 'rotate(15deg)' },
        },
        shake: {
          '0%, 100%': { transform: 'translateX(0)' },
          '10%, 30%, 50%, 70%, 90%': { transform: 'translateX(-10px)' },
          '20%, 40%, 60%, 80%': { transform: 'translateX(10px)' },
        },
        'pulse-beam': {
          '0%':   { opacity: '0.45' },
          '16%':  { opacity: '1' },
          '33%':  { opacity: '0.45' },
          '50%':  { opacity: '1' },
          '66%':  { opacity: '0.45' },
          '83%':  { opacity: '1' },
          '100%': { opacity: '0.45' },
        },
        blink: {
          '0%, 100%': { opacity: '1' },
          '50%': { opacity: '0.08' },
        },
        twinkle: {
          '0%, 100%': { opacity: '1' },
          '50%': { opacity: '0.2' },
        },
        'fade-in': {
          from: { opacity: '0' },
          to: { opacity: '1' },
        },
        'slide-up': {
          from: { transform: 'translateY(24px)', opacity: '0' },
          to: { transform: 'translateY(0)', opacity: '1' },
        },
        'count-in': {
          from: { transform: 'scale(0.8)', opacity: '0' },
          to: { transform: 'scale(1)', opacity: '1' },
        },
      },
    },
  },
  plugins: [],
}
