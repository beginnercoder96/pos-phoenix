/** @type {import('tailwindcss').Config} */
module.exports = {
  darkMode: ['selector', '[data-theme="dark"]'],
  content: ["./web/templates/**/*.html", "./web/static/js/**/*.js", "./internal/**/*.go"],
  theme: { extend: {} },
  plugins: [],
};
