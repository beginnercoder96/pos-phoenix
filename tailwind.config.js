/** @type {import('tailwindcss').Config} */
module.exports = {
  darkMode: ['selector', '[data-theme="dark"]'],
  content: ["./web/templates/**/*.html", "./web/static/js/**/*.js", "./internal/**/*.go"],
  theme: {
    extend: {
      colors: {
        brand: {
          dark: '#000210',
          navy: '#010521',
          surface: '#0a1130',
          accent: '#10b981',
        },
      },
      fontFamily: {
        mono: ['ui-monospace', 'SFMono-Regular', 'Menlo', 'Monaco', 'Consolas', '"Liberation Mono"', '"Courier New"', 'monospace'],
      },
      boxShadow: {
        crisp: '0 1px 2px 0 rgba(0, 0, 0, 0.05)',
        card: '0 4px 6px -1px rgba(0, 0, 0, 0.1), 0 2px 4px -2px rgba(0, 0, 0, 0.1)',
        elevated: '0 10px 15px -3px rgba(0, 0, 0, 0.1), 0 4px 6px -4px rgba(0, 0, 0, 0.1)',
      },
      spacing: {
        '72': '18rem', // 288px for pos-sidebar
      },
      minHeight: {
        '11': '2.75rem', // 44px touch target
      },
      minWidth: {
        '11': '2.75rem', // 44px touch target
      },
    },
  },
  plugins: [],
};
