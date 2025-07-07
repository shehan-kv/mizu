<script lang="ts">
	import ChatDotsIcon from './icons/ChatDotsIcon.svelte';
	import ChatsIcon from './icons/ChatsIcon.svelte';
	import UserIcon from './icons/UserIcon.svelte';
	import WarningCircleIcon from './icons/WarningCircleIcon.svelte';

	interface Props {
		variant: 'user' | 'channel' | 'message' | 'info' | 'warn';
		text: string;
		retry?: () => void;
	}

	let { variant, text, retry }: Props = $props();
</script>

<div
	class="max-h-sm flex h-full flex-col items-center
justify-center gap-2 py-4 text-xs text-neutral-500"
>
	{#if variant == 'user'}
		<UserIcon class="size-5" />
	{:else if variant == 'channel'}
		<ChatDotsIcon class="size-5" />
	{:else if variant == 'message'}
		<ChatsIcon class="size-5" />
	{:else if variant == 'info'}
		<ChatsIcon class="size-5" />
	{:else if variant == 'warn'}
		<WarningCircleIcon class="size-5" />
	{/if}
	<p>{text}</p>

	{#if retry}
		<button
			class="mt-2 cursor-pointer rounded bg-neutral-200 px-3 py-1.5
            text-neutral-950 hover:bg-neutral-300 dark:bg-neutral-800 dark:text-neutral-50
			dark:hover:bg-neutral-700 dark:hover:text-neutral-200"
			onclick={retry}
		>
			Retry
		</button>
	{/if}
</div>
