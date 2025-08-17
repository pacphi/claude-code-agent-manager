(function() {
	const agents = [];
	const seenAgents = new Set();

	console.log('=== NEW AGENT EXTRACTION ===');
	console.log('Page title:', document.title);
	console.log('Page URL:', window.location.href);

	// Look for the main content area that contains the agent grid
	// Based on the snapshot, agents are in a grid container
	const agentContainers = [];

	// Find all View buttons first - these are unique to agent cards
	const viewButtons = Array.from(document.querySelectorAll('button')).filter(btn =>
		btn.textContent && btn.textContent.trim() === 'View'
	);

	console.log('Found View buttons:', viewButtons.length);

	// For each View button, find its agent card container
	viewButtons.forEach((viewButton, index) => {
		let container = viewButton;
		let depth = 0;

		// Walk up the DOM to find the agent card container
		while (container && depth < 15) {
			container = container.parentElement;
			depth++;

			if (!container) break;

			// Look for a container that has both h3 heading and paragraph description
			const headings = container.querySelectorAll('h3');
			const paragraphs = container.querySelectorAll('p');
			const viewBtns = container.querySelectorAll('button');

			// Check if this container has the right structure for an agent card
			const hasHeading = headings.length > 0;
			const hasParagraph = paragraphs.length > 0;
			const hasViewBtn = Array.from(viewBtns).some(btn => btn.textContent?.trim() === 'View');
			const containerText = container.textContent || '';

			// Additional checks to ensure it's an agent card
			const hasRating = containerText.includes('0.0');
			const hasCategory = containerText.includes('AI & Machine Learning');
			const reasonableSize = containerText.length > 100 && containerText.length < 2000;

			if (hasHeading && hasParagraph && hasViewBtn && hasRating && reasonableSize) {
				console.log(`Found agent card container at depth ${depth} for button ${index}`);

				// Extract agent data from this container
				let name = '';
				let description = '';
				let author = '';
				let rating = 0.0;
				let url = '';

				// Get the agent name from h3 heading
				for (const h3 of headings) {
					const headingText = h3.textContent?.trim() || '';
					// Skip non-agent headings
					if (headingText &&
						headingText !== 'Filters' &&
						headingText !== 'Quick Filters' &&
						headingText !== 'Product' &&
						headingText !== 'Legal' &&
						!headingText.includes('Agents Found') &&
						headingText.length > 2 &&
						headingText.length < 50) {
						name = headingText;
						console.log(`Found agent name: ${name}`);
						break;
					}
				}

				// Get the description from paragraph
				for (const p of paragraphs) {
					const pText = p.textContent?.trim() || '';
					// Look for a substantial description that's not UI text
					if (pText.length > 30 &&
						pText.length < 1000 &&
						pText !== name &&
						!pText.includes('subagent') &&
						!pText.includes('claude-code') &&
						!pText.includes('View') &&
						!pText.includes('lst97') &&
						!pText.includes('0.0') &&
						!pText.includes('AI & Machine Learning') &&
						!pText.includes('Discover and share')) {
						description = pText;
						console.log(`Found description: ${description.substring(0, 100)}...`);
						break;
					}
				}

				// Extract other metadata
				if (containerText.includes('lst97')) {
					author = 'lst97';
				}

				if (containerText.includes('0.0')) {
					rating = 0.0;
				}

				// Look for links that might be the agent URL
				const links = container.querySelectorAll('a');
				for (const link of links) {
					const href = link.href || '';
					if (href && (href.includes('/agent/') || href.includes('/subagent/') ||
						(link.textContent?.includes('View') && href.includes('subagents.sh')))) {
						url = href;
						break;
					}
				}

				// Add the agent if we have valid data and haven't seen it before
				if (name && name.length > 2 && !seenAgents.has(name.toLowerCase())) {
					seenAgents.add(name.toLowerCase());
					console.log(`Adding agent: ${name}`);

					agents.push({
						name: name,
						description: description || 'No description available',
						author: author || 'Unknown',
						rating: rating,
						url: url
					});
				}

				// Found the container for this view button, stop searching
				break;
			}
		}
	});

	console.log(`Total agents extracted: ${agents.length}`);
	if (agents.length > 0) {
		console.log('First agent:', agents[0]);
	}

	return {
		agents: agents,
		debug: {
			viewButtonsFound: viewButtons.length,
			agentsExtracted: agents.length,
			bodyLength: document.body.textContent.length,
			hasAgentsFoundText: document.body.textContent.includes('Agents Found'),
			sampleAgentNames: agents.slice(0, 3).map(a => a.name)
		}
	};
})();