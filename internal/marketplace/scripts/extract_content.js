(function() {
	console.log('=== ENHANCED CONTENT EXTRACTION ===');
	console.log('Page URL:', window.location.href);
	console.log('Page title:', document.title);

	// Utility function to check if content indicators are present
	function hasContentLoaded() {
		const hasCopyButton = document.querySelector('button')?.textContent?.includes('Copy Content');
		const hasYamlContent = document.body.textContent.includes('---') &&
							   document.body.textContent.includes('name:') &&
							   document.body.textContent.includes('description:');
		return hasCopyButton || hasYamlContent;
	}

	// Function to extract content using all strategies
	function extractContent() {
		console.log('Starting content extraction...');

		// Strategy 1: Find text content that looks like an agent definition
		console.log('Strategy 1: Searching all elements for YAML content...');
		const allElements = document.querySelectorAll('div, pre, code, p, span, section, article');
		for (const element of allElements) {
			const text = element.textContent.trim();
			// Check if this is an agent definition (starts with YAML frontmatter)
			if (text.startsWith('---') && text.includes('name:') && text.includes('description:')) {
				// Verify it's not just a snippet but the full content
				if (text.length > 200 && (text.includes('tools:') || text.includes('model:') || text.includes('system:'))) {
					console.log('Strategy 1: Found agent content via element text');
					return text;
				}
			}
		}

		// Strategy 2: Look for elements near the "Copy Content" button
		console.log('Strategy 2: Searching near Copy Content button...');
		const copyButtons = Array.from(document.querySelectorAll('button, [role="button"]')).filter(btn =>
			btn.textContent && (btn.textContent.includes('Copy Content') || btn.textContent.includes('Copy') || btn.textContent.includes('Download'))
		);

		for (const copyButton of copyButtons) {
			console.log('Found copy-like button:', copyButton.textContent.trim());
			// Look for content in nearby containers
			let parent = copyButton.parentElement;
			let depth = 0;
			while (parent && parent !== document.body && depth < 10) {
				// Check all text nodes in this container
				const walker = document.createTreeWalker(
					parent,
					NodeFilter.SHOW_TEXT,
					null,
					false
				);

				let textContent = '';
				let node;
				while (node = walker.nextNode()) {
					textContent += node.textContent;
				}

				textContent = textContent.trim();
				if (textContent.startsWith('---') && textContent.includes('name:') && textContent.length > 200) {
					console.log('Strategy 2: Found agent content near copy button');
					return textContent;
				}

				parent = parent.parentElement;
				depth++;
			}
		}

		// Strategy 3: Look for pre/code blocks and content containers
		console.log('Strategy 3: Searching code blocks and content containers...');
		const contentSelectors = [
			'pre', 'code', '.hljs', '.language-yaml', '.language-markdown',
			'[class*="content"]', '[class*="markdown"]', '[class*="code"]',
			'[id*="content"]', '[id*="agent"]', '[data-content]'
		];

		for (const selector of contentSelectors) {
			const elements = document.querySelectorAll(selector);
			for (const block of elements) {
				const text = block.textContent.trim();
				if (text.startsWith('---') && text.includes('name:') && text.length > 200) {
					console.log(`Strategy 3: Found agent content in ${selector}`);
					return text;
				}
			}
		}

		// Strategy 4: Find by content structure - look for the main content area
		console.log('Strategy 4: Searching main content areas...');
		const mainSelectors = ['main', '[role="main"]', '.main-content', '#main', '.content', '.page-content'];
		for (const selector of mainSelectors) {
			const mainContent = document.querySelector(selector);
			if (mainContent) {
				// Get all text content and look for the agent definition
				const allText = mainContent.textContent;
				// Use regex to find the agent definition block
				const patterns = [
					/---[\s\S]*?name:[\s\S]*?description:[\s\S]*?---[\s\S]*/,
					/^---\n[\s\S]*?\n---\n[\s\S]*/m,
					/^---\r?\n[\s\S]*?\r?\n---\r?\n[\s\S]*/m
				];

				for (const pattern of patterns) {
					const match = allText.match(pattern);
					if (match && match[0].length > 200) {
						console.log(`Strategy 4: Found agent content in ${selector} using regex`);
						return match[0].trim();
					}
				}
			}
		}

		// Strategy 5: Look in shadow DOM and iframes
		console.log('Strategy 5: Searching shadow DOM and iframes...');
		const elementsWithShadow = document.querySelectorAll('*');
		for (const element of elementsWithShadow) {
			if (element.shadowRoot) {
				const shadowContent = element.shadowRoot.textContent;
				if (shadowContent && shadowContent.includes('---') && shadowContent.includes('name:')) {
					const match = shadowContent.match(/---[\s\S]*?name:[\s\S]*?description:[\s\S]*/);
					if (match && match[0].length > 200) {
						console.log('Strategy 5: Found agent content in shadow DOM');
						return match[0].trim();
					}
				}
			}
		}

		// Strategy 6: Search in script tags for JSON or embedded content
		console.log('Strategy 6: Searching script tags for embedded content...');
		const scripts = document.querySelectorAll('script');
		for (const script of scripts) {
			try {
				const scriptContent = script.textContent || script.innerHTML;
				if (scriptContent) {
					// Look for YAML content embedded in JSON or strings
					const yamlMatch = scriptContent.match(/["']---[\s\S]*?name:[\s\S]*?description:[\s\S]*?["']/);
					if (yamlMatch) {
						const content = yamlMatch[0].slice(1, -1); // Remove quotes
						if (content.length > 200) {
							console.log('Strategy 6: Found agent content in script tag');
							return content;
						}
					}
				}
			} catch (e) {
				// Ignore parsing errors
			}
		}

		console.log('No agent content found using any strategy');
		return '';
	}

	// Synchronous execution with retry logic for browser automation
	function main() {
		const maxRetries = 3;
		const retryDelays = [500, 1000, 2000]; // Shorter delays for synchronous execution

		for (let attempt = 0; attempt < maxRetries; attempt++) {
			console.log(`Content extraction attempt ${attempt + 1}/${maxRetries}`);

			// Synchronous wait for content to load
			const startTime = Date.now();
			const maxWait = 3000; // 3 seconds max wait

			while (Date.now() - startTime < maxWait) {
				const hasCopyButton = document.querySelector('button')?.textContent?.includes('Copy Content');
				const hasYamlContent = document.body.textContent.includes('---') &&
									   document.body.textContent.includes('name:') &&
									   document.body.textContent.includes('description:');

				if (hasCopyButton || hasYamlContent) {
					console.log(`Content loaded after ${Date.now() - startTime}ms`);
					break;
				}

				// Busy wait for a short time
				const waitStart = Date.now();
				while (Date.now() - waitStart < 100) {
					// Short busy wait
				}
			}

			// Try to extract content
			const content = extractContent();
			if (content) {
				console.log(`Success! Content extracted on attempt ${attempt + 1}`);
				console.log(`Content length: ${content.length} characters`);
				return content;
			}

			// If not the last attempt, wait before retrying
			if (attempt < maxRetries - 1) {
				console.log(`Attempt ${attempt + 1} failed, retrying in ${retryDelays[attempt]}ms...`);
				const waitStart = Date.now();
				while (Date.now() - waitStart < retryDelays[attempt]) {
					// Synchronous wait
				}
			}
		}

		console.log('All extraction attempts failed');
		return '';
	}

	// Return the main function result
	return main();
})();