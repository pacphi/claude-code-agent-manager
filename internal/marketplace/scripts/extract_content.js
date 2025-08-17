(function() {
	// Look for common content containers
	const selectors = [
		'pre', 'code', '.content', '.agent-content',
		'[data-content]', '.markdown-body', '.prose',
		'main', '[role="main"]', '.main-content'
	];

	for (const selector of selectors) {
		const element = document.querySelector(selector);
		if (element && element.textContent.trim().length > 50) {
			return element.textContent.trim();
		}
	}

	// Fallback to body content
	return document.body.textContent.trim();
})();