<script lang="ts">
	import type { Snippet } from 'svelte';

	type Props = { title: string; class?: string; children: Snippet };

	let { title, class: className = '', children }: Props = $props();
</script>

<div
	class={`grid grid-rows-[min-content_1fr] overflow-hidden rounded bg-neutral-50 dark:bg-neutral-950 ${className}`}
>
	<div class="border-b px-4 py-2">
		<p class="text-sm text-neutral-600 dark:text-neutral-300">{title}</p>
	</div>

	{@render children()}
</div>
