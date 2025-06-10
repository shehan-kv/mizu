<script>
	import CheckCircleIcon from '$lib/components/icons/checkCircleIcon.svelte';
	import SpinnerGapIcon from '$lib/components/icons/spinnerGapIcon.svelte';
	import TicketIcon from '$lib/components/icons/ticketIcon.svelte';
	import WarningCircleIcon from '$lib/components/icons/warningCircleIcon.svelte';
	import XCircleIcon from '$lib/components/icons/xCircleIcon.svelte';
	import ProjectCard from '$lib/components/projectCard.svelte';
	import Check from '@lucide/svelte/icons/check';
</script>

<svelte:head>
	<title>Dashboard</title>
</svelte:head>

{#snippet statCard(title, total, metrics)}
	<div class="@container">
		<div class="@sm:grid-cols-3 grid grid-cols-1 rounded border">
			<div class="@sm:py-8 space-y-2 px-4 py-6">
				<p class="text-sm">{title}</p>
				<p class="text-5xl">{total}</p>
			</div>
			<div class="col-span-2 grid grid-cols-3 items-center bg-neutral-100 p-4 dark:bg-neutral-900">
				{#each metrics as metric}
					<div class="space-y-1">
						<div class="@sm:items-center @sm:flex-row flex flex-col gap-1">
							{#if metric.icon}
								<metric.icon class="@sm:size-4.5 size-4" />
							{/if}
							<p class="@sm:text-sm text-xs">{metric.label}</p>
						</div>
						<p class="text-3xl">{metric.count}</p>
					</div>
				{/each}
			</div>
		</div>
	</div>
{/snippet}

<div class="mx-auto grid grid-cols-5 gap-4 lg:container">
	<div class="col-span-2 grid auto-rows-[1fr_min-content] gap-6 rounded border p-6">
		<div class="">Chart goes here</div>
		<div class="space-y-6">
			<div class="space-y-1">
				<p class="text-sm">Total Paid Amount</p>
				<p class="text-4xl">$ 143,000.00</p>
			</div>
			<div class="grid grid-cols-2">
				<div class="space-y-1">
					<p class="text-sm">Outstanding Balance</p>
					<p class="text-2xl">$ 99,000.00</p>
				</div>
				<div class="space-y-1">
					<p class="text-sm">Overdue Amount</p>
					<p class="text-2xl">$ 00.00</p>
				</div>
			</div>
		</div>
	</div>
	<div class="space-y-4">
		{@render statCard('All Projects', 12, [
			{ label: 'Completed', count: 8, icon: CheckCircleIcon },
			{ label: 'Paused', count: 0, icon: WarningCircleIcon },
			{ label: 'Cancelled', count: 4, icon: XCircleIcon }
		])}
		{@render statCard('Quotes Issued', 17, [
			{ label: 'Accepted', count: 8, icon: CheckCircleIcon },
			{ label: 'Rejected', count: 0, icon: WarningCircleIcon },
			{ label: 'Pending', count: 4, icon: XCircleIcon }
		])}
	</div>
	<div class="col-span-2 grid auto-rows-[min-content_1fr_min-content] space-y-4">
		{@render statCard('Invoices Issued', 17, [
			{ label: 'Paid', count: 8, icon: CheckCircleIcon },
			{ label: 'Pending', count: 0, icon: WarningCircleIcon },
			{ label: 'Overdue', count: 4, icon: XCircleIcon }
		])}

		<div class="space-y-6 rounded border px-4 py-6">
			<p class="text-sm">Support Tickets</p>

			<div class="grid grid-cols-3">
				<div class="space-y-2">
					<div class="flex items-center gap-2">
						<TicketIcon class="size-5" />
						<p>Tickets</p>
					</div>
					<p class="text-4xl">0</p>
				</div>
				<div class="space-y-2">
					<div class="flex items-center gap-2">
						<SpinnerGapIcon class="size-5" />
						<p>Open</p>
					</div>
					<p class="text-4xl">0</p>
				</div>
				<div class="space-y-2">
					<div class="flex items-center gap-2">
						<Check class="size-5" />
						<p>Closed</p>
					</div>
					<p class="text-4xl">0</p>
				</div>
			</div>
		</div>

		{@render statCard('Contracts', 17, [
			{ label: 'Signed', count: 8, icon: CheckCircleIcon },
			{ label: 'Rejected', count: 0, icon: WarningCircleIcon },
			{ label: 'Pending', count: 4, icon: XCircleIcon }
		])}
	</div>
</div>

<div class="container mx-auto mt-10 border-t py-4">
	<p class="text-neutral-500">Recent Projects</p>

	<div class="mt-4 grid grid-cols-4 gap-4">
		<ProjectCard
			data={{
				id: 1,
				lastUpdated: new Date().toLocaleString(),
				projectName: 'Atlas CRM Integration',
				status: 'started',
				tasks: { completed: 25, inProgress: 1, total: 50 }
			}}
		/>
	</div>
</div>
