<script>
	import CircleNotchIcon from './icons/CircleNotchIcon.svelte';
	import MagnifyingGlassIcon from './icons/MagnifyingGlassIcon.svelte';
	let isSearching = $state(false);
</script>

<div
	class="grid grid-cols-[1fr_min-content] rounded border border-neutral-200 bg-neutral-100 dark:border-neutral-800 dark:bg-neutral-900"
>
	<input type="search" name="search" id="search" class="px-2 py-1.5 text-sm outline-none" />
	{#if isSearching}
		<CircleNotchIcon
			class="mr-2 size-5 animate-spin self-center text-neutral-700 dark:text-neutral-400"
		/>
	{:else}
		<MagnifyingGlassIcon class="mr-2 size-5 self-center text-neutral-700 dark:text-neutral-400" />
	{/if}
</div>
