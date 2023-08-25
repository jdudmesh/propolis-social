/** @type {import('tailwindcss').Config} */
module.exports = {
  content: ["./ui/views/**/*.html"],
  theme: {
    fontFamily: {
      sans: ["IBM Plex Mono", "Helvetica", "Verdana", "sans-serif"],
    },
    extend: {
      width: {
        "144": "36rem",
      }
    },
  },
  plugins: [],
}

