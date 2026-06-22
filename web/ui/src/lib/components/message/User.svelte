<script lang="ts">
	import { BASE_URL } from '$lib/api/client';
	import { formatDate } from '$lib/utils/formatDate';

	interface Props {
		id: string;
		hasImage: boolean;
		name: string;
		title?: string;
		message: string;
		date: Date;
	}
	let { id, hasImage, name, title, message, date }: Props = $props();
</script>

<div class="flex gap-2 py-3.5">
	<div class="size-10 overflow-hidden rounded-full bg-neutral-300 dark:bg-neutral-800">
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
	<div class="space-y-1">
		<div>
			<p class="flex gap-2 text-sm text-neutral-700 dark:text-neutral-400">
				{name}
				{#if title}
					<span>|</span>
					{title}
				{/if}
			</p>
			<p class="text-xs text-neutral-500">{formatDate(date)}</p>
		</div>

		<p class="max-w-2xl">
			{message}
		</p>
	</div>
</div>
