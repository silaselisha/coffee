const defaultTheme = require('tailwindcss/defaultTheme');

/** @type {import('tailwindcss').Config} */
module.exports = {
  content: ["./views/**/*.html"],
  theme: {
    extend: {
      fontFamily: {
        "briem": ["Briem Hand", "cursive"],
      },
    },
  },
  plugins: [],
}

