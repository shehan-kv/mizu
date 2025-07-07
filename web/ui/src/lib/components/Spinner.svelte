<script lang="ts">
	import CircleNotch from 'phosphor-svelte/lib/CircleNotch';

	let { size = 24, class: className = '' } = $props();
</script>

<div
	class={`max-h-sm flex h-full flex-col items-center
justify-center gap-2 py-4 text-xs text-neutral-500 ${className}`}
>
	<CircleNotch class="animate-spin" {size} />
</div>
