/** @type {import('tailwindcss').Config} */
module.exports = {
  content: ["./**/*.{html,js,templ,go}"],
  theme: {
    extend: {},
  },
  plugins: [require("daisyui")],
  daisyui: {
    themes: ["lofi"],
  },
};
