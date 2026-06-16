var verboseDetailsShown = true;


document.addEventListener("DOMContentLoaded", function(){
	const notificationSrc = new EventSource("/notifications");
	const notificationsElement = document.getElementById("notifications");
	const verboseCheckbox = document.getElementById("verbose");

	verboseCheckbox.addEventListener("change", function () {
		const detailsElements = document.querySelectorAll("details");
		detailsElements.forEach(detail => {
			if (!detail.children[0].classList.contains("important")) {
				verboseDetailsShown = this.checked;
				detail.style.display = this.checked ? "block" : "none";
			}
		})
	});

	notificationSrc.onmessage = (event) => {
		try {
			const data = JSON.parse(event.data);
			const notification = createNotification(data);
			notificationsElement.prepend(notification);
		} catch (error) {
			console.error('Failed to parse JSON in event:', error.message);
		}
	};
});

function createNotification(data) {
	summary = data['Summary']
	details = data['Details']
	important = data['IsImportant']

	var date = new Date();
	timestamp = date.toLocaleTimeString()

	const detail = document.createElement("details");
	detail.innerHTML = `
		<summary class="${details ? 'collapsable' : ''} ${important ? 'important' : '' }">
			<span class="pre"></span>
			<span>${timestamp}</span>
			<span>${summary}</span>
			<span class="post"></span>
		</summary>

	`;
	detail.style.display = verboseDetailsShown ? "block" : "none";
	if (details !== "") {
		detail.insertAdjacentHTML('beforeend', `
		<div class="details">
			${details}
		</div>`);
	}
	return detail;
}
