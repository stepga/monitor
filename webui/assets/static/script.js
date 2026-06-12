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
					date.toLocaleTimeString()
				)
			);
		} catch (error) {
			console.error('Failed to parse JSON in event:', error.message);
		}
			};
});

function createNotification(type, title, details, timestamp) {
    const article = document.createElement("article");
    article.className = `notification ${type}`;

    article.innerHTML = `
        <div class="meta">
            <span>${type.toUpperCase()}</span>
            <time>${timestamp}</time>
        </div>

        <h3>${title}</h3>

        <p>${details}</p>
    `;

    return article;
}
