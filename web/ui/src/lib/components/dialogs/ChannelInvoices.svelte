<script lang="ts">
	import DownloadSimple from 'phosphor-svelte/lib/DownloadSimple';
	import Envelope from 'phosphor-svelte/lib/Envelope';
	import ArrowRight from 'phosphor-svelte/lib/ArrowRight';
	import Checks from 'phosphor-svelte/lib/Checks';

	import * as Table from '$lib/components/ui/table';
	import type { Channel } from '../message/types';
	import Pagination from '../Pagination.svelte';
	import SearchBar from '../SearchBar.svelte';
	import FullScreenDialog from './FullScreenDialog.svelte';
	import { currencyFormatter } from '$lib/utils/currencyFormatter';

	interface Props {
		open: Boolean;
		close: () => void;
		channel?: Channel | null;
	}
	let { open = $bindable(), close, channel }: Props = $props();

	let page = $state(1);

	let invoicesAndQuotes = [
		{
			type: 'Invoice',
			id: 'INV-1001',
			amount: 1250.0,
			status: 'Paid',
			issuedOn: '2025-06-15',
			dueDate: '2025-07-15'
		},
		{
			type: 'Quote',
			id: 'QTE-2001',
			amount: 3500.0,
			status: 'Pending',
			issuedOn: '2025-06-20',
			dueDate: null
		},
		{
			type: 'Invoice',
			id: 'INV-1002',
			amount: 875.5,
			status: 'Overdue',
			issuedOn: '2025-05-10',
			dueDate: '2025-06-10'
		},
		{
			type: 'Quote',
			id: 'QTE-2002',
			amount: 2150.0,
			status: 'Accepted',
			issuedOn: '2025-06-25',
			dueDate: null
		},
		{
			type: 'Invoice',
			id: 'INV-1003',
			amount: 4000.0,
			status: 'Paid',
			issuedOn: '2025-04-30',
			dueDate: '2025-05-30'
		},
		{
			type: 'Invoice',
			id: 'INV-1004',
			amount: 620.0,
			status: 'Pending',
			issuedOn: '2025-06-28',
			dueDate: '2025-07-28'
		},
		{
			type: 'Quote',
			id: 'QTE-2003',
			amount: 1100.0,
			status: 'Rejected',
			issuedOn: '2025-06-18',
			dueDate: null
		},
		{
			type: 'Invoice',
			id: 'INV-1005',
			amount: 3750.0,
			status: 'Paid',
			issuedOn: '2025-06-01',
			dueDate: '2025-07-01'
		},
		{
			type: 'Quote',
			id: 'QTE-2004',
			amount: 900.0,
			status: 'Pending',
			issuedOn: '2025-07-01',
			dueDate: null
		},
		{
			type: 'Invoice',
			id: 'INV-1006',
			amount: 1450.0,
			status: 'Overdue',
			issuedOn: '2025-05-20',
			dueDate: '2025-06-20'
		}
	];
</script>

<FullScreenDialog bind:open {close}>
	<div class="grid auto-rows-[min-content_1fr_min-content] gap-6 overflow-y-auto px-5">
		<div class="flex-none">
			<div class="container mx-auto flex items-end justify-between gap-4">
				<p class="font-bold">Invoices & Quotes of {channel?.name}</p>
				<div class="w-full max-w-xs">
					<SearchBar />
				</div>
			</div>
		</div>

		<div class="overflow-y-auto">
			{#if invoicesAndQuotes.length > 0}
				<Table.Root class="container mx-auto">
					<Table.Header>
						<Table.Row>
							<Table.Head class="font-bold">Type</Table.Head>
							<Table.Head class="font-bold">#ID</Table.Head>
							<Table.Head class="font-bold">Amount</Table.Head>
							<Table.Head class="font-bold">Status</Table.Head>
							<Table.Head class="font-bold">Issued On</Table.Head>
							<Table.Head class="font-bold">Due Date</Table.Head>
							<Table.Head class="font-bold">Actions</Table.Head>
						</Table.Row>
					</Table.Header>
					<Table.Body>
						{#each invoicesAndQuotes as entry}
							<Table.Row>
								<Table.Cell>{entry.type}</Table.Cell>
								<Table.Cell>#{entry.id}</Table.Cell>
								<Table.Cell>{currencyFormatter('USD', entry.amount)}</Table.Cell>
								<Table.Cell class="flex items-center gap-1">
									{entry.status}
									{#if entry.status == 'Paid' || entry.status == 'Accepted'}
										<Checks size={18} class="text-emerald-500" />
									{/if}
								</Table.Cell>
								<Table.Cell>{entry.issuedOn}</Table.Cell>
								<Table.Cell>{entry.dueDate}</Table.Cell>
								<Table.Cell>
									<div
										class="text-xs text-neutral-500 *:cursor-pointer *:px-1.5 *:hover:text-neutral-950 dark:text-neutral-400 *:dark:hover:text-neutral-50"
									>
										<button title="View">
											<ArrowRight size={18} />
										</button>
										<button title="Download as PDF">
											<DownloadSimple size={18} />
										</button>
										<button title="Email Me"><Envelope size={18} /></button>
									</div>
								</Table.Cell>
							</Table.Row>
						{/each}
					</Table.Body>
				</Table.Root>
			{/if}
		</div>
		{#if invoicesAndQuotes.length > 0}
			<div class="container mx-auto flex justify-end">
				<Pagination bind:page />
			</div>
		{/if}
	</div>
</FullScreenDialog>
