function getStorage(key) {
	switch(key) {
		case "verbose":
			return sessionStorage.getItem(key) === 'true';
		default:
			console.error("unknown variable name: " + key)
	}
}

function setStorage(key, val) {
	switch(key) {
		case "verbose":
			return sessionStorage.setItem('verbose', val === true)
		default:
			console.error("unknown variable name: " + key)
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
			if (data == null || !("length" in data)) {
				console.error("received invalid data: " + data);
				return
			}
			criticalDiv.innerHTML = "";
			for (var i = 0; i < data.length ; i++) {
				var notification = createNotification(data[i]);
				criticalDiv.prepend(notification);
			}
		})
		.catch(error => {
			console.error("failed to fetch '/critical': "+ error);
		});
}

document.addEventListener("DOMContentLoaded", function(){
	const eventSource = new EventSource("/notifications");
	window.addEventListener('beforeunload', () => {
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
			const notification = createNotification(data);
			verboseDiv.prepend(notification);
			setNotificationsCritical()
		} catch (error) {
			console.error('failed to handle message: ', error.message);
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

	detail.insertAdjacentHTML('beforeend', `
	<div class="details">
		<pre>
${data['Details']}
		</pre>
	</div>`);

	return detail;
}
