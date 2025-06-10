<script lang="ts">
	import ChatsIcon from './icons/ChatsIcon.svelte';
	import CheckCircleIcon from './icons/CheckCircleIcon.svelte';
	import InvoiceIcon from './icons/InvoiceIcon.svelte';
	import KanbanIcon from './icons/KanbanIcon.svelte';
	import ScrollIcon from './icons/ScrollIcon.svelte';
	import SpinnerIcon from './icons/SpinnerIcon.svelte';
	import TicketIcon from './icons/TicketIcon.svelte';

	interface Props {
		id: number;
		status: 'started' | 'paused' | 'stopped';
		lastUpdated: string;
		projectName: string;
		tasks: {
			total: number;
			completed: number;
			inProgress: number;
		};
	}

	let { data }: { data: Props } = $props();

	let completedPercentage = Math.floor((data.tasks.completed / data.tasks.total) * 100);
</script>

{#snippet startedState()}
	<div class="flex items-center gap-2">
		<span class="relative flex size-3">
			<span
				class="absolute inline-flex h-full w-full animate-ping rounded-full bg-green-500 opacity-75 dark:bg-green-600"
			>
			</span>
			<span class="relative inline-flex size-3 rounded-full bg-green-500 dark:bg-green-600"></span>
		</span>
		<p class="text-xs text-neutral-700 dark:text-neutral-300">Started</p>
	</div>
{/snippet}

{#snippet pausedState()}
	<div class="flex items-center gap-2">
		<span class="size-3 rounded-full bg-yellow-400 dark:bg-yellow-600"></span>
		<p class="text-xs text-neutral-700 dark:text-neutral-300">Paused</p>
	</div>
{/snippet}

{#snippet stoppedState()}
	<div class="flex items-center gap-2">
		<span class="size-3 rounded-full bg-red-400 dark:bg-red-600"></span>
		<p class="text-xs text-neutral-700 dark:text-neutral-300">Stopped</p>
	</div>
{/snippet}

{#snippet naState()}
	<div class="flex items-center gap-2">
		<span class="size-3 rounded-full bg-neutral-400 dark:bg-neutral-600"></span>
		<p class="text-xs text-neutral-700 dark:text-neutral-300">N/A</p>
	</div>
{/snippet}

<div class="rounded border p-4">
	<div class="flex items-center justify-between">
		{#if data.status == 'started'}
			{@render startedState()}
		{:else if data.status == 'paused'}
			{@render pausedState()}
		{:else if data.status == 'stopped'}
			{@render stoppedState()}
		{:else}
			{@render naState()}
		{/if}
		<p class="text-xs text-neutral-600 dark:text-neutral-400">Last Update: {data.lastUpdated}</p>
	</div>

	<p class="mt-2 font-bold">{data.projectName}</p>
	<div class="mt-4 space-y-1.5">
		<div class="flex items-center gap-2">
			<CheckCircleIcon class="size-5" />
			<p class="text-sm">
				{data.tasks.completed} / {data.tasks.total} Tasks Completed ({completedPercentage}%)
			</p>
		</div>
		<div class="flex items-center gap-2">
			<SpinnerIcon class="size-5" />
			<p class="text-sm">{data.tasks.inProgress} Task In Progress</p>
		</div>
	</div>

	<div class="mt-4 grid grid-cols-[1fr_min-content] gap-2">
		<div
			class="h-2 w-full self-end overflow-hidden rounded-full bg-neutral-200 dark:bg-neutral-900"
		>
			<div
				class="h-full rounded-full bg-neutral-950 dark:bg-neutral-200"
				style={`width: ${completedPercentage}%;`}
			></div>
		</div>

		<div class="flex gap-0.5">
			<button class="rounded border p-1.5 hover:bg-neutral-100 dark:hover:bg-neutral-900">
				<KanbanIcon class="size-5" />
			</button>
			<button class="rounded border p-1.5 hover:bg-neutral-100 dark:hover:bg-neutral-900">
				<ScrollIcon class="size-5" />
			</button>
			<button class="rounded border p-1.5 hover:bg-neutral-100 dark:hover:bg-neutral-900">
				<TicketIcon class="size-5" />
			</button>
			<button class="rounded border p-1.5 hover:bg-neutral-100 dark:hover:bg-neutral-900">
				<ChatsIcon class="size-5" />
			</button>
			<button class="rounded border p-1.5 hover:bg-neutral-100 dark:hover:bg-neutral-900">
				<InvoiceIcon class="size-5" />
			</button>
		</div>
	</div>
</div>
