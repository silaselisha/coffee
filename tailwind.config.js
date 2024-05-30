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
      backgroundImage: {
        "hero-pattern": "url('/public/images/hanging_in_the_tree.jpg')",
        "info-classic-blend": "url('/public/images/info-classic-blend.jpg')",
        "info-rare": "url('/public/images/info-rare.jpg')",
        "info-single-origin": "url('/public/images/info-single-origin.jpg')",
      },
    },
  },
  plugins: [],
};
