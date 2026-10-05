<script lang="ts">
	import { currencyDecimals } from '$lib/utils/currencyDecimals';
	import Decimal from 'decimal.js';
	import { onMount } from 'svelte';

	interface Props {
		value: Decimal;
		placeholder?: string;
		name?: string;
		id?: string;
		required?: boolean;
		defaultValue?: string;
		currency?: string;
		class?: string;
	}
	let {
		value = $bindable(),
		placeholder,
		name,
		id,
		required,
		defaultValue,
		currency,
		class: className = ''
	}: Props = $props();

	let digits = currency ? currencyDecimals[currency] : 3;

	const formatter = Intl.NumberFormat(undefined, {
		minimumFractionDigits: digits,
		maximumFractionDigits: digits
	});
	function getSeparators() {
		let parts = formatter.formatToParts(1000);

		let groupSeparator = '';
		let decimalSeparator = '';

		for (let part of parts) {
			if (groupSeparator.length > 0 && decimalSeparator.length > 0) {
				break;
			}

			if (part.type == 'group') {
				groupSeparator = part.value;
			}

			if (part.type == 'decimal') {
				decimalSeparator = part.value;
			}
		}

		return { group: groupSeparator, decimal: decimalSeparator };
	}

	const { group: GROUP_SEPARATOR, decimal: DECIMAL_SEPARATOR } = getSeparators();

	let inputRef: HTMLInputElement;

	function handleStripSeparator(text: string) {
		let numericText = text.split(GROUP_SEPARATOR).join('');
		if (DECIMAL_SEPARATOR && DECIMAL_SEPARATOR !== '.') {
			numericText = numericText.replace(DECIMAL_SEPARATOR, '.');
		}

		return numericText;
	}
	function handleFormatDisplay(text: string) {
		let numericText = handleStripSeparator(text);
		return formatter.format(numericText as Intl.StringNumericLiteral);
	}

	const allowedKeys = [
		'Backspace',
		'ArrowLeft',
		'ArrowRight',
		'Delete',
		GROUP_SEPARATOR,
		DECIMAL_SEPARATOR
	];

	function handleKeyDown(e: KeyboardEvent) {
		e.preventDefault();

		if (!/[0-9]/.test(e.key) && !allowedKeys.includes(e.key)) return;

		const selStart = inputRef.selectionStart ?? 0;
		const selEnd = inputRef.selectionEnd ?? 0;
		let displayValue = inputRef.value;

		if (e.key === 'Backspace') {
			if (selStart === selEnd && selStart > 0) {
				displayValue = displayValue.slice(0, selStart - 1) + displayValue.slice(selEnd);
				inputRef.value = displayValue;

				let strippedDecimal = handleStripSeparator(inputRef.value);
				value = new Decimal(strippedDecimal != '' ? strippedDecimal : 0.0);

				inputRef.setSelectionRange(selStart - 1, selStart - 1);
			}
			return;
		}

		if (e.key === 'Delete') {
			displayValue = displayValue.slice(0, selStart) + displayValue.slice(selEnd + 1);
			inputRef.value = displayValue;
			inputRef.setSelectionRange(selStart, selStart);

			let strippedDecimal = handleStripSeparator(inputRef.value);
			value = new Decimal(strippedDecimal != '' ? strippedDecimal : 0.0);

			return;
		}

		if (e.key === 'ArrowLeft') {
			inputRef.setSelectionRange(selStart - 1, selStart - 1);
			return;
		}

		if (e.key === 'ArrowRight') {
			inputRef.setSelectionRange(selStart + 1, selStart + 1);
			return;
		}

		if (e.key === DECIMAL_SEPARATOR && displayValue.includes(DECIMAL_SEPARATOR)) return;

		let afterInsert = displayValue.slice(0, selStart) + e.key + displayValue.slice(selEnd);

		const formatted = handleFormatDisplay(afterInsert);

		inputRef.value = formatted;

		let strippedDecimal = handleStripSeparator(inputRef.value);
		value = new Decimal(strippedDecimal != '' ? strippedDecimal : 0.0);

		const commasBeforeOld = displayValue.slice(0, selStart).split(GROUP_SEPARATOR).length - 1;
		const commasBeforeNew = formatted.slice(0, selStart + 1).split(GROUP_SEPARATOR).length - 1;

		const commaDiff = commasBeforeNew - commasBeforeOld;

		const newPos = Math.min(formatted.length, selStart + 1 + commaDiff);

		inputRef.setSelectionRange(newPos, newPos);
	}

	function handleOnBlur() {
		inputRef.value = handleFormatDisplay(inputRef.value);
	}

	export function clear() {
		inputRef.value = handleFormatDisplay('0.00');
		value = new Decimal(inputRef.value);
	}

	onMount(() => {
		if (inputRef) {
			let strippedDecimal = handleStripSeparator(inputRef.value);
			value = new Decimal(strippedDecimal != '' ? strippedDecimal : 0.0);
		}
	});
</script>

<input
	bind:this={inputRef}
	defaultValue={formatter.format(defaultValue as Intl.StringNumericLiteral)}
	{required}
	onkeydown={handleKeyDown}
	onblur={handleOnBlur}
	{placeholder}
	type="string"
	inputmode="decimal"
	{name}
	{id}
	class={`rounded bg-white px-4 py-2 [appearance:textfield] placeholder:text-xs placeholder:italic
    dark:bg-neutral-950 [&::-webkit-inner-spin-button]:appearance-none
    [&::-webkit-outer-spin-button]:appearance-none ${className}`}
/>
