<script lang="ts">
	import { debounce } from '$lib/utils/debounce';
	import MagnifyingGlass from 'phosphor-svelte/lib/MagnifyingGlass';

	interface Props {
		value: string;
		delay?: number;
		onchange?: () => any;
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

<div class="relative border-b border-neutral-200 dark:border-neutral-800">
	<input
		type="search"
		name="search"
		id="search"
		bind:value={inputValue}
		oninput={onInput}
		class="w-full py-1.5 pl-1 pr-10 text-sm outline-none"
	/>
	<MagnifyingGlass
		class="absolute right-1 top-1/2 mr-2 -translate-y-1/2 self-center text-neutral-700 dark:text-neutral-400"
	/>
</div>
