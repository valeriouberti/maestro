/** @type {import('tailwindcss').Config} */
module.exports = {
  content: ["./src/**/*.{js,jsx,ts,tsx}"],
  theme: {
    extend: {
      colors: {
        "pastel-blue": "#A7C4BC",
        "pastel-green": "#CCE2CB",
        "pastel-yellow": "#FDE5D4",
        "pastel-orange": "#F4BFBF",
        "pastel-red": "#CDB4DB",
        "pastel-purple": "#B8C0FF",
        "pastel-white": "#F8F9FA",
        "pastel-gray": "#E9ECEF",
        "sidebar-blue": "#EBF4FF", // Subtle blue for sidebar
        "sidebar-hover": "#DBEAFE", // Slightly darker for hover states
        "accent-blue": "#3B82F6", // For accent elements
      },
    },
  },
  plugins: [],
};
