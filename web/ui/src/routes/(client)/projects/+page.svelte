<script>
	import * as Table from '$lib/components/ui/table';
	import SearchBar from '$lib/components/SearchBar.svelte';
	import ArrowRight from 'phosphor-svelte/lib/ArrowRight';
	import Pagination from '$lib/components/Pagination.svelte';

	const projects = [
		{
			projectName: 'Website Redesign',
			status: 'In Progress',
			tasksCompleted: '12 / 36',
			createdDate: '2024-11-10',
			invoicesIssued: 2,
			quotesIssued: 1
		},
		{
			projectName: 'Mobile App Development',
			status: 'Not Started',
			tasksCompleted: '0 / 50',
			createdDate: '2025-01-15',
			invoicesIssued: 0,
			quotesIssued: 2
		},
		{
			projectName: 'Marketing Automation',
			status: 'Completed',
			tasksCompleted: '20 / 20',
			createdDate: '2024-09-25',
			invoicesIssued: 3,
			quotesIssued: 1
		},
		{
			projectName: 'eCommerce Platform',
			status: 'Paused',
			tasksCompleted: '18 / 45',
			createdDate: '2024-10-02',
			invoicesIssued: 1,
			quotesIssued: 2
		},
		{
			projectName: 'Cloud Migration',
			status: 'In Progress',
			tasksCompleted: '9 / 30',
			createdDate: '2025-03-08',
			invoicesIssued: 2,
			quotesIssued: 1
		},
		{
			projectName: 'Security Audit',
			status: 'Completed',
			tasksCompleted: '10 / 10',
			createdDate: '2024-07-19',
			invoicesIssued: 2,
			quotesIssued: 0
		},
		{
			projectName: 'Data Dashboard',
			status: 'Cancelled',
			tasksCompleted: '5 / 40',
			createdDate: '2024-12-01',
			invoicesIssued: 1,
			quotesIssued: 1
		},
		{
			projectName: 'Social Media Analytics',
			status: 'In Progress',
			tasksCompleted: '22 / 60',
			createdDate: '2025-02-11',
			invoicesIssued: 2,
			quotesIssued: 2
		},
		{
			projectName: 'Landing Page A/B Test',
			status: 'Completed',
			tasksCompleted: '8 / 8',
			createdDate: '2024-08-28',
			invoicesIssued: 1,
			quotesIssued: 0
		},
		{
			projectName: 'CRM Integration',
			status: 'In Progress',
			tasksCompleted: '15 / 30',
			createdDate: '2024-11-21',
			invoicesIssued: 3,
			quotesIssued: 2
		},
		{
			projectName: 'Internal Portal',
			status: 'Paused',
			tasksCompleted: '7 / 25',
			createdDate: '2025-01-06',
			invoicesIssued: 1,
			quotesIssued: 1
		},
		{
			projectName: 'Event Booking System',
			status: 'Not Started',
			tasksCompleted: '0 / 42',
			createdDate: '2025-04-02',
			invoicesIssued: 0,
			quotesIssued: 1
		},
		{
			projectName: 'AI Chatbot',
			status: 'In Progress',
			tasksCompleted: '19 / 40',
			createdDate: '2025-02-23',
			invoicesIssued: 2,
			quotesIssued: 3
		},
		{
			projectName: 'API Gateway Setup',
			status: 'Completed',
			tasksCompleted: '12 / 12',
			createdDate: '2024-10-13',
			invoicesIssued: 1,
			quotesIssued: 0
		},
		{
			projectName: 'Email Campaign Builder',
			status: 'Cancelled',
			tasksCompleted: '4 / 20',
			createdDate: '2024-09-30',
			invoicesIssued: 0,
			quotesIssued: 2
		}
	];

	let page = $state(1);
</script>

<svelte:head>
	<title>Projects</title>
</svelte:head>

<div class="grid h-full auto-rows-[min-content_1fr_min-content] gap-6">
	<div class="mx-auto lg:container">
		<div class="max-w-96">
			<SearchBar />
		</div>
	</div>

	<div class="mx-auto gap-4 overflow-y-auto lg:container">
		<div class="overflow-y-auto">
			{#if projects.length > 0}
				<Table.Root class="container mx-auto">
					<Table.Header>
						<Table.Row>
							<Table.Head class="font-bold">Project Name</Table.Head>
							<Table.Head class="font-bold">Status</Table.Head>
							<Table.Head class="font-bold">Tasks</Table.Head>
							<Table.Head class="font-bold">Created Date</Table.Head>
							<Table.Head class="font-bold">Invoices</Table.Head>
							<Table.Head class="font-bold">Quotations</Table.Head>
							<Table.Head class="font-bold">Actions</Table.Head>
						</Table.Row>
					</Table.Header>
					<Table.Body>
						{#each projects as project}
							<Table.Row>
								<Table.Cell>{project.projectName}</Table.Cell>
								<Table.Cell class="flex items-center gap-1.5">
									{#if project.status == 'In Progress'}
										<span class="relative flex size-2">
											<span
												class="absolute inline-flex h-full w-full animate-ping rounded-full
												bg-green-500 opacity-75 dark:bg-green-600"
											>
											</span>
											<span
												class="relative inline-flex size-2 rounded-full bg-green-500 dark:bg-green-600"
											></span>
										</span>
									{/if}
									{project.status}
								</Table.Cell>
								<Table.Cell>{project.tasksCompleted} Completed</Table.Cell>
								<Table.Cell>{project.createdDate}</Table.Cell>
								<Table.Cell>{project.invoicesIssued} / 10 Paid</Table.Cell>
								<Table.Cell>{project.quotesIssued}</Table.Cell>
								<Table.Cell>
									<div
										class="text-xs text-neutral-500 *:cursor-pointer *:px-1.5 *:hover:text-neutral-950 dark:text-neutral-400 *:dark:hover:text-neutral-50"
									>
										<button title="View">
											<ArrowRight size={18} />
										</button>
									</div>
								</Table.Cell>
							</Table.Row>
						{/each}
					</Table.Body>
				</Table.Root>
			{/if}
		</div>
	</div>
	{#if projects.length > 0}
		<div class="container mx-auto flex justify-end">
			<Pagination bind:page />
		</div>
	{/if}
</div>
