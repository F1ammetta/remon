let currentService = null
let logsInterval = null
let inlineLogsInterval = null
let lastLogsContent = ""
let lastInlineLogsContent = ""
const htmx = window.htmx

// Theme management
function initTheme() {
  const savedTheme = localStorage.getItem("theme") || "light"
  document.documentElement.setAttribute("data-theme", savedTheme)
  updateThemeIcon(savedTheme)
}

function toggleTheme() {
  const currentTheme = document.documentElement.getAttribute("data-theme") || "light"
  const newTheme = currentTheme === "light" ? "dark" : "light"

  document.documentElement.setAttribute("data-theme", newTheme)
  localStorage.setItem("theme", newTheme)
  updateThemeIcon(newTheme)
}

function updateThemeIcon(theme) {
  const themeButton = document.querySelector(".theme-toggle")
  if (themeButton) {
    themeButton.innerHTML =
      theme === "light"
        ? `<svg width="16" height="16" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
        <path d="M21 12.79A9 9 0 1 1 11.21 3 7 7 0 0 0 21 12.79z" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"></path>
      </svg>`
        : `<svg width="16" height="16" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
        <circle cx="12" cy="12" r="5" stroke="currentColor" stroke-width="2"></circle>
        <line x1="12" y1="1" x2="12" y2="3" stroke="currentColor" stroke-width="2"></line>
        <line x1="12" y1="21" x2="12" y2="23" stroke="currentColor" stroke-width="2"></line>
        <line x1="4.22" y1="4.22" x2="5.64" y2="5.64" stroke="currentColor" stroke-width="2"></line>
        <line x1="18.36" y1="18.36" x2="19.78" y2="19.78" stroke="currentColor" stroke-width="2"></line>
        <line x1="1" y1="12" x2="3" y2="12" stroke="currentColor" stroke-width="2"></line>
        <line x1="21" y1="12" x2="23" y2="12" stroke="currentColor" stroke-width="2"></line>
        <line x1="4.22" y1="19.78" x2="5.64" y2="18.36" stroke="currentColor" stroke-width="2"></line>
        <line x1="18.36" y1="5.64" x2="19.78" y2="4.22" stroke="currentColor" stroke-width="2"></line>
      </svg>`
  }
}

// Service selection
function selectService(element) {
  document.querySelectorAll(".service-item").forEach((item) => {
    item.classList.remove("selected")
  })

  element.classList.add("selected")
  currentService = element.dataset.serviceName
}

// Inline logs management
function refreshInlineLogs() {
  if (!currentService) return

  const logsContent = document.getElementById("logs-content-inline")
  const lines = document.getElementById("log-lines-inline").value

  if (logsContent) {
    const wasAtBottom = isScrolledToBottom(logsContent)

    htmx
      .ajax("GET", `/api/services/${currentService}/logs?lines=${lines}`, {
        target: "#logs-content-inline",
        swap: "innerHTML",
      })
      .then(() => {
        // Scroll to bottom if user was at bottom or it's the first load
        if (wasAtBottom || lastInlineLogsContent === "") {
          setTimeout(() => {
            scrollToBottom(logsContent)
          }, 50)
        }
      })
  }
}

function toggleInlineLogsFollow() {
  const followCheckbox = document.getElementById("follow-logs-inline")

  if (inlineLogsInterval) {
    clearInterval(inlineLogsInterval)
    inlineLogsInterval = null
  }

  if (followCheckbox && followCheckbox.checked && currentService) {
    inlineLogsInterval = setInterval(() => {
      refreshInlineLogs()
    }, 5000)
  }
}

// Show logs modal
function showLogs(serviceName) {
  currentService = serviceName
  const modal = document.getElementById("logs-modal")
  const overlay = document.getElementById("modal-overlay")
  const title = document.getElementById("logs-title")

  title.textContent = `${serviceName} - Service Logs`
  modal.classList.add("active")
  overlay.classList.add("active")

  // Reset last content
  lastLogsContent = ""

  // Load initial logs
  loadLogs(true)

  // Set up auto-refresh if follow is enabled
  setupLogsAutoRefresh()
}

// Close logs modal
function closeLogs() {
  const modal = document.getElementById("logs-modal")
  const overlay = document.getElementById("modal-overlay")

  modal.classList.remove("active")
  overlay.classList.remove("active")

  // Clear auto-refresh
  if (logsInterval) {
    clearInterval(logsInterval)
    logsInterval = null
  }

  lastLogsContent = ""
}

// Load logs for current service
function loadLogs(isInitialLoad = false) {
  if (!currentService) return

  const logsContent = document.getElementById("logs-content")
  const lines = document.getElementById("log-lines").value

  fetch(`/api/services/${currentService}/logs?lines=${lines}`)
    .then((response) => response.text())
    .then((data) => {
      if (data !== lastLogsContent) {
        const wasAtBottom = isScrolledToBottom(logsContent)

        logsContent.textContent = data
        lastLogsContent = data

        if (wasAtBottom || isInitialLoad) {
          scrollToBottom(logsContent)
        }
      }
    })
    .catch((error) => {
      const errorMsg = `Error loading logs: ${error.message}`
      if (errorMsg !== lastLogsContent) {
        logsContent.textContent = errorMsg
        lastLogsContent = errorMsg
      }
    })
}

// Check if user is scrolled to bottom
function isScrolledToBottom(element) {
  const threshold = 50 // pixels from bottom
  return element.scrollTop >= element.scrollHeight - element.clientHeight - threshold
}

// Scroll to bottom smoothly
function scrollToBottom(element) {
  element.scrollTop = element.scrollHeight
}

// Refresh logs manually
function refreshLogs() {
  lastLogsContent = ""
  loadLogs(true)
}

// Setup auto-refresh for logs
function setupLogsAutoRefresh() {
  const followCheckbox = document.getElementById("follow-logs")

  if (logsInterval) {
    clearInterval(logsInterval)
    logsInterval = null
  }

  if (followCheckbox.checked) {
    logsInterval = setInterval(() => loadLogs(false), 5000)
  }
}

// Refresh services list
function refreshServicesList() {
  const servicesList = document.getElementById("services-list")
  if (servicesList) {
    htmx.ajax("GET", "/api/services", {
      target: "#services-list",
      swap: "innerHTML",
    })
  }
}

// Event listeners
document.addEventListener("DOMContentLoaded", () => {
  // Initialize theme
  initTheme()

  // Close modal when clicking overlay
  document.getElementById("modal-overlay").addEventListener("click", (e) => {
    if (e.target === e.currentTarget) {
      if (document.getElementById("add-service-modal").classList.contains("active")) {
        closeAddServiceModal()
      } else {
        closeLogs()
      }
    }
  })

  // Handle follow logs checkbox (modal)
  const followLogsCheckbox = document.getElementById("follow-logs")
  if (followLogsCheckbox) {
    followLogsCheckbox.addEventListener("change", setupLogsAutoRefresh)
  }

  // Handle log lines selection change (modal)
  const logLinesSelect = document.getElementById("log-lines")
  if (logLinesSelect) {
    logLinesSelect.addEventListener("change", () => {
      lastLogsContent = ""
      loadLogs(true)
    })
  }

  // Handle escape key to close modal
  document.addEventListener("keydown", (e) => {
    if (e.key === "Escape") {
      if (document.getElementById("logs-modal").classList.contains("active")) {
        closeLogs()
      } else if (document.getElementById("add-service-modal").classList.contains("active")) {
        closeAddServiceModal()
      }
    }
  })

  // Auto-refresh services list every 30 seconds
  setInterval(() => {
    if (
      !document.getElementById("logs-modal").classList.contains("active") &&
      !document.getElementById("add-service-modal").classList.contains("active")
    ) {
      refreshServicesList()
    }
  }, 30000)
})

// HTMX event handlers
document.body.addEventListener("htmx:afterRequest", (event) => {
  if (event.detail.xhr.status === 200) {
    if (event.detail.target.id === "services-list") {
      console.log("Services list updated")
      // Restore selection if we have a current service
      if (currentService) {
        const serviceItem = document.querySelector(`[data-service-name="${currentService}"]`)
        if (serviceItem) {
          serviceItem.classList.add("selected")
        }
      }
    } else if (event.detail.target.id === "service-details") {
      console.log("Service details updated")
      showNotification("Action completed successfully", "success")
      // Also refresh the services list to update status
      refreshServicesList()

      // Setup inline logs auto-refresh if follow is enabled
      setTimeout(() => {
        toggleInlineLogsFollow()
        // Scroll inline logs to bottom on initial load
        const inlineLogsContent = document.getElementById("logs-content-inline")
        if (inlineLogsContent) {
          scrollToBottom(inlineLogsContent)
        }
      }, 100)
    } else if (event.detail.target.id === "logs-content-inline") {
      console.log("Inline logs updated")
      const logsContent = document.getElementById("logs-content-inline")
      if (logsContent) {
        const wasAtBottom = isScrolledToBottom(logsContent)
        // Always scroll to bottom on first load or if user was at bottom
        if (wasAtBottom || lastInlineLogsContent === "") {
          scrollToBottom(logsContent)
        }
        lastInlineLogsContent = logsContent.textContent
      }
    }
  }
})

document.body.addEventListener("htmx:responseError", (event) => {
  console.error("Request failed:", event.detail.xhr.responseText)
  showNotification("Operation failed. Please try again.", "error")
})

// Simple notification system
function showNotification(message, type = "info") {
  const existingNotifications = document.querySelectorAll(".notification")
  existingNotifications.forEach((n) => n.remove())

  const notification = document.createElement("div")
  notification.className = `notification notification-${type}`
  notification.textContent = message

  notification.style.cssText = `
    position: fixed;
    top: 20px;
    right: 20px;
    padding: 12px 20px;
    border-radius: 8px;
    color: white;
    font-weight: 500;
    z-index: 10000;
    max-width: 300px;
    box-shadow: 0 4px 12px rgba(0,0,0,0.3);
    transition: all 0.3s ease;
  `

  switch (type) {
    case "success":
      notification.style.backgroundColor = "#1ea885"
      break
    case "error":
      notification.style.backgroundColor = "#dc3545"
      break
    case "warning":
      notification.style.backgroundColor = "#ff9500"
      break
    default:
      notification.style.backgroundColor = "#6d4aff"
  }

  document.body.appendChild(notification)

  setTimeout(() => {
    notification.style.opacity = "0"
    notification.style.transform = "translateX(100%)"
    setTimeout(() => notification.remove(), 300)
  }, 3000)
}

// Add Service Modal Functions
function showAddServiceModal() {
  const modal = document.getElementById("add-service-modal")
  const overlay = document.getElementById("modal-overlay")

  modal.classList.add("active")
  overlay.classList.add("active")

  document.getElementById("add-service-form").reset()
  const result = document.getElementById("add-service-result")
  result.style.display = "none"
  result.className = ""

  setTimeout(() => {
    document.getElementById("service-name").focus()
  }, 100)
}

function closeAddServiceModal() {
  const modal = document.getElementById("add-service-modal")
  const overlay = document.getElementById("modal-overlay")

  modal.classList.remove("active")
  overlay.classList.remove("active")
}

function handleAddServiceResponse(event) {
  const result = document.getElementById("add-service-result")
  const xhr = event.detail.xhr

  if (xhr.status === 200) {
    result.className = "success"
    result.textContent = "Service added successfully!"
    result.style.display = "block"

    refreshServicesList()

    setTimeout(() => {
      closeAddServiceModal()
      showNotification("Service added successfully!", "success")
    }, 1000)
  } else {
    result.className = "error"
    result.textContent = xhr.responseText || "Failed to add service"
    result.style.display = "block"
  }
}

