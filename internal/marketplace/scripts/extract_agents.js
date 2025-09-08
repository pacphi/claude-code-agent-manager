(function() {
	const agents = [];
	const seenAgents = new Set();
	const originalUrl = window.location.href;

	console.log('=== OPTIMIZED AGENT EXTRACTION ===');
	console.log('Page title:', document.title);
	console.log('Page URL:', originalUrl);
	console.log('Ready state:', document.readyState);

	// OPTIMIZED: Minimal wait - assume pagination has already loaded all content
	console.log('Quick content verification...');
	let contentWait = 0;
	const maxWait = 2000; // Maximum 2 seconds - pagination should have loaded everything

	while (contentWait < maxWait) {
		const start = Date.now();
		while (Date.now() - start < 500) { /* busy wait 500ms intervals */ }
		contentWait += 500;

		// Quick check if we have agents loaded
		const hasContent = document.body.textContent.includes('Agents Found');
		const currentAgentCount = Array.from(document.querySelectorAll('button')).filter(btn =>
			btn.textContent && btn.textContent.trim().toLowerCase() === 'view'
		).length;

		console.log(`Quick check ${contentWait}ms: agents=${currentAgentCount}, hasContent=${hasContent}`);

		// If we have content and some agents, proceed immediately
		if (hasContent && currentAgentCount > 0) {
			console.log(`Content ready after ${contentWait}ms with ${currentAgentCount} agents`);
			break;
		}
	}

	// OPTIMIZED: Direct extraction without redundant methods
	console.log('Starting optimized extraction...');

	// Primary method: Find all View buttons - most reliable
	const viewButtons = Array.from(document.querySelectorAll('button')).filter(btn =>
		btn.textContent && btn.textContent.trim().toLowerCase() === 'view'
	);

	console.log('Found View buttons:', viewButtons.length);

	// Extract agents from View button containers
	viewButtons.forEach((viewButton, index) => {
		let container = viewButton;
		let depth = 0;

		// Walk up the DOM to find the agent card container
		while (container && depth < 15) {
			container = container.parentElement;
			depth++;

			if (!container) break;

			// Look for agent card structure
			const headings = container.querySelectorAll('h3');
			const paragraphs = container.querySelectorAll('p');
			const viewBtns = container.querySelectorAll('button');

			const hasHeading = headings.length > 0;
			const hasParagraph = paragraphs.length > 0;
			const hasViewBtn = Array.from(viewBtns).some(btn => btn.textContent?.trim().toLowerCase() === 'view');
			const containerText = container.textContent || '';
			const reasonableSize = containerText.length > 50 && containerText.length < 3000;

			if (hasHeading && hasViewBtn && reasonableSize) {
				// Extract agent data quickly
				let name = '';
				let description = '';
				let author = 'Unknown';
				let rating = 0.0;
				let url = '';

				// Get agent name
				for (const h3 of headings) {
					const headingText = h3.textContent?.trim() || '';
					if (headingText &&
						headingText.length > 2 &&
						headingText.length < 80 &&
						!headingText.includes('Filters') &&
						!headingText.includes('Agents Found')) {
						name = headingText;
						break;
					}
				}

				// Get description - first suitable paragraph
				for (const p of paragraphs) {
					const pText = p.textContent?.trim() || '';
					if (pText.length > 15 &&
						pText.length < 1000 &&
						pText !== name &&
						!pText.includes('View') &&
						!pText.match(/^[0-9.]+$/)) {
						description = pText;
						break;
					}
				}

				// Quick author detection
				if (containerText.includes('lst97')) {
					author = 'lst97';
				}

				// Quick URL extraction - try basic methods only
				const links = container.querySelectorAll('a');
				for (const link of links) {
					if (link.href && (link.href.includes('/agents/') || link.href.includes('/agent/'))) {
						url = link.href;
						break;
					}
				}

				// Add agent if valid and unique
				if (name && name.length > 2 && !seenAgents.has(name)) {
					seenAgents.add(name);
					agents.push({
						name: name,
						description: description || 'No description available',
						author: author,
						rating: rating,
						url: url
					});
				}

				break; // Found container, stop searching
			}
		}
	});

	console.log(`Agents extracted: ${agents.length}`);

	// Quick validation against expected count
	const pageText = document.body.textContent;
	const countMatch = pageText.match(/(\d+)\s+Agents?\s+Found/i);
	const expectedCount = countMatch ? parseInt(countMatch[1]) : null;

	if (expectedCount) {
		console.log(`Expected: ${expectedCount}, Extracted: ${agents.length}`);
		if (agents.length === expectedCount) {
			console.log('SUCCESS: All agents extracted correctly');
		} else if (agents.length < expectedCount) {
			console.log(`Missing ${expectedCount - agents.length} agents - this may indicate incomplete pagination`);
		}
	}

	// Debug output
	const agentsWithUrls = agents.filter(a => a.url).length;
	console.log(`Agents with URLs: ${agentsWithUrls}/${agents.length}`);
	console.log('Agent names:', agents.map(a => a.name).join(', '));

	return {
		agents: agents,
		debug: {
			viewButtonsFound: viewButtons.length,
			agentsExtracted: agents.length,
			expectedCount: expectedCount,
			countMatch: expectedCount === agents.length,
			agentsWithUrls: agentsWithUrls,
			urlExtractionRate: agents.length > 0 ? (agentsWithUrls / agents.length * 100).toFixed(1) + '%' : '0%',
			extractionTime: 'optimized',
			sampleAgentNames: agents.slice(0, 3).map(a => a.name)
		}
	};
})();