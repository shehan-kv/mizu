<script lang="ts">
	import Checks from 'phosphor-svelte/lib/Checks';
	import DownloadSimple from 'phosphor-svelte/lib/DownloadSimple';
	import ArrowRight from 'phosphor-svelte/lib/ArrowRight';
	import Envelope from 'phosphor-svelte/lib/Envelope';
	import * as Table from '$lib/components/ui/table';
	import SearchBar from '$lib/components/SearchBar.svelte';
	import Pagination from '$lib/components/Pagination.svelte';

	let invoicesAndQuotes = [
		{
			type: 'Invoice',
			id: 'INV-1001',
			amount: 1250.0,
			status: 'Paid',
			issuedOn: '2025-06-15',
			dueDate: '2025-07-15',
			projectName: 'Website Redesign'
		},
		{
			type: 'Quote',
			id: 'QTE-2001',
			amount: 3500.0,
			status: 'Pending',
			issuedOn: '2025-06-20',
			dueDate: null,
			projectName: 'Mobile App Development'
		},
		{
			type: 'Invoice',
			id: 'INV-1002',
			amount: 875.5,
			status: 'Overdue',
			issuedOn: '2025-05-10',
			dueDate: '2025-06-10',
			projectName: 'Cloud Migration'
		},
		{
			type: 'Quote',
			id: 'QTE-2002',
			amount: 2150.0,
			status: 'Accepted',
			issuedOn: '2025-06-25',
			dueDate: null,
			projectName: 'Digital Marketing Campaign'
		},
		{
			type: 'Invoice',
			id: 'INV-1003',
			amount: 4000.0,
			status: 'Paid',
			issuedOn: '2025-04-30',
			dueDate: '2025-05-30',
			projectName: 'E-commerce Platform Setup'
		},
		{
			type: 'Invoice',
			id: 'INV-1004',
			amount: 620.0,
			status: 'Pending',
			issuedOn: '2025-06-28',
			dueDate: '2025-07-28',
			projectName: 'SEO Optimization'
		},
		{
			type: 'Quote',
			id: 'QTE-2003',
			amount: 1100.0,
			status: 'Rejected',
			issuedOn: '2025-06-18',
			dueDate: null,
			projectName: 'Content Creation'
		},
		{
			type: 'Invoice',
			id: 'INV-1005',
			amount: 3750.0,
			status: 'Paid',
			issuedOn: '2025-06-01',
			dueDate: '2025-07-01',
			projectName: 'CRM Integration'
		},
		{
			type: 'Quote',
			id: 'QTE-2004',
			amount: 900.0,
			status: 'Pending',
			issuedOn: '2025-07-01',
			dueDate: null,
			projectName: 'Social Media Strategy'
		},
		{
			type: 'Invoice',
			id: 'INV-1006',
			amount: 1450.0,
			status: 'Overdue',
			issuedOn: '2025-05-20',
			dueDate: '2025-06-20',
			projectName: 'Network Security Upgrade'
		},
		{
			type: 'Invoice',
			id: 'INV-1007',
			amount: 2800.0,
			status: 'Paid',
			issuedOn: '2025-06-10',
			dueDate: '2025-07-10',
			projectName: 'Mobile Payment Integration'
		},
		{
			type: 'Quote',
			id: 'QTE-2005',
			amount: 4200.0,
			status: 'Pending',
			issuedOn: '2025-07-05',
			dueDate: null,
			projectName: 'AI Chatbot Development'
		},
		{
			type: 'Invoice',
			id: 'INV-1008',
			amount: 1600.0,
			status: 'Overdue',
			issuedOn: '2025-05-25',
			dueDate: '2025-06-25',
			projectName: 'Data Analytics Setup'
		},
		{
			type: 'Quote',
			id: 'QTE-2006',
			amount: 1950.0,
			status: 'Accepted',
			issuedOn: '2025-06-30',
			dueDate: null,
			projectName: 'Email Marketing Campaign'
		},
		{
			type: 'Invoice',
			id: 'INV-1009',
			amount: 3500.0,
			status: 'Paid',
			issuedOn: '2025-06-05',
			dueDate: '2025-07-05',
			projectName: 'Website Security Audit'
		},
		{
			type: 'Invoice',
			id: 'INV-1010',
			amount: 800.0,
			status: 'Pending',
			issuedOn: '2025-07-02',
			dueDate: '2025-08-02',
			projectName: 'UX/UI Improvements'
		},
		{
			type: 'Quote',
			id: 'QTE-2007',
			amount: 1250.0,
			status: 'Rejected',
			issuedOn: '2025-06-22',
			dueDate: null,
			projectName: 'Video Production'
		},
		{
			type: 'Invoice',
			id: 'INV-1011',
			amount: 2900.0,
			status: 'Paid',
			issuedOn: '2025-06-15',
			dueDate: '2025-07-15',
			projectName: 'Cloud Backup Solution'
		},
		{
			type: 'Quote',
			id: 'QTE-2008',
			amount: 2100.0,
			status: 'Pending',
			issuedOn: '2025-07-03',
			dueDate: null,
			projectName: 'Mobile Game Design'
		},
		{
			type: 'Invoice',
			id: 'INV-1012',
			amount: 1750.0,
			status: 'Overdue',
			issuedOn: '2025-05-30',
			dueDate: '2025-06-30',
			projectName: 'API Development'
		}
	];
	let page = $state(1);

	const usdFormatter = new Intl.NumberFormat('en-US', {
		style: 'currency',
		currency: 'USD'
	});
</script>

<svelte:head>
	<title>Invoices and Quotes</title>
</svelte:head>

<div class="grid h-full auto-rows-[min-content_1fr_min-content] gap-6">
	<div class="mx-auto lg:container">
		<div class="max-w-96">
			<SearchBar />
		</div>
	</div>
	<div class="mx-auto gap-4 overflow-y-auto lg:container">
		<div class="overflow-y-auto">
			{#if invoicesAndQuotes.length > 0}
				<Table.Root class="container mx-auto">
					<Table.Header>
						<Table.Row>
							<Table.Head class="font-bold">Type</Table.Head>
							<Table.Head class="font-bold">#ID</Table.Head>
							<Table.Head class="font-bold">Project</Table.Head>
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
								<Table.Cell>{entry.projectName}</Table.Cell>
								<Table.Cell>{usdFormatter.format(entry.amount)}</Table.Cell>
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
	</div>
	{#if invoicesAndQuotes.length > 0}
		<div class="container mx-auto flex justify-end">
			<Pagination bind:page />
		</div>
	{/if}
</div>
