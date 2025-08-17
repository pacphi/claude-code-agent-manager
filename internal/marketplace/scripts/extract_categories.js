(function() {
	console.log('=== CATEGORY EXTRACTION SCRIPT RUNNING ===');
	console.log('Page URL:', window.location.href);
	console.log('Page title:', document.title);
	console.log('Page ready state:', document.readyState);

	// Wait for the page to be fully loaded
	if (document.readyState !== 'complete') {
		console.log('Page not fully loaded, waiting...');
		let waited = 0;
		while (document.readyState !== 'complete' && waited < 5000) {
			const start = Date.now();
			while (Date.now() - start < 100) { /* busy wait */ }
			waited += 100;
		}
	}

	console.log('Page loaded, ready state:', document.readyState);

	// Wait longer for React/Next.js content to render
	console.log('Waiting for dynamic content to load...');
	let maxAttempts = 25; // 5 seconds total
	let attempt = 0;
	let categoryLinks = [];

	while (attempt < maxAttempts) {
		console.log(`Attempt ${attempt + 1}/${maxAttempts}`);

		// Get all links and debug what we find
		const allLinks = Array.from(document.querySelectorAll('a'));
		console.log(`Total links found: ${allLinks.length}`);

		// Log first few links for debugging
		if (allLinks.length > 0) {
			console.log('First 5 link hrefs:', allLinks.slice(0, 5).map(a => a.href));
			console.log('First 5 link texts:', allLinks.slice(0, 5).map(a => a.textContent?.trim()));
		}

		// Look for category links with different patterns
		const categoryPattern1 = allLinks.filter(a => a.href && a.href.includes('/categories/') && !a.href.endsWith('/categories'));
		const categoryPattern2 = allLinks.filter(a => a.href && a.href.match(/\/categories\/[a-z-]+$/));
		const categoryPattern3 = allLinks.filter(a => a.textContent && a.textContent.includes('agents') && a.href && a.href.includes('/categories/'));

		console.log(`Category pattern 1 (href contains /categories/): ${categoryPattern1.length}`);
		console.log(`Category pattern 2 (href matches /categories/slug): ${categoryPattern2.length}`);
		console.log(`Category pattern 3 (text contains 'agents'): ${categoryPattern3.length}`);

		if (categoryPattern1.length > 0) {
			console.log('Found category links with pattern 1');
			categoryLinks = categoryPattern1;
			break;
		}

		if (categoryPattern2.length > 0) {
			console.log('Found category links with pattern 2');
			categoryLinks = categoryPattern2;
			break;
		}

		if (categoryPattern3.length > 0) {
			console.log('Found category links with pattern 3');
			categoryLinks = categoryPattern3;
			break;
		}

		// Wait 200ms before next attempt
		const start = Date.now();
		while (Date.now() - start < 200) { /* busy wait */ }
		attempt++;
	}

	console.log(`Final category links found: ${categoryLinks.length}`);

	if (categoryLinks.length === 0) {
		// Enhanced debugging
		console.log('No category links found. Enhanced debugging:');
		console.log('Document HTML length:', document.documentElement.innerHTML.length);
		console.log('Body HTML sample:', document.body.innerHTML.substring(0, 1000));

		// Look for any element that might contain category information
		const possibleCategoryElements = document.querySelectorAll('[class*="category"], [class*="card"], [data-*="category"]');
		console.log('Possible category elements found:', possibleCategoryElements.length);

		return {
			categories: [],
			debug: {
				totalLinks: allLinks.length,
				pageUrl: window.location.href,
				readyState: document.readyState,
				bodyLength: document.body.innerHTML.length,
				possibleElements: possibleCategoryElements.length,
				sampleHtml: document.body.innerHTML.substring(0, 500)
			}
		};
	}

	// Extract category information
	const categories = [];
	console.log('Processing category links...');

	categoryLinks.forEach((link, index) => {
		console.log(`Processing link ${index + 1}:`, link.href);

		const href = link.href;
		if (!href || !href.includes('/categories/')) {
			console.log(`Skipping link ${index + 1}: no valid href`);
			return;
		}

		// Extract slug from URL
		let slug;
		if (href.includes('https://subagents.sh/categories/')) {
			slug = href.split('https://subagents.sh/categories/')[1];
		} else if (href.includes('/categories/')) {
			slug = href.split('/categories/')[1];
		}

		if (!slug || slug.includes('?') || slug.includes('#') || slug.length === 0) {
			console.log(`Skipping link ${index + 1}: invalid slug '${slug}'`);
			return;
		}

		console.log(`Link ${index + 1} slug: '${slug}'`);

		// Get text content and parse it
		const linkText = link.textContent?.trim() || '';
		console.log(`Link ${index + 1} text: '${linkText}'`);

		if (linkText.length < 3) {
			console.log(`Skipping link ${index + 1}: text too short`);
			return;
		}

		// The link text often comes as one big string like "AI & Machine Learning10 agentsLeverage artificial..."
		// We need to parse this more carefully

		console.log(`Link ${index + 1} full text: '${linkText}'`);

		// Try to extract name, agent count, and description from the combined text
		let name = '';
		let agentCount = 0;
		let description = '';

		// First, try to find the agent count pattern in the text
		const agentCountMatch = linkText.match(/(\d+)\s+agents?/i);
		if (agentCountMatch) {
			agentCount = parseInt(agentCountMatch[1]);
			console.log(`Found agent count: ${agentCount}`);

			// Split the text at the agent count pattern
			const beforeAgentCount = linkText.substring(0, agentCountMatch.index);
			const afterAgentCount = linkText.substring(agentCountMatch.index + agentCountMatch[0].length);

			// The name should be everything before the agent count
			name = beforeAgentCount.trim();

			// The description should be everything after the agent count
			description = afterAgentCount.trim();
		} else {
			// No agent count found - the whole text might be the name
			// Try to split by common patterns
			const lines = linkText.split('\n').map(line => line.trim()).filter(line => line.length > 0);

			if (lines.length > 0) {
				name = lines[0];
				if (lines.length > 1) {
					description = lines.slice(1).join(' ').trim();
				}
			} else {
				name = linkText;
			}
		}

		console.log(`Parsed - Name: '${name}', Agent Count: ${agentCount}, Description: '${description}'`);

		// Clean up the name - remove any trailing numbers or "agents" text
		name = name.replace(/\d+\s*agents?\s*$/i, '').trim();

		// Clean up description - remove leading agent count if it leaked in
		description = description.replace(/^\d+\s*agents?\s*/i, '').trim();

		// Create category object
		const category = {
			id: slug,
			name: name,
			slug: slug,
			description: description || `${name} tools and utilities`,
			url: href,
			agentCount: agentCount
		};

		console.log(`Created category:`, category);
		categories.push(category);
	});

	console.log(`Total categories extracted: ${categories.length}`);

	// Sort alphabetically by name
	categories.sort((a, b) => a.name.toLowerCase().localeCompare(b.name.toLowerCase()));

	console.log('Categories sorted alphabetically');

	return {
		categories: categories,
		debug: {
			totalLinksFound: categoryLinks.length,
			categoriesExtracted: categories.length,
			source: 'dynamic'
		}
	};
})();