<script>
	import CircleNotch from 'phosphor-svelte/lib/CircleNotch';
	import MagnifyingGlass from 'phosphor-svelte/lib/MagnifyingGlass';
	let isSearching = $state(false);
</script>

<div class="relative border-b border-neutral-200 dark:border-neutral-800">
	<input
		type="search"
		name="search"
		id="search"
		class="w-full py-1.5 pl-1 pr-10 text-sm outline-none"
	/>
	{#if isSearching}
		<CircleNotch
			class="absolute right-2 mr-2 animate-spin self-center text-neutral-700 dark:text-neutral-400"
		/>
	{:else}
		<MagnifyingGlass
			class="absolute right-1 top-1/2 mr-2 -translate-y-1/2 self-center text-neutral-700 dark:text-neutral-400"
		/>
	{/if}
</div>
