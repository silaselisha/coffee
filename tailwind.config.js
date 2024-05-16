/** @type {import('tailwindcss').Config} */
module.exports = {
  content: ["./views/**/*.html"],
  theme: {
    extend: {
      fontFamily: {
        "futura-light": ["Futura light", "sans-serif"],
        "futura-regular": ["Futura regular", "sans-serif"],
        "futura-medium": ["Futura medium", "sans-serif"],
        "futura-bold": ["Futura bold", "sans-serif"],
      },
    },
  },
  plugins: [],
}

