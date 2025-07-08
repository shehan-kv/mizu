<script lang="ts">
	import ChatDots from 'phosphor-svelte/lib/ChatDots';
	import Chats from 'phosphor-svelte/lib/Chats';
	import User from 'phosphor-svelte/lib/User';
	import WarningCircle from 'phosphor-svelte/lib/WarningCircle';
	import Info from 'phosphor-svelte/lib/Info';

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
		<User size={18} />
	{:else if variant == 'channel'}
		<ChatDots size={18} />
	{:else if variant == 'message'}
		<Chats size={18} />
	{:else if variant == 'info'}
		<Info size={18} />
	{:else if variant == 'warn'}
		<WarningCircle size={18} />
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
