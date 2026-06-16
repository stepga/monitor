var verboseDetailsShown = true;

function setNotificationsSticky() {
	const notificationsSticky = document.getElementById("notifications_sticky");
	fetch("/sticky")
		.then(response => response.json())
		.then(data => {
			notificationsSticky.innerHTML = "";
			for (var i = 0; i < data.length ; i++) {
				notification = createNotification(data[i]);
				notificationsSticky.prepend(notification);
			}
		})
		.catch(error => {
			console.error('failed to fetch sticky: ' + error);
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
			// enforce `data['IsImportant'] = false` for notifications in the verbose log
			data['IsImportant'] = false;
			const notification = createNotification(data);
			notificationsVerbose.prepend(notification);
			setNotificationsSticky()
		} catch (error) {
			console.error('Failed to parse JSON in event:', error.message);
		}
	};

	setNotificationsSticky();
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
	detail.className = data['IsImportant'] ? "important" : "";
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
