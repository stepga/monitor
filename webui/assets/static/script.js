let pendingDeleteElement = null;

function getStorage(key) {
	switch(key) {
		case "verbose":
			return sessionStorage.getItem(key) === "true";
		default:
			console.error("unknown variable name: " + key);
	}
}

function setStorage(key, val) {
	switch(key) {
		case "verbose":
			return sessionStorage.setItem("verbose", val === true);
		default:
			console.error("unknown variable name: " + key);
	}
}

function displayVerboseDiv(doDisplay) {
	const verboseDiv = document.getElementById("verbose");
	verboseDiv.style.display = doDisplay ? "block" : "none";
}

function verboseCheckboxOnChange(event) {
	setStorage("verbose", event.target.checked);
	displayVerboseDiv(event.target.checked);
}

function setNotificationsCritical() {
	const criticalDiv = document.getElementById("critical");

	fetch("/critical")
		.then(response => response.json())
		.then(data => {
			if (!Array.isArray(data)) {
				console.error("invalid data:", data);
				return;
			}

			// remove vanished
			criticalDiv.querySelectorAll("details").forEach(detail => {
				const exists = data.some(obj => obj.identifier === detail.id);
				if (!exists) detail.remove();
			});

			// add new
			data.forEach(obj => {
				if (!document.getElementById(obj.identifier)) {
					const notification = createNotification(obj, true);
					criticalDiv.prepend(notification);
				}
			});
		})
		.catch(err => console.error("fetch /critical failed:", err));
}

function openConfirmModal(targetElement) {
	pendingDeleteElement = targetElement;
	const modalNotificationName = document.getElementById("modalNotificationName");
	modalNotificationName.textContent = targetElement.id;
	const modal = document.getElementById("confirmModal");
	modal.classList.remove("hidden");
}

function closeConfirmModal() {
	pendingDeleteElement = null;
	const modalNotificationName = document.getElementById("modalNotificationName");
	modalNotificationName.textContent = "";
	const modal = document.getElementById("confirmModal");
	modal.classList.add("hidden");
}

function setupModal() {
	const modal = document.getElementById("confirmModal");
	const btnYes = document.getElementById("confirmYes");
	const btnNo = document.getElementById("confirmNo");

	btnYes.addEventListener("click", () => {
		if (pendingDeleteElement) {
			pendingDeleteElement.remove();

			// TODO: backend sync:
			// fetch(`/critical/${pendingDeleteElement.id}`, { method: "DELETE" });
		}
		closeConfirmModal();
	});

	btnNo.addEventListener("click", closeConfirmModal);

	// close modal when clicking outside
	modal.addEventListener("click", (e) => {
		if (e.target === modal) { closeConfirmModal(); }
	});
}

function createNotification(data, isCritical) {
	const timestamp = document.createElement("span");
	timestamp.classList.add("timestamp");
	timestamp.textContent = data.timestamp;

	const summaryText = document.createElement("span");
	summaryText.classList.add("summary");
	summaryText.textContent = data.summary;

	const summary = document.createElement("summary");
	summary.appendChild(timestamp);
	summary.appendChild(summaryText);

	// add delete button for critical notifications
	if (isCritical) {
		const btn = document.createElement("button");
		btn.textContent = "✕";
		btn.classList.add("delete-btn");
		btn.addEventListener("click", (e) => {
			e.preventDefault();
			e.stopPropagation();

			const detail = summary.parentElement;
			openConfirmModal(detail);
		});
		summary.appendChild(btn);
	}

	const pre = document.createElement("pre");
	pre.textContent = data.details;

	const detailsDiv = document.createElement("div");
	detailsDiv.classList.add("details");
	detailsDiv.appendChild(pre);

	const detail = document.createElement("details");
	detail.appendChild(summary);
	detail.appendChild(detailsDiv);

	if (isCritical) {
		detail.id = data.identifier;
	}

	return detail;
}

document.addEventListener("DOMContentLoaded", function() {
	const eventSource = new EventSource("/notifications");

	window.addEventListener("beforeunload", () => {
		// prevent js error on page reload
		// see: https://bugzilla.mozilla.org/show_bug.cgi?id=833462
		eventSource.close();
	});

	const verboseCheckbox = document.getElementById("verboseCheckbox");
	verboseCheckbox.addEventListener("change", verboseCheckboxOnChange);
	verboseCheckbox.checked = getStorage("verbose");

	const verboseDiv = document.getElementById("verbose");
	displayVerboseDiv(getStorage("verbose"));

	eventSource.onmessage = (event) => {
		try {
			const data = JSON.parse(event.data);
			const notification = createNotification(data, false);
			verboseDiv.prepend(notification);
			setNotificationsCritical();
		} catch (err) {
			console.error("message error:", err);
		}
	};

	setupModal();
	setNotificationsCritical();
});
