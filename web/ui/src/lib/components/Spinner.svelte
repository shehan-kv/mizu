<script lang="ts">
	import CircleNotch from 'phosphor-svelte/lib/CircleNotch';

	let { size = 24, class: className = '' } = $props();
</script>

<div class={`flex h-full items-center justify-center text-neutral-500 ${className}`}>
	<CircleNotch class="animate-spin" {size} />
</div>
