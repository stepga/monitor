document.addEventListener("DOMContentLoaded", function(){
	const eventSrc = new EventSource("/events");
	const feed = document.getElementById("notifications");

	eventSrc.onmessage = (event) => {
		var date = new Date();
		try {
			const data = JSON.parse(event.data);
			feed.prepend(
				createNotification(
					data['SubSystemName'],
					data['Summary'],
					data['Report'],
					date.toLocaleTimeString(),
					data['IsCritical']
				)
			);
		} catch (error) {
			console.error('Failed to parse JSON in event:', error.message);
		}
	};
});

function createNotification(subsystem, summary, report, timestamp, critical) {
	const detail = document.createElement("details");
	detail.innerHTML = `
		<summary subsystem="${subsystem}" class="${report ? 'collapsable' : ''} ${critical ? 'critical' : '' }">
			<span class="pre"></span>
			<span>${timestamp}</span>
			<span>${summary}</span>
			<span class="post"></span>
		</summary>

	`;
	if (report !== "") {
		detail.insertAdjacentHTML('beforeend', `
		<div class="content">
			${report}
		</div>`);
	}
	return detail;
}
