<script lang="ts">
	import { phrases } from '$lib/components/message/mockData';
	import { debounce } from '$lib/utils/debounce';
	import { onMount } from 'svelte';

	interface Props {
		disabled?: boolean;
		autoSuggest?: boolean;
		placeholder?: string;
		onSubmit: (text: string) => any;
	}
	let {
		disabled = false,
		autoSuggest = false,
		placeholder = 'Write your message here',
		onSubmit
	}: Props = $props();

	let suggestion = $state('');

	let textInput: HTMLDivElement | null = null;
	let suggestionSpan: HTMLSpanElement | null = null;
	let emptySpan: HTMLElement | null = null;

	export function submit() {
		if (!textInput) return;

		onSubmit(textInput.innerText || '');
		textInput.innerText = '';
	}

	function removeSuggestionSpan() {
		suggestionSpan?.remove();
		emptySpan?.remove();
	}

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

	function getTextBeforeCursor() {
		const selection = window.getSelection();
		if (!selection || !selection.anchorNode) return;

		const text = selection.anchorNode?.textContent ?? '';

		if (!text) return;

		const delimiters = ['.', ','];
		const lastIndices = delimiters.map((d) => text.lastIndexOf(d));
		const lastDelimiterIndex = Math.max(...lastIndices);

		return lastDelimiterIndex === -1
			? text.trimStart()
			: text.substring(lastDelimiterIndex + 1).trimStart();
	}

	function fetchSuggestion() {
		let sentence = getTextBeforeCursor();

		if (!sentence) return;

		let _suggested: string[] = [];
		if (sentence) {
			_suggested = phrases.filter((phrase) => phrase.startsWith(sentence));
		}

		if (_suggested.length > 0) {
			suggestion = _suggested[0].substring(sentence.length);
		} else {
			suggestion = '';
		}
	}

	const suggest = debounce(() => {
		fetchSuggestion();
		showSuggestionSpan();
	}, 250);

	function handleKeyDown(e: KeyboardEvent) {
		if (!textInput) return;

		if (e.key == 'Delete' && suggestionSpan && textInput.contains(suggestionSpan)) {
			e.preventDefault();
			removeSuggestionSpan();
			suggestion = '';
			return;
		}
		if (!e.shiftKey && e.key == 'Enter') {
			e.preventDefault();
			removeSuggestionSpan();
			onSubmit(textInput.innerText);
			textInput.innerText = '';
		}

		if ((e.key == 'Backspace' || e.key == 'Escape') && suggestion) {
			removeSuggestionSpan();
			suggestion = '';
		} else if (e.key == 'Tab' && suggestion) {
			e.preventDefault();
			insertTextAtCaret(suggestion);
			removeSuggestionSpan();
			suggestion = '';
		}

		removeSuggestionSpan();
	}

	function handleInput() {
		if (textInput && autoSuggest) {
			suggest();
		}
	}

	onMount(() => {
		suggestionSpan = document.createElement('span');
		suggestionSpan.setAttribute('contentEditable', 'false');
		suggestionSpan.classList.add('text-neutral-500');
		suggestionSpan.style.pointerEvents = 'none';
		suggestionSpan.style.userSelect = 'none';

		emptySpan = document.createElement('span');

		return () => {
			suggest.stop();
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
	bind:this={textInput}
></div>
