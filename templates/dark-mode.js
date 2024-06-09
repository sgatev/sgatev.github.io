function darkModeEnabled() {
  return localStorage.getItem("darkModeEnabled") === "true";
}

function setDarkMode(enabled) {
  document.body.classList.toggle("dark-mode", enabled);
  document.getElementById("dark-mode-toggle").style.display = "block";
  localStorage.setItem("darkModeEnabled", enabled ? "true" : "false");
}

function toggleDarkMode() {
  setDarkMode(!darkModeEnabled());
}

setDarkMode(darkModeEnabled());
