(function() {
	const agents = [];
	const seenAgents = new Set();
	const originalUrl = window.location.href;

	console.log('=== ENHANCED AGENT EXTRACTION ===');
	console.log('Page title:', document.title);
	console.log('Page URL:', originalUrl);

	// Utility function to wait for navigation and extract URL
	function waitForNavigation(timeout = 2000) {
		return new Promise((resolve) => {
			const startTime = Date.now();
			const checkNavigation = () => {
				if (window.location.href !== originalUrl) {
					resolve(window.location.href);
				} else if (Date.now() - startTime < timeout) {
					setTimeout(checkNavigation, 100);
				} else {
					resolve(null);
				}
			};
			checkNavigation();
		});
	}

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

				// Enhanced URL extraction strategies
				// Strategy 1: Look for traditional links
				const links = container.querySelectorAll('a');
				for (const link of links) {
					const href = link.href || '';
					if (href && (href.includes('/agents/') || href.includes('/agent/') ||
						href.includes('/subagent/') ||
						(link.textContent?.includes('View') && href.includes('subagents.sh')))) {
						url = href;
						console.log(`Found agent URL via link: ${url}`);
						break;
					}
				}

				// Strategy 2: Look for data attributes containing UUIDs
				if (!url) {
					const allElements = container.querySelectorAll('*');
					for (const element of allElements) {
						for (const attr of element.attributes) {
							if (attr.name.startsWith('data-') && attr.value) {
								// Check if value looks like a UUID
								const uuidPattern = /[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}/i;
								const match = attr.value.match(uuidPattern);
								if (match) {
									url = `https://subagents.sh/agents/${match[0]}`;
									console.log(`Found agent URL via data attribute ${attr.name}: ${url}`);
									break;
								}
							}
						}
						if (url) break;
					}
				}

				// Strategy 3: Look for onClick handlers or event data
				if (!url) {
					const clickableElements = container.querySelectorAll('[onclick], [data-href], [data-url], [data-id]');
					for (const element of clickableElements) {
						// Check onClick attribute
						const onclick = element.getAttribute('onclick');
						if (onclick && onclick.includes('agents')) {
							const uuidMatch = onclick.match(/[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}/i);
							if (uuidMatch) {
								url = `https://subagents.sh/agents/${uuidMatch[0]}`;
								console.log(`Found agent URL via onclick: ${url}`);
								break;
							}
						}

						// Check data-href, data-url, data-id attributes
						const dataAttrs = ['data-href', 'data-url', 'data-id'];
						for (const attr of dataAttrs) {
							const value = element.getAttribute(attr);
							if (value) {
								if (value.includes('/agents/')) {
									url = value.startsWith('http') ? value : `https://subagents.sh${value}`;
									console.log(`Found agent URL via ${attr}: ${url}`);
									break;
								}
								// Check if it's a UUID
								const uuidMatch = value.match(/[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}/i);
								if (uuidMatch) {
									url = `https://subagents.sh/agents/${uuidMatch[0]}`;
									console.log(`Found agent URL via ${attr} UUID: ${url}`);
									break;
								}
							}
						}
						if (url) break;
					}
				}

				// Strategy 4: Look for React props or embedded JSON data
				if (!url) {
					const scripts = container.querySelectorAll('script');
					for (const script of scripts) {
						try {
							const scriptContent = script.textContent || script.innerHTML;
							if (scriptContent) {
								// Look for UUID patterns in script content
								const uuidMatch = scriptContent.match(/[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}/i);
								if (uuidMatch) {
									url = `https://subagents.sh/agents/${uuidMatch[0]}`;
									console.log(`Found agent URL via script content: ${url}`);
									break;
								}
							}
						} catch (e) {
							// Ignore parsing errors
						}
					}
				}

				// Strategy 5: Click simulation (last resort) - Note: This should be used carefully in automation
				if (!url && index < 3) { // Limit to first 3 agents to avoid excessive navigation
					console.log(`Attempting click simulation for agent: ${name}`);
					try {
						// Store current URL
						const beforeClickUrl = window.location.href;

						// Find clickable area - could be the entire container or View button
						const clickTarget = viewButton.closest('[role="button"], button, a, [onclick]') || viewButton;

						// Simulate click
						clickTarget.click();

						// Wait a short time for potential navigation
						setTimeout(() => {
							if (window.location.href !== beforeClickUrl && window.location.href.includes('/agents/')) {
								url = window.location.href;
								console.log(`Found agent URL via click simulation: ${url}`);

								// Navigate back to original page
								window.history.back();
							}
						}, 500);
					} catch (e) {
						console.log(`Click simulation failed for ${name}:`, e.message);
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
	const agentsWithUrls = agents.filter(a => a.url).length;
	console.log(`Agents with URLs: ${agentsWithUrls}/${agents.length}`);

	if (agents.length > 0) {
		console.log('First agent:', agents[0]);
		if (agentsWithUrls > 0) {
			const firstWithUrl = agents.find(a => a.url);
			console.log('First agent with URL:', firstWithUrl);
		}
	}

	return {
		agents: agents,
		debug: {
			viewButtonsFound: viewButtons.length,
			agentsExtracted: agents.length,
			agentsWithUrls: agentsWithUrls,
			urlExtractionRate: agents.length > 0 ? (agentsWithUrls / agents.length * 100).toFixed(1) + '%' : '0%',
			bodyLength: document.body.textContent.length,
			hasAgentsFoundText: document.body.textContent.includes('Agents Found'),
			sampleAgentNames: agents.slice(0, 3).map(a => a.name),
			sampleUrls: agents.filter(a => a.url).slice(0, 3).map(a => a.url)
		}
	};
})();