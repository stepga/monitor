function addText(boxId, text) {
	const box = document.getElementById(boxId);
	const div = document.createElement('div');
	div.className = 'text-box';
	div.textContent = text;
	box.prepend(div);
}

document.addEventListener("DOMContentLoaded", function(){
	const eventSrc = new EventSource("/events");
	eventSrc.onmessage = (event) => {
		// TODO: detect critical messages and post them into box-critical
		addText('box-verbose', event.data);
	};
});
