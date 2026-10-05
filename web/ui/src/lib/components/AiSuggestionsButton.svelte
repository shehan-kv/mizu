<script lang="ts">
	import { onMount } from 'svelte';
	import { aiState } from '$lib/ai/state.svelte';
	import { initialize, enable, disable } from '$lib/ai';

	interface Props {
		class?: string;
	}

	let { class: className = '' }: Props = $props();

	onMount(() => {
		void initialize();
	});

	async function toggleAI() {
		if (aiState.enabled) {
			disable();
			return;
		}

		await enable();
	}
</script>

<button
	type="button"
	class={className}
	onclick={toggleAI}
	disabled={aiState.status === 'checking' ||
		aiState.status === 'loading' ||
		aiState.status === 'downloading' ||
		aiState.status === 'unavailable'}
>
	<span
		class="mr-1.5 inline-block size-2 rounded-full border transition"
		class:bg-emerald-300={aiState.enabled}
		class:dark:bg-emerald-500={aiState.enabled}
		class:bg-neutral-50={!aiState.enabled}
		class:dark:bg-neutral-950={!aiState.enabled}
	></span>

	{#if aiState.status === 'checking'}
		Checking AI...
	{:else if aiState.status === 'available'}
		Enable AI Suggestions
	{:else if aiState.status === 'downloading'}
		Downloading AI... {Math.round(aiState.progress * 100)}%
	{:else if aiState.status === 'loading'}
		Loading AI...
	{:else if aiState.status === 'unavailable'}
		AI Unavailable
	{:else if aiState.status === 'error'}
		Retry AI
	{:else if aiState.enabled}
		AI Suggestions On
	{:else}
		AI Suggestions Off
	{/if}
</button>
