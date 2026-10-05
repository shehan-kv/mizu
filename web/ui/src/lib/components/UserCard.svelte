<script lang="ts">
	import { BASE_URL } from '$lib/api/client';
	import { toTitleCaseDashed } from '$lib/utils/toTitleCaseDashed';

	interface Props {
		id: string;
		hasImage: boolean;
		name: string;
		title?: string;
		role?: string;
	}
	let { id, hasImage, name, title, role }: Props = $props();
</script>

<div class="flex gap-2">
	<div class="size-10 shrink-0 overflow-hidden rounded-full bg-neutral-300 dark:bg-neutral-800">
		{#if hasImage}
			<img
				src={`${BASE_URL}users/profile-images/${id}`}
				alt={`${name} profile picture`}
				class="size-full object-cover"
			/>
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
