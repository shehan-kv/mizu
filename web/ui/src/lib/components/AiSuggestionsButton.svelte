<script lang="ts">
	let { isAiEnabled, onclick, class: className = '' } = $props();
</script>

<button {onclick} class={className}>
	<span
		class="mr-1.5 inline-block size-2 rounded-full border transition"
		class:bg-emerald-300={isAiEnabled}
		class:dark:bg-emerald-500={isAiEnabled}
		class:bg-neutral-50={!isAiEnabled}
		class:dark:bg-neutral-950={!isAiEnabled}
	></span>AI Suggestions
</button>
