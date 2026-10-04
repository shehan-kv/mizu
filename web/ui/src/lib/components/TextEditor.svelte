<script lang="ts">
	import { debounce } from '$lib/utils/debounce';
	import { onMount } from 'svelte';
	import { aiState, suggest as generateSuggestion } from '$lib/ai';

	interface Props {
		value: string;
		disabled?: boolean;
		placeholder?: string;
		onSubmit?: () => void;
	}
	let {
		value = $bindable(),
		disabled = false,
		placeholder = 'Write your message here',
		onSubmit
	}: Props = $props();

	// Ghost text currently shown after the caret.
	// Only stores the missing portion of the suggestion.
	let suggestion = $state('');

	let textInput: HTMLDivElement | null = null;
	let suggestionSpan: HTMLSpanElement | null = null;
	let emptySpan: HTMLElement | null = null;

	// Suggestions are rendered as temporary DOM nodes.
	// Rebuild them instead of trying to keep them in sync.
	function removeSuggestionSpan() {
		suggestionSpan?.remove();
		emptySpan?.remove();
	}

	// Inserts accepted autocomplete text without moving the caret.
	function insertTextAtCaret(text: string) {
		const textNode = document.createTextNode(text);
		let selection = window.getSelection();
		let range = selection?.getRangeAt(0).cloneRange();
		range?.insertNode(textNode);

		const newRange = document.createRange();
		newRange.setStartAfter(textNode);
		newRange.setEndAfter(textNode);
		selection?.removeAllRanges();
		selection?.addRange(newRange);
	}

	function showSuggestionSpan() {
		if (!suggestion || !suggestionSpan || !emptySpan) return;

		suggestionSpan.textContent = suggestion;

		let selection = window.getSelection();
		let range = selection?.getRangeAt(0).cloneRange();

		range?.insertNode(suggestionSpan);
		range?.insertNode(emptySpan);

		const newRange = document.createRange();
		newRange.setStartBefore(emptySpan);
		newRange.setEndBefore(emptySpan);

		selection?.removeAllRanges();
		selection?.addRange(newRange);
	}

	function getTextBeforeCursor(): string {
		const selection = window.getSelection();

		if (!selection || !selection.rangeCount || !textInput) {
			return '';
		}

		const range = selection.getRangeAt(0);

		if (!textInput.contains(range.startContainer)) {
			return '';
		}

		const cursorRange = range.cloneRange();
		cursorRange.selectNodeContents(textInput);
		cursorRange.setEnd(range.startContainer, range.startOffset);

		const fragment = cursorRange.cloneContents();

		const temporary = document.createElement('div');
		temporary.appendChild(fragment);

		return (temporary.textContent ?? '').replace(/\u200B/g, '').replace(/\r\n?/g, '\n');
	}

	function getCurrentSentence(): string {
		const text = getTextBeforeCursor();

		if (!text) {
			return '';
		}

		const lastBoundary = Math.max(
			text.lastIndexOf('.'),
			text.lastIndexOf('!'),
			text.lastIndexOf('?'),
			text.lastIndexOf('\n')
		);

		return text.slice(lastBoundary + 1).trimStart();
	}

	let suggestionController: AbortController | null = null;

	async function fetchSuggestion() {
		removeSuggestionSpan();

		if (!aiState.enabled) {
			suggestion = '';
			return;
		}

		const sentence = getCurrentSentence();

		if (!sentence) {
			suggestion = '';
			return;
		}

		suggestionController?.abort();

		const controller = new AbortController();
		suggestionController = controller;

		const result = await generateSuggestion(sentence, controller.signal);

		if (controller.signal.aborted || suggestionController !== controller) {
			return;
		}

		suggestionController = null;
		suggestion = result ?? '';

		if (suggestion) {
			showSuggestionSpan();
		}
	}
	// Delay suggestion rendering to avoid fighting the caret
	// on every keystroke.
	const suggest = debounce(() => {
		fetchSuggestion();
	}, 250);

	// Convert the contenteditable DOM into plain text.
	//
	// This is the source of truth used by consumers.
	// Keep suggestion nodes out of the final value.
	function setValue() {
		let text = '';

		if (!textInput) return text;

		for (const node of textInput.childNodes) {
			if (node.nodeType === Node.TEXT_NODE) {
				text += node.nodeValue ?? '';
			} else if (node.nodeType === Node.ELEMENT_NODE) {
				const el = node as HTMLElement;

				if (el.classList.contains('suggestion-span')) continue;

				if (el.tagName === 'BR') {
					text += '\n';
				} else if (el.tagName === 'DIV' || el.tagName === 'P') {
					text += el.textContent?.trimEnd() + '\n';
				} else {
					text += el.textContent ?? '';
				}
			}
		}

		// Normalizing all line breaks
		text = text.replace(/\r\n?/g, '\n');
		value = text;
	}

	function clearEditor() {
		suggestionController?.abort();
		suggestionController = null;

		value = '';
		suggestion = '';

		if (textInput) {
			// eslint-disable-next-line svelte/no-dom-manipulating
			textInput.replaceChildren();
		}
	}

	function handleKeyDown(e: KeyboardEvent) {
		if (!textInput) return;

		if (e.key == 'Delete' && suggestionSpan && textInput.contains(suggestionSpan)) {
			e.preventDefault();
			suggestion = '';
		}
		if (e.key == 'Enter') {
			// Handle line breaks ourselves so browser-specific
			// contenteditable markup doesn't leak into the editor.

			if (onSubmit && !e.shiftKey) {
				e.preventDefault();

				removeSuggestionSpan();
				setValue();

				onSubmit();

				clearEditor();
				return;
			}

			e.preventDefault();

			const br = document.createElement('br');

			const sel = window.getSelection();
			if (!sel || !sel.rangeCount) return;

			const range = sel.getRangeAt(0);
			range.deleteContents();

			range.insertNode(br);

			const isAtEnd =
				!br.nextSibling ||
				(br.nextSibling.nodeType === Node.TEXT_NODE && br.nextSibling.nodeValue === '');

			let placeholder;

			// Trailing <br> nodes often need a placeholder text node
			// so the caret has somewhere valid to sit.
			if (isAtEnd) {
				placeholder = document.createTextNode('\u200B');
				br.parentNode?.insertBefore(placeholder, br.nextSibling);
			} else {
				placeholder = document.createTextNode('');
				br.parentNode?.insertBefore(placeholder, br.nextSibling);
			}

			const newRange = document.createRange();
			newRange.setStart(placeholder, 0);
			newRange.collapse(true);

			sel.removeAllRanges();
			sel.addRange(newRange);

			textInput.scrollTo({ top: textInput.scrollHeight, behavior: 'smooth' });

			suggestion = '';
		}

		if ((e.key == 'Backspace' || e.key == 'Escape') && suggestion) {
			suggest.stop();

			suggestionController?.abort();
			suggestionController = null;

			suggestion = '';

			// Accept the current suggestion.
		} else if (e.key == 'Tab' && suggestion) {
			e.preventDefault();

			suggest.stop();

			suggestionController?.abort();
			suggestionController = null;

			insertTextAtCaret(suggestion);
			suggestion = '';
		}

		removeSuggestionSpan();
		setValue();
	}

	// Remove temporary zero-width placeholders that exist
	// purely to keep caret positioning working after Enter.
	function normalizeText() {
		const sel = window.getSelection();
		if (!sel || !sel.rangeCount) return;

		const range = sel.getRangeAt(0);
		const node = range.startContainer;

		if (node.nodeType !== Node.TEXT_NODE) return;

		let value = node.nodeValue || '';
		const offset = range.startOffset;

		if (value.length > 1 && value.includes('\u200B')) {
			const newOffset = offset - (value.slice(0, offset).includes('\u200B') ? 1 : 0);

			value = value.replace('\u200B', '');
			node.nodeValue = value;

			const newRange = document.createRange();
			newRange.setStart(node, Math.max(0, newOffset));
			newRange.collapse(true);

			sel.removeAllRanges();
			sel.addRange(newRange);
		}
	}

	function handleInput() {
		if (!textInput) return;

		normalizeText();

		setValue();

		if (aiState.enabled) {
			suggest();
		}
	}

	// Default contenteditable paste brings HTML with it.
	// Force plain text so the editor DOM stays predictable.
	function handlePaste(e: ClipboardEvent) {
		if (!textInput) return;

		e.preventDefault();

		const text = e.clipboardData?.getData('text/plain') ?? '';
		if (!text) return;

		const lines = text.split(/\r?\n/);

		for (let i = 0; i < lines.length; i++) {
			insertTextAtCaret(lines[i]);
			if (i < lines.length - 1) {
				const br = document.createElement('br');
				const selection = window.getSelection();
				if (!selection || selection.rangeCount === 0) continue;
				const range = selection.getRangeAt(0);
				range.insertNode(br);
				range.setStartAfter(br);
				range.setEndAfter(br);
				selection.removeAllRanges();
				selection.addRange(range);
			}
		}

		setValue();
	}

	onMount(() => {
		// Reused ghost-text node shown for autocomplete.

		suggestionSpan = document.createElement('span');
		suggestionSpan.setAttribute('contentEditable', 'false');
		suggestionSpan.classList.add('text-neutral-500');
		suggestionSpan.classList.add('suggestion-span');
		suggestionSpan.style.pointerEvents = 'none';
		suggestionSpan.style.userSelect = 'none';

		emptySpan = document.createElement('span');

		return () => {
			suggest.stop();
			suggestionController?.abort();
		};
	});
</script>

<div
	class="h-full w-full overflow-y-auto outline-none
		empty:before:pointer-events-none empty:before:text-sm
		empty:before:text-neutral-500 empty:before:content-[attr(placeholder)]"
	role="textbox"
	tabindex="0"
	contenteditable={!disabled}
	{placeholder}
	oninput={handleInput}
	onkeydown={handleKeyDown}
	onpaste={handlePaste}
	bind:this={textInput}
></div>
