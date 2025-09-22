<script lang="ts">
	import { debounce } from '$lib/utils/debounce';
	import MagnifyingGlass from 'phosphor-svelte/lib/MagnifyingGlass';

	let { search } = $props();

	let inputValue = $state('');

	const updateKeywordDebounced = debounce((val: string) => {
		search(val);
	}, 300);

	function onInput() {
		updateKeywordDebounced(inputValue);
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
