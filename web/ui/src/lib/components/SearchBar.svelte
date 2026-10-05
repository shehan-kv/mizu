<script lang="ts">
	import { debounce } from '$lib/utils/debounce';
	import MagnifyingGlass from 'phosphor-svelte/lib/MagnifyingGlass';

	interface Props {
		value: string;
		delay?: number;
		onchange?: () => unknown;
	}
	let { value = $bindable(), onchange, delay = 300 }: Props = $props();

	let inputValue = $state('');

	const updateKeywordDebounced = debounce(() => {
		value = inputValue;
		if (onchange) {
			onchange();
		}
	}, delay);

	function onInput() {
		updateKeywordDebounced();
	}
</script>

<div class="relative w-full rounded bg-neutral-100 dark:bg-neutral-900">
	<input
		type="search"
		name="search"
		id="search"
		bind:value={inputValue}
		oninput={onInput}
		class="w-full py-1.5 pr-10 pl-2 text-sm outline-none"
	/>
	<MagnifyingGlass
		class="absolute top-1/2 right-1 mr-2 -translate-y-1/2 self-center text-neutral-700 dark:text-neutral-400"
	/>
</div>
