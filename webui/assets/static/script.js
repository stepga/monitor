var verboseDetailsShown = true;

function setNotificationsCritical() {
	const notificationsCritical = document.getElementById("notifications_critical");
	fetch("/critical")
		.then(response => response.json())
		.then(data => {
			notificationsCritical.innerHTML = "";
			for (var i = 0; i < data.length ; i++) {
				notification = createNotification(data[i]);
				notificationsCritical.prepend(notification);
			}
		})
		.catch(error => {
			console.error('failed to fetch critical: ' + error);
		});
}

document.addEventListener("DOMContentLoaded", function(){
	const eventSource = new EventSource("/notifications");
	window.addEventListener('beforeunload', () => {
		// prevent js error on page reload
		// see: https://bugzilla.mozilla.org/show_bug.cgi?id=833462
		eventSource.close();
	});

	const verboseCheckbox = document.getElementById("verbose");
	const notificationsVerbose = document.getElementById("notifications_verbose");
	verboseCheckbox.addEventListener("change", function() {
		verboseDetailsShown = this.checked;
		notificationsVerbose.style.display = this.checked ? "block" : "none";
	});

	eventSource.onmessage = (event) => {
		try {
			const data = JSON.parse(event.data);
			// enforce `data['IsCritical'] = false` for notifications in the verbose log
			data['IsCritical'] = false;
			const notification = createNotification(data);
			notificationsVerbose.prepend(notification);
			setNotificationsCritical()
		} catch (error) {
			console.error('Failed to parse JSON in event:', error.message);
		}
	};

	setNotificationsCritical();
});

function createNotification(data) {
	var date = new Date();
	var timestamp = date.toLocaleTimeString()

	const detail = document.createElement("details");
	detail.innerHTML = `
		<summary>
			<span>${timestamp}</span>
			<span>${data['Summary']}</span>
		</summary>
	`;
	detail.style.display = verboseDetailsShown ? "block" : "none";
	detail.className = data['IsCritical'] ? "critical" : "";
	if (data['Details']) {
		detail.insertAdjacentHTML('beforeend', `
		<div class="details">
			<pre>
${JSON.stringify(JSON.parse(data['Details']), null, 2)}
			</pre>
		</div>`);
	}

	return detail;
}
