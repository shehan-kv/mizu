<script lang="ts">
	import { toTitleCaseDashed } from '$lib/utils/toTitleCaseDashed';

	interface Props {
		image?: string;
		name: string;
		title?: string;
		role?: string;
	}
	let { image, name, title, role }: Props = $props();
</script>

<div class="flex gap-2">
	<div class="size-10 overflow-hidden rounded-full bg-neutral-300 dark:bg-neutral-800">
		{#if image}
			<img src={image} alt={`${name} profile picture`} class="size-full object-cover" />
		{:else if name}
			<div class="flex size-full items-center justify-center text-neutral-500">
				{name[0]}
			</div>
		{/if}
	</div>

	<div>
		<p class="text-sm">
			{name || 'N/A'}
		</p>
		<p class="text-xs text-neutral-500 dark:text-neutral-400">
			{#if title || role}
				{title ? title : ''}
				{role ? `${title ? ' - ' : ''}${toTitleCaseDashed(role)}` : ''}
			{:else}
				N/A
			{/if}
		</p>
	</div>
</div>
